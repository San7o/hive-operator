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

	kivev1alpha1 "github.com/San7o/kivebpf/api/v1alpha1"
	container "github.com/San7o/kivebpf/internal/controller/container"
)

const (
	KernelIDPath = "/proc/sys/kernel/random/boot_id"
	TrapIdLabel  = "trap-id"
)

var (
	KernelID string = ""
)

type KivePolicyReconciler struct {
	client.Client
	UncachedClient client.Reader
	Scheme         *runtime.Scheme
}

// +kubebuilder:rbac:groups=kivebpf.san7o.github.io,resources=kivepolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kivebpf.san7o.github.io,resources=kivepolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kivebpf.san7o.github.io,resources=kivepolicies/finalizers,verbs=update

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods/status,verbs=get
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=deployments/status,verbs=get

// The KivePolicy reconciliation is responsible for the following:
//   - For each KivePolicy, fetch files' information such as the inode
//     number from the matched container.
//   - create KiveData resources with the previously fetched information
//     if not already present.
func (r *KivePolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := log.FromContext(ctx)
	var err error = nil
	// deleted is set to true when an KivePolicy is in deletion. This is
	// used to trigger a reconcile on KiveData
	var deleted bool = false

	log.Info("KivePolicy reconcile triggered.")

	kivePolicyList := &kivev1alpha1.KivePolicyList{}
	err = r.Client.List(ctx, kivePolicyList)
	if err != nil { // Fatal
		return ctrl.Result{}, fmt.Errorf("Reconcile Error Failed to get KivePolicy resource: %w", err)
	}

	// Loop over the KivePolicies and check if all the corresponsing
	// KiveData exist. In case they does not, a new KiveData is created.
Policy:
	for _, kivePolicy := range kivePolicyList.Items {

		// Check if this policy is being deleted
		if !kivePolicy.ObjectMeta.DeletionTimestamp.IsZero() {
			if controllerutil.ContainsFinalizer(&kivePolicy, kivev1alpha1.KivePolicyFinalizerName) {
				deleted = true
				controllerutil.RemoveFinalizer(&kivePolicy, kivev1alpha1.KivePolicyFinalizerName)
				if err := r.Update(ctx, &kivePolicy); err != nil {
					log.Error(err, fmt.Sprintf("Reconcile Error Update finalizer for KivePolicy %s", kivePolicy.Name))
					continue Policy
				}
			}
			continue Policy
		}

	Trap:
		for _, kiveTrap := range kivePolicy.Spec.Traps {

			// Saves which containers are already matched by this trap so
			// that they do not get matched twice and have two different
			// KiveAlerts from the same trap
			matchedContainers := map[string]bool{}

			trapID, err := KiveTrapHashID(kiveTrap)
			if err != nil {
				log.Error(err, fmt.Sprintf("Reconcile Error Generate TrapID for Trap at path %s in KivePolicy %s", kiveTrap.Path, kivePolicy.Name))
				continue Trap
			}

		Match:
			for _, kiveTrapMatch := range kiveTrap.MatchAny {

				labels := client.MatchingLabels{
					TrapIdLabel: trapID,
				}
				kiveDataList := &kivev1alpha1.KiveDataList{}
				err = r.Client.List(ctx, kiveDataList, labels)
				if err != nil { // Fatal
					return ctrl.Result{}, fmt.Errorf("Reconcile Error Failed to get KiveData resource: %w", err)
				}

				// Get Pods that match this KiveTrap
				labelMap := make(client.MatchingLabels)
				labelMap = kiveTrapMatch.MatchLabels
				matchingFields := client.MatchingFields{}
				if len(kiveTrapMatch.PodName) != 0 {
					matchingFields["metadata.name"] = kiveTrapMatch.PodName
				}
				if len(kiveTrapMatch.Namespace) != 0 {
					matchingFields["metadata.namespace"] = kiveTrapMatch.Namespace
				}
				if len(kiveTrapMatch.IP) != 0 {
					matchingFields["metadata.podIP"] = kiveTrapMatch.IP
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

						match, err := RegexMatch(kiveTrapMatch.ContainerName, containerStatus.Name)
						if err != nil {
							log.Error(err, fmt.Sprintf("Reconcile error matching regex %s with container %s", kiveTrapMatch.ContainerName, containerStatus.Name))
							continue Container
						}
						if !match {
							continue Container
						}

						found := false
					Data:
						for _, kiveData := range kiveDataList.Items {
							if KiveDataContainerCmp(kiveData, pod, containerStatus) {
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

						containerData, err := container.GetContainerData(ctx, containerStatus, kiveTrap)
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

						// Here we are crating a new KiveData since an already existing
						// one for this Pod and this KivePolicy has not been found
						kiveData := &kivev1alpha1.KiveData{
							TypeMeta: metav1.TypeMeta{
								Kind:       "KiveData",
								APIVersion: "kivebpf.san7o.github.io/v1alpha1",
							},
							ObjectMeta: metav1.ObjectMeta{
								// Give it an unique name
								Name:      NewKiveDataName(inode, containerStatus),
								Namespace: kivev1alpha1.Namespace,
								// Annotations are used as information for the KiveAlert
								Annotations: map[string]string{
									"kive_policy_name": kivePolicy.Name,
									"callback":         kiveTrap.Callback,
									"pod_name":         pod.Name,
									"namespace":        pod.Namespace,
									"pod_ip":           pod.Status.PodIPs[0].IP,
									"path":             kiveTrap.Path,
									"container_id":     containerData.ID,
									"container_name":   containerData.Name,
									"node_name":        pod.Spec.NodeName,
								},
								Labels: map[string]string{
									// The trap-id is used to link this KiveData to this trap
									TrapIdLabel: trapID,
								},
							},
							Spec: kivev1alpha1.KiveDataSpec{
								InodeNo:  containerData.Ino,
								KernelID: KernelID,
							},
						}

						err = r.Client.Patch(ctx, kiveData, client.Apply, client.ForceOwnership, client.FieldOwner("kivepolicy-controller"))
						if err != nil {
							log.Error(err, fmt.Sprintf("Reconcile Error patch KiveData resource %s", kiveData.Name))
							continue Container
						}
						log.Info("Created / Updated KiveData resource.")
					}
				}
			}
		}
	}

	if !deleted {
		return ctrl.Result{}, nil
	}

	// Force a reconciliation for KiveData, which will delete any
	// KiveData that does not belong anymore to a KivePolicy.
	// We trigger a reconciliation by updating an annotation with the
	// current time.
	kiveDataList := &kivev1alpha1.KiveDataList{}
	err = r.Client.List(ctx, kiveDataList)
	if err != nil { // Fatal
		return ctrl.Result{}, fmt.Errorf("Reconcile Error Failed to get KiveData resource: %w", err)
	}

	if len(kiveDataList.Items) != 0 {
		kiveData := kiveDataList.Items[0]
		orig := kiveData.DeepCopy()
		if kiveData.Annotations == nil {
			kiveData.Annotations = map[string]string{}
		}
		kiveData.Annotations["force-reconcile"] = time.Now().Format(time.RFC3339)
		err = r.Patch(ctx, &kiveData, client.MergeFrom(orig))
		if err != nil {
			log.Error(err, fmt.Sprintf("Reconcile Error Patch KiveData %s", kiveData.Name))
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *KivePolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {

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
		For(&kivev1alpha1.KivePolicy{}).
		Complete(r)
}
