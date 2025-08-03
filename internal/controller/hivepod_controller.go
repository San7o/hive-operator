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

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
)

type HivePodReconciler struct {
	client.Client
}

// There are two main operations we are concearned about with
// pods: pod creation and pod termination.
//   - creation: upon creation, the controller should send a
//     reconcile request for HivePolicy so that new HiveData will
//     be generated for the new pod.
//   - termination: upon termination, the controller should check if
//     each HiveData refers to an existing pod. If it doesn't, then
//     that resource should be eliminated.
//
// Failures are treated as terminations.
func (r *HivePodReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {

	log := log.FromContext(ctx)
	log.Info("Pod watch event triggered.")

	hiveDataList := &hivev1alpha1.HiveDataList{}
	err := r.Client.List(ctx, hiveDataList)
	if err != nil { // Fatal
		return reconcile.Result{}, fmt.Errorf("Reconcile Error Failed to get HiveData resource: %w", err)
	}
	podList := &corev1.PodList{}
	err = r.Client.List(ctx, podList)
	if err != nil { // Fatal
		return reconcile.Result{}, fmt.Errorf("Reconcile Error Failed to get PodList resource: %w", err)
	}

Data:
	for _, hiveData := range hiveDataList.Items {

		if hiveData.Spec.KernelID != KernelID {
			// We are only concearned about the hiveData on this machine,
			// to avoid conflicts.
			continue Data
		}
		found := false

	Pod:
		for _, pod := range podList.Items {
			if hiveData.Annotations["pod_name"] == pod.Name &&
				hiveData.Annotations["namespace"] == pod.Namespace &&
				hiveData.Annotations["pod_ip"] == pod.Status.PodIPs[0].IP &&
				// If the pod has terminated or has failed, we want to
				// remove the HiveData so that it will be regenerated
				// later during the reconciliation of HivePolicy. This
				// is needed because the inode number or kernel id may
				// change when a pod gets restarted / rescheduled.
				pod.Status.Phase != corev1.PodSucceeded &&
				pod.Status.Phase != corev1.PodFailed {
				found = true
				break Pod
			}
		}

		if !found {
			if err := r.Client.Delete(ctx, &hiveData); err != nil {
				log.Error(err, fmt.Sprintf("Reconcile Error Deleting HiveData %s after pod event", hiveData.Name))
				continue Data
			}
		}
	}

	// Trigger a HivePolicy reconciliation event to handle
	// pod creation. If a pod is not yet ready, the HiveData
	// Reconciliation should loop until all the pods are ready.
	hivePolicyList := &hivev1alpha1.HivePolicyList{}
	err = r.Client.List(ctx, hivePolicyList)
	if err != nil { // Fatal
		return reconcile.Result{}, fmt.Errorf("Reconcile Error Failed to get Hive Policy resource: %w", err)
	}
	if len(hivePolicyList.Items) != 0 {
		hivePolicy := hivePolicyList.Items[0]
		orig := hivePolicy.DeepCopy()
		if hivePolicy.Annotations == nil {
			hivePolicy.Annotations = map[string]string{}
		}
		hivePolicy.Annotations["force-reconcile"] = time.Now().Format(time.RFC3339)
		if err = r.Patch(ctx, &hivePolicy, client.MergeFrom(orig)); err != nil {
			log.Error(err, fmt.Sprintf("Reconcile Error Patch HivePolicy %s", hivePolicy.Name))
			return ctrl.Result{}, nil
		}
	}
	return reconcile.Result{}, nil
}

func (r *HivePodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}
