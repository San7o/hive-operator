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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
)

const (
	MapMaxEntries = 1024
)

var (
	Loaded bool = false
)

// HiveDataReconciler reconciles a HiveData object
type HiveDataReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hive.com,resources=hivedata,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hive.com,resources=hivedata/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hive.com,resources=hivedata/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HiveData object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
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

	hivePolicyList := &hivev1alpha1.HiveList{}
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

	// Check if all HiveData (referring to this kernel id) have a
	// corresponding HivePolicy
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
			// Update the eBPF map
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
	for ; i < MapMaxEntries; i++ {
		err = Objs.TracedInodes.Update(i, uint64(0), ebpf.UpdateAny)
		if err != nil {
			log.Error(err, "Error Updating map with empty value")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HiveDataReconciler) SetupWithManager(mgr ctrl.Manager) error {

	podWatchHandler := handler.EnqueueRequestsFromMapFunc(
		// There are two main operations we are concearned about with
		// pods: pod creation and pod deletion.
		// - creation: upon creation, the controller should send a
		//   reconcile request for Hive Policy so that new a new HiveData
		//   will be generated for the new pod.
		// - deletion: upon deletion, the controller should check each
		//   HiveData if it refers to an existing pod. If it doesn't, then
		//   that resource should be eliminated.
		func(ctx context.Context, obj client.Object) []reconcile.Request {

			log := log.FromContext(ctx)
			log.Info("TODO: Reconciling Watched Pod")

			// TODO

			return []reconcile.Request{}
		})

	return ctrl.NewControllerManagedBy(mgr).
		For(&hivev1alpha1.HiveData{}).
		Watches(&corev1.Pod{}, podWatchHandler).
		Complete(r)
}
