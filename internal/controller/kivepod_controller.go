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
	"time"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kivev1alpha1 "github.com/San7o/kivebpf/api/v1alpha1"
)

type KivePodReconciler struct {
	client.Client
}

// There are two main operations we are concearned about with
// pods: pod creation and pod termination.
//   - creation: upon creation, the controller should send a
//     reconcile request for KivePolicy so that new KiveData will
//     be generated for the new pod.
//   - termination: upon termination, the controller should check if
//     each KiveData refers to an existing pod. If it doesn't, then
//     that resource should be eliminated.
//
// Failures are treated as terminations.
func (r *KivePodReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {

	log := log.FromContext(ctx)
	log.Info("Pod watch event triggered.")

	kiveDataList := &kivev1alpha1.KiveDataList{}
	err := r.Client.List(ctx, kiveDataList)
	if err != nil { // Fatal
		return reconcile.Result{}, fmt.Errorf("Reconcile Error Failed to get KiveData resource: %w", err)
	}
	podList := &corev1.PodList{}
	err = r.Client.List(ctx, podList)
	if err != nil { // Fatal
		return reconcile.Result{}, fmt.Errorf("Reconcile Error Failed to get PodList resource: %w", err)
	}

Data:
	for _, kiveData := range kiveDataList.Items {

		if kiveData.Spec.KernelID != KernelID {
			// We are only concearned about the kiveData on this machine,
			// to avoid conflicts.
			continue Data
		}
		found := false

	Pod:
		for _, pod := range podList.Items {
			if kiveData.Annotations["pod_name"] == pod.Name &&
				kiveData.Annotations["namespace"] == pod.Namespace &&
				kiveData.Annotations["pod_ip"] == pod.Status.PodIPs[0].IP &&
				// If the pod has terminated or has failed, we want to
				// remove the KiveData so that it will be regenerated
				// later during the reconciliation of KivePolicy. This
				// is needed because the inode number or kernel id may
				// change when a pod gets restarted / rescheduled.
				pod.Status.Phase != corev1.PodSucceeded &&
				pod.Status.Phase != corev1.PodFailed {
				found = true
				break Pod
			}
		}

		if !found {
			if err := r.Client.Delete(ctx, &kiveData); err != nil {
				log.Error(err, fmt.Sprintf("Reconcile Error Deleting KiveData %s after pod event", kiveData.Name))
				continue Data
			}
		}
	}

	// Trigger a KivePolicy reconciliation event to handle
	// pod creation. If a pod is not yet ready, the KiveData
	// Reconciliation should loop until all the pods are ready.
	kivePolicyList := &kivev1alpha1.KivePolicyList{}
	err = r.Client.List(ctx, kivePolicyList)
	if err != nil { // Fatal
		return reconcile.Result{}, fmt.Errorf("Reconcile Error Failed to get Kive Policy resource: %w", err)
	}
	if len(kivePolicyList.Items) != 0 {
		kivePolicy := kivePolicyList.Items[0]
		orig := kivePolicy.DeepCopy()
		if kivePolicy.Annotations == nil {
			kivePolicy.Annotations = map[string]string{}
		}
		kivePolicy.Annotations["force-reconcile"] = time.Now().Format(time.RFC3339)
		if err = r.Patch(ctx, &kivePolicy, client.MergeFrom(orig)); err != nil {
			log.Error(err, fmt.Sprintf("Reconcile Error Patch KivePolicy %s", kivePolicy.Name))
			return ctrl.Result{}, nil
		}
	}
	return reconcile.Result{}, nil
}

func (r *KivePodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}
