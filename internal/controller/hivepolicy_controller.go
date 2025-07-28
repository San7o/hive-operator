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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
	container "github.com/San7o/hive-operator/internal/controller/container"
)

const (
	KernelIDPath = "/proc/sys/kernel/random/boot_id"
)

var (
	KernelID string = ""
)

type HivePolicyReconciler struct {
	client.Client
	UncachedClient client.Reader
	Scheme         *runtime.Scheme
}

// +kubebuilder:rbac:groups=hive.com,resources=hivepolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hive.com,resources=hivepolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hive.com,resources=hivepolicies/finalizers,verbs=update

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods/status,verbs=get
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=deployments/status,verbs=get

// The HivePolicy reconciliation is responsible for the following:
//   - For each HivePolicy, fetch files' information such as the inode
//     number from the matched container.
//   - create HiveData resources with the previously fetched information
//     if not already present.
func (r *HivePolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := log.FromContext(ctx)
	var err error = nil

	log.Info("HivePolicy reconcile triggered.")

	hivePolicyList := &hivev1alpha1.HivePolicyList{}
	err = r.Client.List(ctx, hivePolicyList)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("Reconcile Error Failed to get HivePolicy resource: %w", err)
	}

	// Loop over the HivePolicies and check if all the corresponsing
	// HiveData exist. In case they does not, a new HiveData is created.
	for _, hivePolicy := range hivePolicyList.Items {

		// Get HiveData resources associated with this HivePolicy
		labels := client.MatchingLabels{
			"policy-id": hivePolicy.Labels["policy-id"],
		}
		hiveDataList := &hivev1alpha1.HiveDataList{}
		err = r.Client.List(ctx, hiveDataList, labels)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("Reconcile Error Failed to get HiveData resource: %w", err)
		}

		// Get Pods that match this HivePolicy
		labelMap := make(client.MatchingLabels)
		labelMap = hivePolicy.Spec.Match.MatchLabels
		matchingFields := client.MatchingFields{}
		if len(hivePolicy.Spec.Match.PodName) != 0 {
			matchingFields["metadata.name"] = hivePolicy.Spec.Match.PodName
		}
		if len(hivePolicy.Spec.Match.Namespace) != 0 {
			matchingFields["metadata.namespace"] = hivePolicy.Spec.Match.Namespace
		}
		if len(hivePolicy.Spec.Match.IP) != 0 {
			matchingFields["metadata.podIP"] = hivePolicy.Spec.Match.IP
		}
		podList := &corev1.PodList{}
		err = r.UncachedClient.List(ctx, podList, labelMap, matchingFields)
		if err != nil {
			return ctrl.Result{}, err
		}

		for _, pod := range podList.Items {

			for _, containerStatus := range pod.Status.ContainerStatuses {

				// HivePolicy match check
				if hivePolicy.Spec.Match.ContainerName != "" &&
					hivePolicy.Spec.Match.ContainerName != containerStatus.Name {
					continue
				}

				found := false
				for _, hiveData := range hiveDataList.Items {
					if HiveDataContainerCmp(hiveData, pod, containerStatus) {
						found = true
						break
					}
				}

				if found {
					continue
				}

				if pod.Status.Phase != corev1.PodRunning {
					return ctrl.Result{Requeue: true}, nil
				}

				containerData, err := container.GetContainerData(ctx, containerStatus, hivePolicy)
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("Reconcile Error Get contianer data: %w", err)
				}
				if containerData.ShouldRequeue {
					return ctrl.Result{Requeue: true}, nil
				}
				if !containerData.IsFound { // Inode was not found
					// DEBUG
					//log.Info("Inode of " + hivePolicy.Spec.Path + " in matched pod " + pod.Name + " not found.")
					continue
				}
				inode := containerData.Ino

				policyID, err := PolicyHashID(hivePolicy)
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("Reconcile Error calculating policy ID: %w", err)
				}

				// Here we are crating a new HiveData since an already existing
				// one for this Pod and this HivePolicy has not been found
				hiveData := &hivev1alpha1.HiveData{
					ObjectMeta: metav1.ObjectMeta{
						// Give it an unique name
						Name:      NewHiveDataName(inode, containerStatus),
						Namespace: hivev1alpha1.Namespace,
						Annotations: map[string]string{
							"hive_policy_name": hivePolicy.Name,
							"callback":         hivePolicy.Spec.Callback,
							"pod_name":         pod.Name,
							"namespace":        pod.Namespace,
							"pod_ip":           pod.Status.PodIPs[0].IP,
							"path":             hivePolicy.Spec.Path,
							"container_id":     containerData.ID,
							"container_name":   containerData.Name,
							"node_name":        pod.Spec.NodeName,
						},
						Labels: map[string]string{
							"policy-id": policyID,
						},
					},
					Spec: hivev1alpha1.HiveDataSpec{
						InodeNo:  containerData.Ino,
						KernelID: KernelID,
					},
				}
				for label, value := range hivePolicy.Spec.Match.MatchLabels {
					hiveData.Annotations["match-label-"+label] = value
				}

				err = r.Client.Create(ctx, hiveData)
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("Reconcile Error Create HiveData resource: %w", err)
				}
				log.Info("Created new HiveData resource.")

				orig := hivePolicy.DeepCopy()
				hivePolicy.ObjectMeta.Labels = map[string]string{
					"policy-id": policyID,
				}
				err = r.Client.Patch(ctx, &hivePolicy, client.MergeFrom(orig))
				if err != nil {
					return ctrl.Result{}, fmt.Errorf("Reconcile Error Update PolicyID in HivePolicy: %w", err)
				}
			}
		}
	}

	// Force a reconciliation for HiveData, which will delete any
	// HiveData that does not belong anymore to a HivePolicy. This is
	// necessary to handle deletion of HivePolicies.
	// We trigger a reconciliation by updating an annotation with the
	// current time.
	hiveDataList := &hivev1alpha1.HiveDataList{}
	err = r.Client.List(ctx, hiveDataList)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("Reconcile Error Failed to get HiveData resource: %w", err)
	}

	if len(hiveDataList.Items) != 0 {
		orig := hiveDataList.Items[0].DeepCopy()
		if hiveDataList.Items[0].Annotations == nil {
			hiveDataList.Items[0].Annotations = map[string]string{}
		}
		hiveDataList.Items[0].Annotations["force-reconcile"] = time.Now().Format(time.RFC3339)
		err = r.Patch(ctx, &hiveDataList.Items[0], client.MergeFrom(orig))
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("Reconcile Error Patch HiveData: %w", err)
		}
	}

	return ctrl.Result{}, nil
}

func (r *HivePolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {

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

	err = fieldIndexer.IndexField(context.Background(), &corev1.Pod{}, "metadata.podIP",
		func(rawObj client.Object) []string {
			pod := rawObj.(*corev1.Pod)
			return []string{pod.Status.PodIP}
		})
	if err != nil {
		return fmt.Errorf("SetupWithManager Error Index Pod Ip: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&hivev1alpha1.HivePolicy{}).
		Complete(r)
}
