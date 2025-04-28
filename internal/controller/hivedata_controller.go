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
	"errors"

	"github.com/cilium/ebpf"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
)

const (
	MapMaxEntries = 1024
)

var (
	Loaded bool = false
)

type HiveDataReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hive.com,resources=hivedata,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hive.com,resources=hivedata/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hive.com,resources=hivedata/finalizers,verbs=update

func (r *HiveDataReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("HiveData reconcile triggered.")

	if !Loaded {
		log.Info("Loading eBPF program")
		if err := LoadBpf(ctx); err != nil {
			log.Error(nil, "Error loading eBPF program")
			return ctrl.Result{}, err
		}
		Loaded = true
		go LogData(context.Background())
	}

	hivePolicyList := &hivev1alpha1.HivePolicyList{}
	err := r.Client.List(ctx, hivePolicyList)
	if err != nil {
		log.Error(err, "Failed to get Hive Policy resource")
		return ctrl.Result{}, nil
	}

	hiveDataList := &hivev1alpha1.HiveDataList{}
	err = r.Client.List(ctx, hiveDataList)
	if err != nil {
		log.Error(err, "Failed to get Hive Data resource")
		return ctrl.Result{}, nil
	}

	// Check if each HiveData (referring to this kernel id) does have a
	// corresponding HivePolicy. If it does, then we update the eBPF
	// map with the information from the HiveData. It it doesn't, then
	// the HivePolicy has been eliminated and the HiveData should be
	// deleted.
	var i uint32 = 0
	for _, hiveData := range hiveDataList.Items {
		if hiveData.Spec.KernelID != KernelID {
			continue
		}

		found := false
		for _, hivePolicy := range hivePolicyList.Items {
			if (hiveData.Spec.PathName == hivePolicy.Spec.Path) &&
				(hivePolicy.Spec.Match.PodName != "" ||
					hiveData.Spec.PodName != hivePolicy.Spec.Match.PodName) &&
				(hivePolicy.Spec.Match.Namespace != "" ||
					hiveData.Spec.PodNamespace != hivePolicy.Spec.Match.Namespace) &&
				(hivePolicy.Spec.Match.IP != "" ||
					hiveData.Spec.PodIP != hivePolicy.Spec.Match.IP) {
				// TODO: match labels
				found = true
				break
			}
		}

		if !found {
			if err := r.Client.Delete(ctx, &hiveData); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			if i > MapMaxEntries {
				return ctrl.Result{}, errors.New("Number of Traced inodes exceeds the maximum number")
			}
			err = Objs.TracedInodes.Update(i, uint64(hiveData.Spec.InodeNo), ebpf.UpdateAny)
			if err != nil {
				log.Error(err, "Error Updating map with inode")
				return ctrl.Result{}, err
			}
			i++
		}
	}
	// Fill the rest of the eBPF map with zeros so that we do not leave
	// old values that where there before.
	for ; i < MapMaxEntries; i++ {
		err = Objs.TracedInodes.Update(i, uint64(0), ebpf.UpdateAny)
		if err != nil {
			log.Error(err, "Error Updating map with empty value")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *HiveDataReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&hivev1alpha1.HiveData{}).
		Complete(r)
}
