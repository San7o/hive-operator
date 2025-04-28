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
	"strconv"
	"time"

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	// Currently only containerd is supported
	containerd "github.com/containerd/containerd"
	containerdCio "github.com/containerd/containerd/cio"
)

const (
	ContainerdAddress = "/run/containerd/containerd.sock"
	Namespace         = "k8s.io"
	KernelIDPath      = "/proc/sys/kernel/random/boot_id"
)

var (
	ContainerdClient *containerd.Client = nil
	KernelID         string             = ""
)

type HivePolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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
	log.Info("HivePolicy reconcile triggered.")
	var err error = nil

	// Check if a containerd client exists or create one
	if ContainerdClient != nil {
		serving, err := ContainerdClient.IsServing(ctx)
		if err != nil || !serving {
			return ctrl.Result{}, err
		}
	} else {
		opt := containerd.WithDefaultNamespace(Namespace)
		ContainerdClient, err = containerd.New(ContainerdAddress, opt)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	hivePolicyList := &hivev1alpha1.HivePolicyList{}
	err = r.Client.List(ctx, hivePolicyList)
	if err != nil {
		log.Error(err, "Failed to get Hive resource")
		return ctrl.Result{}, nil
	}

	hiveDataList := &hivev1alpha1.HiveDataList{}
	err = r.Client.List(ctx, hiveDataList)
	if err != nil {
		log.Error(err, "Failed to get Hive Data resource")
		return ctrl.Result{}, nil
	}

	// Loop over the HivePolicies and check if all the corresponsing
	// HiveData exist. In case they does not, a new HiveData is created.
	for _, hivePolicy := range hivePolicyList.Items {

		labelMap := make(client.MatchingLabels)
		for _, label := range hivePolicy.Spec.Match.Label {
			labelMap[label.Key] = label.Value
		}

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
		err = r.Client.List(ctx, podList, labelMap, matchingFields)
		if err != nil {
			return ctrl.Result{}, err
		}

		for _, pod := range podList.Items {
			found := false
			for _, data := range hiveDataList.Items {
				if data.Spec.PodName == pod.Name &&
					data.Spec.PodNamespace == pod.Namespace &&
					data.Spec.PathName == hivePolicy.Spec.Path {
					found = true
					break
				}
			}

			if !found {

				inode, devID, ok, err := getInodeDevidFromPod(ctx, pod, hivePolicy)
				if err != nil {
					return ctrl.Result{}, err
				}
				if !ok {
					log.Info("Inode in matched pod not found.")
					continue
				}

				// Here we are crating a new HiveData since an already existing
				// one for this Pod and this HivePolicy has not been found
				hiveData := &hivev1alpha1.HiveData{
					ObjectMeta: metav1.ObjectMeta{
						// Give it an unique name
						Name:      "hive-data-" + pod.Name + "-" + pod.Namespace + "-" + strconv.FormatUint(uint64(inode), 10),
						Namespace: "hive-operator-system",
					},
					Spec: hivev1alpha1.HiveDataSpec{
						PathName:     hivePolicy.Spec.Path,
						PodName:      pod.Name,
						PodNamespace: pod.Namespace,
						PodIP:        pod.Status.PodIPs[0].IP,
						InodeNo:      inode,
						DevID:        devID,
						KernelID:     KernelID,
					},
				}
				err = r.Client.Create(ctx, hiveData)
				if err != nil {
					return ctrl.Result{}, err
				}
				log.Info("Created new Hive Data resource.")
			}
		}
	}

	// Force a reconciliation for HiveData, which will delete any
	// HiveData that does not belong anymore to a HivePolicy. This is
	// necessary to handle deletion of HivePolicies.
	// We trigger a reconciliation by updating an annotation with the
	// current time.
	if len(hiveDataList.Items) != 0 {
		orig := hiveDataList.Items[0].DeepCopy()
		if hiveDataList.Items[0].Annotations == nil {
			hiveDataList.Items[0].Annotations = map[string]string{}
		}
		hiveDataList.Items[0].Annotations["force-reconcile"] = time.Now().Format(time.RFC3339)
		if err := r.Patch(ctx, &hiveDataList.Items[0], client.MergeFrom(orig)); err != nil {
			return ctrl.Result{}, err
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
		return err
	}

	err = fieldIndexer.IndexField(context.Background(), &corev1.Pod{}, "metadata.namespace",
		func(rawObj client.Object) []string {
			return []string{rawObj.GetNamespace()}
		})
	if err != nil {
		return err
	}

	err = fieldIndexer.IndexField(context.Background(), &corev1.Pod{}, "metadata.podIP",
		func(rawObj client.Object) []string {
			pod := rawObj.(*corev1.Pod)
			return []string{pod.Status.PodIP}
		})
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&hivev1alpha1.HivePolicy{}).
		Complete(r)
}

// Return: Inode number, Device id, is found, error
func getInodeDevidFromPod(ctx context.Context, Pod corev1.Pod, HivePolicy hivev1alpha1.HivePolicy) (uint32, uint64, bool, error) {
	//log := log.FromContext(ctx)
	var err error = nil

	if Pod.Status.Phase != corev1.PodRunning {
		// TODO: Resend Reconciliation
		return 0, 0, false, nil
	}

	containers, err := ContainerdClient.Containers(ctx)
	if err != nil {
		return 0, 0, false, err
	}
	attach := containerdCio.NewAttach()

	for _, containerStatus := range Pod.Status.ContainerStatuses {
		if !containerStatus.Ready {
			// TODO: Resend Reconciliation
			continue
		}

		matched := doesMatchPodPolicy(Pod, HivePolicy)
		if !matched {
			continue
		}

		runtime, id, err := SplitContainerRuntimeID(containerStatus.ContainerID)
		if err != nil {
			return 0, 0, false, err
		}
		supported := IsContainerRuntimeSupported(runtime)
		if !supported {
			return 0, 0, false, errors.New("Container runtime " + runtime + " is not suported.")
		}

		for _, container := range containers {
			if container.ID() == id {
				task, err := container.Task(ctx, attach)
				if err != nil {
					return 0, 0, false, err
				}

				inode, devID, err := GetInodeDevID(task.Pid(),
					HivePolicy.Spec.Path, HivePolicy.Spec.Create, HivePolicy.Spec.Mode)
				if err != nil {
					return 0, 0, false, err
				}

				return uint32(inode), devID, true, nil
			}
		}
	}
	return 0, 0, false, nil
}
