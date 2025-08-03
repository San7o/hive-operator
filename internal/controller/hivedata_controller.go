/*
                    GNU GENERAL PUBLIC LICENSE
                       Version 2, June 1991

 Copyright (C) 1989, 1991 Free Software Foundation, Inc.,
 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.
*/

// SPDX-License-Identifier: GPL-2.0-only

package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
	hivebpf "github.com/San7o/hive-operator/internal/controller/ebpf"
)

type HiveDataReconciler struct {
	client.Client
	UncachedClient client.Reader
	Scheme         *runtime.Scheme
}

// +kubebuilder:rbac:groups=hive-operator.com,resources=hivedata,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hive-operator.com,resources=hivedata/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hive-operator.com,resources=hivedata/finalizers,verbs=update

func (r *HiveDataReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := logger.FromContext(ctx)
	log.Info("HiveData reconcile triggered.")

	if !hivebpf.Loaded {
		log.Info("Loading eBPF program")
		if err := hivebpf.LoadEbpf(ctx); err != nil { // Fatal
			return ctrl.Result{}, fmt.Errorf("Reconcile Error Load eBPF program: %w", err)
		}
		go Output(r.UncachedClient)
	}

	hiveDataList := &hivev1alpha1.HiveDataList{}
	err := r.Client.List(ctx, hiveDataList)
	if err != nil { // Fatal
		return ctrl.Result{}, fmt.Errorf("Reconcile Error Failed to get Hive Data resource: %w", err)
	}

	hivePolicyList := &hivev1alpha1.HivePolicyList{}
	err = r.Client.List(ctx, hivePolicyList)
	if err != nil { // Fatal
		return ctrl.Result{}, fmt.Errorf("Reconcile Error Failed to get HivePolicy resource: %w", err)
	}

	var it uint32 = 0

	// Check if each HiveData (referring to this kernel id) does have a
	// corresponding HivePolicy. If it does, then we update the eBPF
	// map with the information from the HiveData. If it doesn't, then
	// the HivePolicy has been eliminated and the HiveData should be
	// deleted.
Data:
	for _, hiveData := range hiveDataList.Items {

		if hiveData.Spec.KernelID != KernelID {
			continue Data
		}

		found := false
	Policy:
		for _, hivePolicy := range hivePolicyList.Items {

			if !hivePolicy.ObjectMeta.DeletionTimestamp.IsZero() {
				continue Policy
			}

		Trap:
			for _, hiveTrap := range hivePolicy.Spec.Traps {

				found, err = HiveDataTrapCmp(hiveData, hiveTrap)
				if err != nil {
					log.Error(err, fmt.Sprintf("Reconcile Error Failed compare HiveData %s and Trap with path %s", hiveData.Name, hiveTrap.Path))
					continue Trap
				}
				if found {
					break Trap
				}
			}
		}

		if !found {
			if err := r.Client.Delete(ctx, &hiveData); err != nil {
				log.Error(err, fmt.Sprintf("Reconciler Error Delete HiveData %s", hiveData.Name))
				continue Data
			}

			log.Info("Deleted HiveData")
			continue Data
		}

		if it > hivebpf.MapMaxEntries {
			log.Error(fmt.Errorf("Number of Traced inodes exceeds the maximum number %d", hivebpf.MapMaxEntries), "Reconcile Error")
			continue Data
		}

		err = hivebpf.UpdateTracedInodes(it, uint64(hiveData.Spec.InodeNo))
		if err != nil {
			log.Error(err, fmt.Sprintf("Reconcile Error Update map with inode %d for HiveData %s", hiveData.Spec.InodeNo, hiveData.Name))
			continue Data
		}
		it++
	}

	// Fill the rest of the eBPF map with zeros so that we do not leave
	// old values that where there before.
	if err = hivebpf.ResetTracedInodes(it); err != nil {
		log.Error(err, fmt.Sprintf("Reconcile Error Update map with empty values"))
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *HiveDataReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&hivev1alpha1.HiveData{}).
		Complete(r)
}
