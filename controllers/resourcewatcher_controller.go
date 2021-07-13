/*
Copyright 2021.

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

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	ibmdevv1alpha1 "github.com/vincent-pli/operator-trigger/api/v1alpha1"
	resources "github.com/vincent-pli/operator-trigger/controllers/resources"
)

const (
	watcherImageEnvVar = "WATCH_IMAGE"
	finalizerName      = "resourcewatchers.tekton.dev/finalizer"
)

// ResourceWatcherReconciler reconciles a ResourceWatcher object
type ResourceWatcherReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	watcherImage string
}

//+kubebuilder:rbac:groups=ibm.dev.asset.ibm,resources=resourcewatchers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ibm.dev.asset.ibm,resources=resourcewatchers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ibm.dev.asset.ibm,resources=resourcewatchers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ResourceWatcher object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ResourceWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// your logic here
	// Fetch the ResourceWatcher instance
	instance := &ibmdevv1alpha1.ResourceWatcher{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// examine DeletionTimestamp to determine if object is under deletion
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(instance.GetFinalizers(), finalizerName) {
			addFinalizer(instance, finalizerName)
			if err := r.Client.Update(context.TODO(), instance); err != nil {
				return reconcile.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(instance.GetFinalizers(), finalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.removeClusterrolebinding(instance); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return reconcile.Result{}, err
			}

			// remove our finalizer from the list and update it.
			removeFinalizer(instance, finalizerName)
			if err := r.Client.Update(context.TODO(), instance); err != nil {
				return reconcile.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return reconcile.Result{}, nil
	}

	// Define a new Rolebinding object
	rolebinding, err := resources.MakeRolebinding(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	// Set ResourceWatcher instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, rolebinding, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Rolebinding already exists
	foundRolebinding := &rbacv1.ClusterRoleBinding{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: rolebinding.Name, Namespace: rolebinding.Namespace}, foundRolebinding)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new Rolebinding", "Rolebinding.Namespace", rolebinding.Namespace, "Rolebinding.Name", rolebinding.Name)
		err = r.Client.Create(context.TODO(), rolebinding)
		if err != nil {
			logger.Error(err, "Create Rolebinding raise exception.")
			return reconcile.Result{}, err
		}
	}

	// Define a new Pod object
	deploy := resources.MakeWatchDeploy(instance, r.watcherImage)

	// Set ResourceWatcher instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, deploy, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &v1.Deployment{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new Pod", "Deployment.Namespace", deploy.Namespace, "Deployment.Name", deploy.Name)
		err = r.Client.Create(context.TODO(), deploy)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}
	return ctrl.Result{}, nil
}

// RemoveFinalizer accepts an Object and removes the provided finalizer if present.
func removeFinalizer(o *ibmdevv1alpha1.ResourceWatcher, finalizer string) {
	f := o.GetFinalizers()
	for i := 0; i < len(f); i++ {
		if f[i] == finalizer {
			f = append(f[:i], f[i+1:]...)
			i--
		}
	}
	o.SetFinalizers(f)
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// AddFinalizer accepts an Object and adds the provided finalizer if not present.
func addFinalizer(o *ibmdevv1alpha1.ResourceWatcher, finalizer string) {
	f := o.GetFinalizers()
	for _, e := range f {
		if e == finalizer {
			return
		}
	}
	o.SetFinalizers(append(f, finalizer))
}

func (r *ResourceWatcherReconciler) removeClusterrolebinding(o *ibmdevv1alpha1.ResourceWatcher) error {
	name := resources.GetRolebindingName(o)

	// Check if this Rolebinding already exists
	foundRolebinding := &rbacv1.ClusterRoleBinding{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: ""}, foundRolebinding)
	if err == nil {
		err = r.Client.Delete(context.TODO(), foundRolebinding)
	}

	if err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ResourceWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ibmdevv1alpha1.ResourceWatcher{}).
		For(&ibmdevv1alpha1.ResourceWatcher{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
