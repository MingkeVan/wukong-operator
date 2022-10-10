/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"
	"reflect"

	wukongv1alpha1 "github.com/MingkeVan/wukong-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WukongEnvReconciler reconciles a WukongEnv object
type WukongEnvReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=wukong.wukong.io,resources=wukongenvs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=wukong.wukong.io,resources=wukongenvs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=wukong.wukong.io,resources=wukongenvs/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the WukongEnv object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *WukongEnvReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// TODO(user): your logic here
	// log := r.Log.WithValues("wukongEnv", req.NamespacedName)

	fmt.Println("curren namespace", req.NamespacedName.Name, req.NamespacedName.Namespace)
	// Fetch the Memcached instance
	wukongEnv := &wukongv1alpha1.WukongEnv{}
	err := r.Get(ctx, req.NamespacedName, wukongEnv)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			fmt.Println("WukongEnv resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		fmt.Println(err, "Failed to get WukongEnv")
		return ctrl.Result{}, err
	}

	namespaces := wukongEnv.Spec.Namespaces

	for _, name := range namespaces {
		envNs := &corev1.Namespace{}
		err := r.Get(ctx, client.ObjectKey{
			Name: name,
		}, envNs)
		if err != nil {
			if errors.IsNotFound(err) {
				// Request object not found, could have been deleted after reconcile request.
				// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
				// Return and don't requeue
				fmt.Println("Namespace not found. Ignoring since object must be deleted")
				return ctrl.Result{}, nil
			}
			// Error reading the object - requeue the request.
			fmt.Println(err, "Failed to get Namespace")
			return ctrl.Result{}, err
		}

	}

	apps := wukongEnv.Spec.Apps

	for _, app := range apps {
		// Fetch the Memcached instance
		ns := app.Namespace
		appName := app.Name

		// Check if the deployment already exists, if not create a new one
		found := &v1.Deployment{}
		err = r.Get(ctx, client.ObjectKey{
			Namespace: ns,
			Name:      appName,
		}, found)
		if err != nil && errors.IsNotFound(err) {
			// Define a new deployment
			dep := r.deploymentForWukongApp(wukongEnv, &app)
			fmt.Println("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			err = r.Create(ctx, dep)
			if err != nil {
				fmt.Println(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
				return ctrl.Result{}, err
			}
			// Deployment created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		} else if err != nil {
			fmt.Println(err, "Failed to get Deployment")
			return ctrl.Result{}, err
		}

		fmt.Println(ns, appName, "Deployment exists")

		// Ensure the deployment size is the same as the spec
		size := app.Size
		if *found.Spec.Replicas != size {
			found.Spec.Replicas = &size
			err = r.Update(ctx, found)
			if err != nil {
				fmt.Println(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
				return ctrl.Result{}, err
			}
			// Spec updated - return and requeue
			return ctrl.Result{Requeue: true}, nil
		}

		// Update the Memcached status with the pod names
		// List the pods for this memcached's deployment
		podList := &corev1.PodList{}
		listOpts := []client.ListOption{
			client.InNamespace(""),
			client.MatchingLabels(labelsForWukongApp(wukongEnv.Name)),
		}
		if err = r.List(ctx, podList, listOpts...); err != nil {
			fmt.Println(err, "Failed to list pods", "WukongEnv.Name", wukongEnv.Name)
			return ctrl.Result{}, err
		}
		fmt.Println("pod list", podList.Items)
		podNames := getPodNames(podList.Items)

		// Update status.Nodes if needed
		if !reflect.DeepEqual(podNames, wukongEnv.Status.Pods) {
			wukongEnv.Status.Pods = podNames
			err := r.Status().Update(ctx, wukongEnv)
			if err != nil {
				fmt.Println(err, "Failed to update Memcached status")
				return ctrl.Result{}, err
			}
		}

	}
	return ctrl.Result{}, nil
}

// deploymentForWukongApp returns a wukongApp Deployment object
func (r *WukongEnvReconciler) deploymentForWukongApp(env *wukongv1alpha1.WukongEnv, app *wukongv1alpha1.WukongApp) *v1.Deployment {
	ls := labelsForWukongApp(env.Name)
	replicas := app.Size
	image := app.Image

	dep := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: image,
						Name:  app.Name,
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          app.Name,
						}},
					}},
				},
			},
		},
	}
	// Set Memcached instance as the owner and controller
	ctrl.SetControllerReference(env, dep, r.Scheme)
	return dep
}

// labelsForMemcached returns the labels for selecting the resources
// belonging to the given memcached CR name.
func labelsForWukongApp(name string) map[string]string {
	return map[string]string{"app": "wukongApp", "wukong_app_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// SetupWithManager sets up the controller with the Manager.
func (r *WukongEnvReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&wukongv1alpha1.WukongEnv{}).
		Owns(&v1.Deployment{}).
		Owns(&v1.StatefulSet{}).
		Owns(&corev1.Namespace{}).
		Complete(r)
}
