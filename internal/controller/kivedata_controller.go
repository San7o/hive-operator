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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"

	kivev2alpha1 "github.com/San7o/kivebpf/api/v2alpha1"
	kivebpf "github.com/San7o/kivebpf/internal/controller/ebpf"
)

type KiveDataReconciler struct {
	client.Client
	UncachedClient client.Reader
	Scheme         *runtime.Scheme
}

// +kubebuilder:rbac:groups=kivebpf.san7o.github.io,resources=kivedata,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kivebpf.san7o.github.io,resources=kivedata/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kivebpf.san7o.github.io,resources=kivedata/finalizers,verbs=update

func (r *KiveDataReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := logger.FromContext(ctx)
	log.Info("KiveData reconcile triggered.")

	if !kivebpf.Loaded {
		log.Info("Loading eBPF program")
		if err := kivebpf.LoadEbpf(ctx); err != nil { // Fatal
			return ctrl.Result{}, fmt.Errorf("Reconcile Error Load eBPF program: %w", err)
		}
		go Output(r.UncachedClient)
	}

	kiveDataList := &kivev2alpha1.KiveDataList{}
	err := r.Client.List(ctx, kiveDataList)
	if err != nil { // Fatal
		return ctrl.Result{}, fmt.Errorf("Reconcile Error Failed to get Kive Data resource: %w", err)
	}

	kivePolicyList := &kivev2alpha1.KivePolicyList{}
	err = r.Client.List(ctx, kivePolicyList)
	if err != nil { // Fatal
		return ctrl.Result{}, fmt.Errorf("Reconcile Error Failed to get KivePolicy resource: %w", err)
	}

	var it uint32 = 0

	// Check if each KiveData (referring to this kernel id) does have a
	// corresponding KivePolicy. If it does, then we update the eBPF
	// map with the information from the KiveData. If it doesn't, then
	// the KivePolicy has been eliminated and the KiveData should be
	// deleted.
Data:
	for _, kiveData := range kiveDataList.Items {

		if kiveData.Spec.KernelID != KernelID {
			continue Data
		}

		found := false
	Policy:
		for _, kivePolicy := range kivePolicyList.Items {

			if !kivePolicy.ObjectMeta.DeletionTimestamp.IsZero() {
				continue Policy
			}

		Trap:
			for _, kiveTrap := range kivePolicy.Spec.Traps {

				found, err = KiveDataTrapCmp(kiveData, kiveTrap)
				if err != nil {
					log.Error(err, fmt.Sprintf("Reconcile Error Failed compare KiveData %s and Trap with path %s", kiveData.Name, kiveTrap.Path))
					continue Trap
				}

				if !found {
					continue Trap
				}

				// Check that container exists
				matchingFields := client.MatchingFields{}
				matchingFields["metadata.name"] = kiveData.Annotations["pod-name"]
				matchingFields["metadata.namespace"] = kiveData.Annotations["namespace"]
				matchingFields["spec.nodeName"] = kiveData.Annotations["node-name"]
				podList := &corev1.PodList{}
				err = r.UncachedClient.List(ctx, podList, matchingFields)
				if err != nil {
					log.Error(err, "Reconcile Error Failed to list pods")
					continue Trap
				}

				found = false
				for _, pod := range podList.Items {
					for _, containerStatus := range pod.Status.ContainerStatuses {
						if KiveDataContainerCmp(kiveData, pod, containerStatus) {
							found = true
							break Policy
						}
					}
				}
			}
		}

		if !found {

			if err := r.Client.Delete(ctx, &kiveData); err != nil {
				log.Error(err, fmt.Sprintf("Reconciler Error Delete KiveData %s", kiveData.Name))
				continue Data
			}

			log.Info("Deleted KiveData")
			continue Data
		}

		if it > kivebpf.MapMaxEntries {
			log.Error(fmt.Errorf("Number of Traced inodes exceeds the maximum number %d", kivebpf.MapMaxEntries), "Reconcile Error")
			continue Data
		}

		err = kivebpf.UpdateTracedInodes(it, uint64(kiveData.Spec.InodeNo))
		if err != nil {
			log.Error(err, fmt.Sprintf("Reconcile Error Update map with inode %d for KiveData %s", kiveData.Spec.InodeNo, kiveData.Name))
			continue Data
		}
		it++
	}

	// Fill the rest of the eBPF map with zeros so that we do not leave
	// old values that where there before.
	if err = kivebpf.ResetTracedInodes(it); err != nil {
		log.Error(err, fmt.Sprintf("Reconcile Error Update map with empty values"))
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *KiveDataReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// Index pod name, namespace and ip so we can query a pod
	fieldIndexer := mgr.GetFieldIndexer()
	err := fieldIndexer.IndexField(context.Background(), &corev1.Pod{}, "metadata.name",
		func(rawObj client.Object) []string {
			return []string{rawObj.GetName()}
		})
	if err != nil {
		return fmt.Errorf("SetupWithManager Error Index Pod Name: %w", err)
	}

	err = fieldIndexer.IndexField(context.Background(), &corev1.Pod{}, "metadata.namespace",
		func(rawObj client.Object) []string {
			return []string{rawObj.GetNamespace()}
		})
	if err != nil {
		return fmt.Errorf("SetupWithManager Error Index Pod Namespace: %w", err)
	}

	err = fieldIndexer.IndexField(context.Background(), &corev1.Pod{}, "spec.nodeName",
		func(rawObj client.Object) []string {
			return []string{rawObj.GetNamespace()}
		})
	if err != nil {
		return fmt.Errorf("SetupWithManager Error Index Pod NodeName: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&kivev2alpha1.KiveData{}).
		Complete(r)
}
