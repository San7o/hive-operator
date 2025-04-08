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
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
)

var (
	Loaded      bool       = false
	LoadedMutex sync.Mutex = sync.Mutex{}
)

// HiveDataReconciler reconciles a HiveData object
type HiveDataReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hive.dynatrace.com,resources=hivedata,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hive.dynatrace.com,resources=hivedata/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hive.dynatrace.com,resources=hivedata/finalizers,verbs=update

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

	LoadedMutex.Lock()
	if !Loaded {
		log.Info("Loading eBPF program")
		if err := LoadBpf(ctx); err != nil {
			log.Error(nil, "Error loading eBPF program")
			return ctrl.Result{}, err
		}
		Loaded = true
		go LogData(context.Background())
	}
	LoadedMutex.Unlock()

	log.Info("TODO: HiveData reconciler triggered")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HiveDataReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hivev1alpha1.HiveData{}).
		Complete(r)
}
