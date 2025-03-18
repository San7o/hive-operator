/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"errors"

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kuberrors "k8s.io/apimachinery/pkg/api/errors"
	labels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	// Currently only containerd is supported
	containerd "github.com/containerd/containerd"
	containerdCio "github.com/containerd/containerd/cio"
)

const (
	ContainerdAddress = "/run/containerd/containerd.sock"
	Namespace         = "k8s.io"
	KernelIDPath      = "/proc/sys/kernel/random/boot_id"
)

type logVerbosity = int

const (
	logVerbosityError logVerbosity = 0
	logVerbosityInfo  logVerbosity = 1
	logVerbosityDump  logVerbosity = 2
)

// TODO: move into main or somewhere called by main
// TODO: defer containerdClient.Close()
var (
	ContainerdClient  *containerd.Client = nil
	KernelID          string             = ""
	logVerbosityLevel logVerbosity       = logVerbosityInfo
)

// HiveReconciler reconciles a Hive object
type HiveReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hive.dynatrace.com,resources=hives,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hive.dynatrace.com,resources=hives/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hive.dynatrace.com,resources=hives/finalizers,verbs=update

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods/status,verbs=get
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=deployments/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *HiveReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconcile triggered with req: ")

	if err := printPIDs(r.Client, ctx); err != nil {
		log.Error(err, "Error printing pids")
		return ctrl.Result{}, err
	}

	// Logic here
	hive := &hivev1alpha1.Hive{}
	err := r.Client.Get(ctx, req.NamespacedName, hive)
	if err != nil {
		if kuberrors.IsNotFound(err) {
			// Request object not found, could have been
			// deleted after reconcile request.
			// Owned objects are automatically garbage
			// collected. For additional cleanup logic
			// use finalizers.
			// Return and don't requeue
			log.Info("Hive resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request
		log.Error(err, "Failed to get Hive")
		return ctrl.Result{}, nil
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Client.Get(ctx,
		types.NamespacedName{Name: hive.Name, Namespace: hive.Namespace},
		found)
	if err != nil && kuberrors.IsNotFound(err) {
		log.Info("TODO: Creating a new Deployment")
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	log.Info("TODO: Deployment already created")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HiveReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// EnqueueRequestsFromMapFunc enqueues Requests by running a
	// transformation function that outputs a collection of
	// reconcile.Requests on each Event. The reconcile.Requests
	// may be for an arbitrary set of objects defined by some user
	// specified transformation of the source Event. (e.g. trigger
	// Reconciler for a set of objects in response to a cluster
	// resize event caused by adding or deleting a Node)
	//
	// EnqueueRequestsFromMapFunc is frequently used to fan-out
	// updates from one object to one or more other objects of a
	// differing type.
	//
	// For UpdateEvents which contain both a new and old object,
	// the transformation function is run on both objects and both
	// sets of Requests are enqueue.
	podWatchHandler := handler.EnqueueRequestsFromMapFunc(
		// type MapFunc = TypedMapFunc[client.Object, reconcile.Request]
		// type TypedMapFunc[object any, request comparable] func(context.Context, object) []request
		func(ctx context.Context, obj client.Object) []reconcile.Request {
			log := log.FromContext(ctx)
			log.Info("TODO: Reconciling Watched Pod")
			if err := printPIDs(r.Client, ctx); err != nil {
				log.Error(err, "Error Printing PIDs")
			}
			return []reconcile.Request{}
		})

	return ctrl.NewControllerManagedBy(mgr).
		For(&hivev1alpha1.Hive{}).
		// TODO: Watch for Deployment
		Watches(&corev1.Pod{}, podWatchHandler).
		Complete(r)
}

func printPIDs(c client.Client, ctx context.Context) error {
	log := log.FromContext(ctx)
	podList := &corev1.PodList{}
	var err error = nil
	
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

	//logVerbosityLevel = hiveList.Items[0].Spec.LogLevel

	// TODO: get filters from configuration

	// TODO: create the pattern from the configuration
	pattern := "x in (k8s.io)" // test
	// Check that the selector is valid
	_, err = labels.Parse(pattern)
	if err != nil {
		return err
	}

	//labelOpt := client.MatchingLabels{"control-pane": "controller-manager"}
	//namespaceOpt := client.MatchingFields{"status.phase": "k8s.io"}
	if err := c.List(ctx, podList); err != nil {
		return err
	}

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

	if logVerbosityLevel >= logVerbosityDump {
		log.Info("kernelID", "kernelID", KernelID)
	}

	for _, pod := range podList.Items {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if !containerStatus.Ready {
				continue
			}
			if logVerbosityLevel >= logVerbosityDump {
				log.Info("Found container", "Container", containerStatus.Name)
			}

			// Get PIDs
			runtime, id, err := SplitContainerRuntimeID(containerStatus.ContainerID)
			supported := IsContainerRuntimeSupported(runtime)
			if !supported {
				return errors.New("Container runtime " + runtime + " is not suported.")
			}

			if logVerbosityLevel >= logVerbosityDump {
				log.Info("ContainerID", "ID", id)
			}

			// Get local containers from ContainerD
			containers, err := ContainerdClient.Containers(ctx)
			if err != nil {
				return err
			}
			attach := containerdCio.NewAttach()
			for _, container := range containers {
				if container.ID() == id {
					task, err := container.Task(ctx, attach)
					if err != nil {
						return err
					}

					if logVerbosityLevel >= logVerbosityDump {
						log.Info("Found container with PID", "PID", task.Pid())
					}
					inode, devID, err := GetInodeDevID(task.Pid(), "/etc/passwd")
					if err != nil {
						return err
					}

					if logVerbosityLevel >= logVerbosityInfo {
						log.Info("Inode number", "inode", inode)
						log.Info("DevID", "devID", devID)
					}
				} else {
					//log.Info("Container not found", "ID", containerID[1])
				}
			}
		}
	}
	return nil
}
