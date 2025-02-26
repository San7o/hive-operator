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
	"strings"

	hivev1alpha1 "github.com/San7o/hive-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kuberrors "k8s.io/apimachinery/pkg/api/errors"
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

	if err = printPIDs(r.Client, ctx); err != nil {
		log.Error(err, "Error printing pids")
		return ctrl.Result{}, err
	}

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

// TODO: Add support to more container runtimes
func printPIDs(c client.Client, ctx context.Context) error {
	log := log.FromContext(ctx)
	podList := &corev1.PodList{}

	// Note that Client.List can also use a filter
	if err := c.List(ctx, podList); err != nil {
		return err
	}

	containerdAddress := "/run/containerd/containerd.sock"
	namespace := "k8s.io"
	opt := containerd.WithDefaultNamespace(namespace)
	containerdClient, err := containerd.New(containerdAddress, opt)
	if err != nil {
		return err
	}
	defer containerdClient.Close()

	for _, pod := range podList.Items {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if !containerStatus.Ready {
				continue
			}

			log.Info("Found container", "Container", containerStatus.Name)

			// Get PIDs

			// containerID is of the form "<type>://<container_id>".
			// For example, the type could be "containerd"
			containerID := strings.SplitN(containerStatus.ContainerID, "://", 2)

			if len(containerID) != 2 {
				return errors.New("Error parsing containerID")
			}
			if containerID[0] != "containerd" {
				return errors.New("Container runtimes other that containerd are not supported yet")
			}

			log.Info("ContainerID", "ID", containerID[1])

			// Get containers from ContainerD
			containers, err := containerdClient.Containers(ctx)
			if err != nil {
				return err
			}
			attach := containerdCio.NewAttach()
			for _, container := range containers {
				if container.ID() == containerID[1] {
					task, err := container.Task(ctx, attach)
					if err != nil {
						return err
					}
					log.Info("Found container PID", "PID", task.Pid())
				} else {
					//log.Info("Container not found", "ID", containerID[1])
				}
			}
		}
	}
	return nil
}
