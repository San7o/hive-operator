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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
	container "github.com/San7o/hive-operator/internal/controller/container"
)

const (
	KernelIDPath = "/proc/sys/kernel/random/boot_id"
	TrapIdLabel  = "trap-id"
)

var (
	KernelID string = ""
)

type HivePolicyReconciler struct {
	client.Client
	UncachedClient client.Reader
	Scheme         *runtime.Scheme
}

// +kubebuilder:rbac:groups=hive-operator.com,resources=hivepolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hive-operator.com,resources=hivepolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hive-operator.com,resources=hivepolicies/finalizers,verbs=update

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
	// deleted is set to true when an HivePolicy is in deletion. This is
	// used to trigger a reconcile on HiveData
	var deleted bool = false

	log.Info("HivePolicy reconcile triggered.")

	hivePolicyList := &hivev1alpha1.HivePolicyList{}
	err = r.Client.List(ctx, hivePolicyList)
	if err != nil { // Fatal
		return ctrl.Result{}, fmt.Errorf("Reconcile Error Failed to get HivePolicy resource: %w", err)
	}

	// Loop over the HivePolicies and check if all the corresponsing
	// HiveData exist. In case they does not, a new HiveData is created.
Policy:
	for _, hivePolicy := range hivePolicyList.Items {

		// Check if this policy is being deleted
		if !hivePolicy.ObjectMeta.DeletionTimestamp.IsZero() {
			if controllerutil.ContainsFinalizer(&hivePolicy, hivev1alpha1.HivePolicyFinalizerName) {
				deleted = true
				controllerutil.RemoveFinalizer(&hivePolicy, hivev1alpha1.HivePolicyFinalizerName)
				if err := r.Update(ctx, &hivePolicy); err != nil {
					log.Error(err, fmt.Sprintf("Reconcile Error Update finalizer for HivePolicy %s", hivePolicy.Name))
					continue Policy
				}
			}
			continue Policy
		}

	Trap:
		for _, hiveTrap := range hivePolicy.Spec.Traps {

			// Saves which containers are already matched by this trap so
			// that they do not get matched twice and have two different
			// HiveAlerts from the same trap
			matchedContainers := map[string]bool{}

			trapID, err := HiveTrapHashID(hiveTrap)
			if err != nil {
				log.Error(err, fmt.Sprintf("Reconcile Error Generate TrapID for Trap at path %s in HivePolicy %s", hiveTrap.Path, hivePolicy.Name))
				continue Trap
			}

		Match:
			for _, hiveTrapMatch := range hiveTrap.MatchAny {

				labels := client.MatchingLabels{
					TrapIdLabel: trapID,
				}
				hiveDataList := &hivev1alpha1.HiveDataList{}
				err = r.Client.List(ctx, hiveDataList, labels)
				if err != nil { // Fatal
					return ctrl.Result{}, fmt.Errorf("Reconcile Error Failed to get HiveData resource: %w", err)
				}

				// Get Pods that match this HiveTrap
				labelMap := make(client.MatchingLabels)
				labelMap = hiveTrapMatch.MatchLabels
				matchingFields := client.MatchingFields{}
				if len(hiveTrapMatch.PodName) != 0 {
					matchingFields["metadata.name"] = hiveTrapMatch.PodName
				}
				if len(hiveTrapMatch.Namespace) != 0 {
					matchingFields["metadata.namespace"] = hiveTrapMatch.Namespace
				}
				if len(hiveTrapMatch.IP) != 0 {
					matchingFields["metadata.podIP"] = hiveTrapMatch.IP
				}
				podList := &corev1.PodList{}
				err = r.UncachedClient.List(ctx, podList, labelMap, matchingFields)
				if err != nil {
					log.Error(err, "Reconcile Error Failed to list pods")
					continue Match
				}

				for _, pod := range podList.Items {

				Container:
					for _, containerStatus := range pod.Status.ContainerStatuses {

						match, err := RegexMatch(hiveTrapMatch.ContainerName, containerStatus.Name)
						if err != nil {
							log.Error(err, fmt.Sprintf("Reconcile error matching regex %s with container %s", hiveTrapMatch.ContainerName, containerStatus.Name))
							continue Container
						}
						if !match {
							continue Container
						}

						found := false
					Data:
						for _, hiveData := range hiveDataList.Items {
							if HiveDataContainerCmp(hiveData, pod, containerStatus) {
								found = true
								break Data
							}
						}

						if found {
							continue Container
						}

						if pod.Status.Phase != corev1.PodRunning {
							return ctrl.Result{Requeue: true}, nil
						}

						matchID := pod.Name + pod.Namespace + containerStatus.ContainerID
						if _, ok := matchedContainers[matchID]; ok {
							// This container was already registered
							continue Container
						}
						matchedContainers[matchID] = true

						containerData, err := container.GetContainerData(ctx, containerStatus, hiveTrap)
						if err != nil {
							log.Error(err, fmt.Sprintf("Reconcile Error Get contianer data for container %s", containerStatus.Name))
							continue Container
						}
						if containerData.ShouldRequeue {
							return ctrl.Result{Requeue: true}, nil
						}
						if !containerData.IsFound { // Inode was not found
							continue Container
						}
						inode := containerData.Ino

						// Here we are crating a new HiveData since an already existing
						// one for this Pod and this HivePolicy has not been found
						hiveData := &hivev1alpha1.HiveData{
							TypeMeta: metav1.TypeMeta{
								Kind:       "HiveData",
								APIVersion: "hive-operator.com/v1alpha1",
							},
							ObjectMeta: metav1.ObjectMeta{
								// Give it an unique name
								Name:      NewHiveDataName(inode, containerStatus),
								Namespace: hivev1alpha1.Namespace,
								// Annotations are used as information for the HiveAlert
								Annotations: map[string]string{
									"hive_policy_name": hivePolicy.Name,
									"callback":         hiveTrap.Callback,
									"pod_name":         pod.Name,
									"namespace":        pod.Namespace,
									"pod_ip":           pod.Status.PodIPs[0].IP,
									"path":             hiveTrap.Path,
									"container_id":     containerData.ID,
									"container_name":   containerData.Name,
									"node_name":        pod.Spec.NodeName,
								},
								Labels: map[string]string{
									// The trap-id is used to link this HiveData to this trap
									TrapIdLabel: trapID,
								},
							},
							Spec: hivev1alpha1.HiveDataSpec{
								InodeNo:  containerData.Ino,
								KernelID: KernelID,
							},
						}

						err = r.Client.Patch(ctx, hiveData, client.Apply, client.ForceOwnership, client.FieldOwner("hivepolicy-controller"))
						if err != nil {
							log.Error(err, fmt.Sprintf("Reconcile Error patch HiveData resource %s", hiveData.Name))
							continue Container
						}
						log.Info("Created / Updated HiveData resource.")
					}
				}
			}
		}
	}

	if !deleted {
		return ctrl.Result{}, nil
	}

	// Force a reconciliation for HiveData, which will delete any
	// HiveData that does not belong anymore to a HivePolicy.
	// We trigger a reconciliation by updating an annotation with the
	// current time.
	hiveDataList := &hivev1alpha1.HiveDataList{}
	err = r.Client.List(ctx, hiveDataList)
	if err != nil { // Fatal
		return ctrl.Result{}, fmt.Errorf("Reconcile Error Failed to get HiveData resource: %w", err)
	}

	if len(hiveDataList.Items) != 0 {
		hiveData := hiveDataList.Items[0]
		orig := hiveData.DeepCopy()
		if hiveData.Annotations == nil {
			hiveData.Annotations = map[string]string{}
		}
		hiveData.Annotations["force-reconcile"] = time.Now().Format(time.RFC3339)
		err = r.Patch(ctx, &hiveData, client.MergeFrom(orig))
		if err != nil {
			log.Error(err, fmt.Sprintf("Reconcile Error Patch HiveData %s", hiveData.Name))
			return ctrl.Result{}, nil
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
