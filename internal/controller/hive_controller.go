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

// HiveReconciler reconciles a Hive object
type HiveReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hive.com,resources=hives,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hive.com,resources=hives/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hive.com,resources=hives/finalizers,verbs=update

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods/status,verbs=get
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=deployments/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *HiveReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	hiveList := &hivev1alpha1.HiveList{}
	err = r.Client.List(ctx, hiveList)
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

	// Check if HiveData resources exist for each hive resource.
	// In case it does not or It is incomplete, create it.
	for _, hive := range hiveList.Items {

		labelMap := make(client.MatchingLabels)
		for _, label := range hive.Spec.Match.Label {
			labelMap[label.Key] = label.Value
		}

		matchingFields := client.MatchingFields{}
		if len(hive.Spec.Match.PodName) != 0 {
			matchingFields["metadata.name"] = hive.Spec.Match.PodName
		}
		if len(hive.Spec.Match.Namespace) != 0 {
			matchingFields["metadata.namespace"] = hive.Spec.Match.Namespace
		}
		if len(hive.Spec.Match.IP) != 0 {
			matchingFields["metadata.podIP"] = hive.Spec.Match.IP
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
					data.Spec.PathName == hive.Spec.Path {
					found = true
					break
				}
			}

			if !found {

				inode, devID, ok, err := getInodeDevidFromPod(ctx, pod, hive)
				if err != nil {
					return ctrl.Result{}, err
				}
				if !ok {
					continue
				}

				// Create a hive data resource
				hiveData := &hivev1alpha1.HiveData{
					ObjectMeta: metav1.ObjectMeta{
						// Unique name
						Name:      "hive-data-" + pod.Name + "-" + pod.Namespace + "-" + strconv.FormatUint(uint64(inode), 10),
						Namespace: "hive-operator-system",
					},
					Spec: hivev1alpha1.HiveDataSpec{
						PathName:     hive.Spec.Path,
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
	// HiveData that does not belong anymore to a HivePolicy when this
	// is deleted.
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

// SetupWithManager sets up the controller with the Manager.
func (r *HiveReconciler) SetupWithManager(mgr ctrl.Manager) error {

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
		For(&hivev1alpha1.Hive{}).
		Complete(r)
}

func getInodeDevidFromPod(ctx context.Context, Pod corev1.Pod, Hive hivev1alpha1.Hive) (uint32, uint64, bool, error) {
	//log := log.FromContext(ctx)
	var err error = nil

	// Get local containers from ContainerD
	containers, err := ContainerdClient.Containers(ctx)
	if err != nil {
		return 0, 0, false, err
	}
	attach := containerdCio.NewAttach()

	for _, containerStatus := range Pod.Status.ContainerStatuses {
		if !containerStatus.Ready {
			// TODO Resend
			continue
		}
		//log.Info("Found container", "Container", containerStatus.Name)

		matched := doesMatchPodPolicy(Pod, Hive)
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
		// log.Info("ContainerID", "ID", id)

		for _, container := range containers {
			if container.ID() == id {
				task, err := container.Task(ctx, attach)
				if err != nil {
					return 0, 0, false, err
				}

				//log.Info("Found container with PID", "PID", task.Pid())
				inode, devID, err := GetInodeDevID(task.Pid(),
					Hive.Spec.Path, Hive.Spec.Create, Hive.Spec.Mode)
				if err != nil {
					return 0, 0, false, err
				}

				//log.Info("Inode number", "inode", inode)
				//log.Info("DevID", "devID", devID)
				return uint32(inode), devID, true, nil
			} else {
				//log.Info("Container not found", "ID", containerID[1])
			}
		}
	}
	return 0, 0, false, nil
}

// Debug function
func printPIDs(c client.Client, ctx context.Context) error {
	log := log.FromContext(ctx)
	podList := &corev1.PodList{}
	var err error = nil

	// Check if a client for containerd exists
	if ContainerdClient != nil {
		serving, err := ContainerdClient.IsServing(ctx)
		if err != nil || !serving {
			return err
		}
	} else {
		opt := containerd.WithDefaultNamespace(Namespace)
		ContainerdClient, err = containerd.New(ContainerdAddress, opt)
		if err != nil {
			return err
		}
	}

	// Get local containers from ContainerD
	containers, err := ContainerdClient.Containers(ctx)
	if err != nil {
		return err
	}
	attach := containerdCio.NewAttach()

	// Read the custom resource
	hiveList := &hivev1alpha1.HiveList{}
	err = c.List(ctx, hiveList)
	if err != nil {
		return err
	}
	if len(hiveList.Items) == 0 {
		// Nothing to do here
		return nil
	}
	if err := c.List(ctx, podList); err != nil {
		return err
	}
	//log.Info("kernelID", "kernelID", KernelID)

	// For each pod
	for _, pod := range podList.Items {
		// For each container in a pod
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if !containerStatus.Ready {
				continue
			}
			//log.Info("Found container", "Container", containerStatus.Name)

			// For each Hive Resource
			for _, hive := range hiveList.Items {

				// Check if the pod is matched
				matched := doesMatchPodPolicy(pod, hive)
				if !matched {
					continue
				}

				// Get PIDs
				runtime, id, err := SplitContainerRuntimeID(containerStatus.ContainerID)
				if err != nil {
					return err
				}
				supported := IsContainerRuntimeSupported(runtime)
				if !supported {
					return errors.New("Container runtime " + runtime + " is not suported.")
				}
				// log.Info("ContainerID", "ID", id)

				// For each container managed by the container runtime
				for _, container := range containers {
					if container.ID() == id {
						task, err := container.Task(ctx, attach)
						if err != nil {
							return err
						}

						//log.Info("Found container with PID", "PID", task.Pid())
						inode, devID, err := GetInodeDevID(task.Pid(),
							hive.Spec.Path, hive.Spec.Create, hive.Spec.Mode)
						if err != nil {
							return err
						}

						log.Info("Inode number", "inode", inode)
						log.Info("DevID", "devID", devID)
					} else {
						//log.Info("Container not found", "ID", containerID[1])
					}
				}
			}
		}
	}
	return nil
}
