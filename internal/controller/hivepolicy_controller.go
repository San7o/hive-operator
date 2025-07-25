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
	"strconv"
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
//     number from the matched contianers.
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

	hiveDataList := &hivev1alpha1.HiveDataList{}
	err = r.Client.List(ctx, hiveDataList)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("Reconcile Error Failed to get HiveData resource: %w", err)
	}

	// Loop over the HivePolicies and check if all the corresponsing
	// HiveData exist. In case they does not, a new HiveData is created.
	for _, hivePolicy := range hivePolicyList.Items {

		labelMap := make(client.MatchingLabels)
		labelMap = hivePolicy.Spec.Match.Labels
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
			found := false
			for _, hiveData := range hiveDataList.Items {
				if HiveDataPolicyCmp(hiveData, hivePolicy) &&
					HiveDataPodCmp(hiveData, pod) {
					found = true
					break
				}
			}

			if found {
				continue
			}

			containerData, err := container.GetContainerData(ctx, pod, hivePolicy)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("Reconcile Error Get contianer data: %w", err)
			}
			if containerData.ShouldRequeue {
				return ctrl.Result{Requeue: true}, nil
			}
			if !containerData.IsFound {
				// DEBUG
				//log.Info("Inode of " + hivePolicy.Spec.Path + " in matched pod " + pod.Name + " not found.")
				continue
			}
			inode := containerData.Ino

			// Here we are crating a new HiveData since an already existing
			// one for this Pod and this HivePolicy has not been found
			hiveData := &hivev1alpha1.HiveData{
				ObjectMeta: metav1.ObjectMeta{
					// Give it an unique name
					Name:      "hive-data-" + pod.Name + "-" + pod.Namespace + "-" + strconv.FormatUint(uint64(inode), 10),
					Namespace: "hive-operator-system",

					Annotations: map[string]string{
						"hive_policy_name": hivePolicy.Name,
						"callback":         hivePolicy.Spec.Callback,
						"pod_name":         pod.Name,
						"namespace":        pod.Namespace,
						"pod_ip":           pod.Status.PodIPs[0].IP,
						"path":             hivePolicy.Spec.Path,
						"container_id":     containerData.ContainerID,
						"container_name":   containerData.ContainerName,
					},
				},
				Spec: hivev1alpha1.HiveDataSpec{
					InodeNo:  containerData.Ino,
					KernelID: KernelID,
				},
			}
			for label, value := range hivePolicy.Spec.Match.Labels {
				hiveData.Annotations["match-label-"+label] = value
			}

			err = r.Client.Create(ctx, hiveData)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("Reconcile Error Create HiveData resource: %w", err)
			}
			log.Info("Created new HiveData resource.")
		}
	}

	// Force a reconciliation for HiveData, which will delete any
	// HiveData that does not belong anymore to a HivePolicy. This is
	// necessary to handle deletion of HivePolicies.
	// We trigger a reconciliation by updating an annotation with the
	// current time.
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
		if err := r.Patch(ctx, &hiveDataList.Items[0], client.MergeFrom(orig)); err != nil {
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
