/*


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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	examplev1alpha1 "github.com/example/finalizers/api/v1alpha1"
)

// ServiceInstanceReconciler reconciles a ServiceInstance object
type ServiceInstanceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=example.com,resources=serviceinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.com,resources=serviceinstances/status,verbs=get;update;patch

// Reconcile logic
func (r *ServiceInstanceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("serviceinstance", req.NamespacedName)

	serviceinstance := &examplev1alpha1.ServiceInstance{}
	err := r.Get(ctx, req.NamespacedName, serviceinstance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("serviceinstance resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get serviceinstance")
		return ctrl.Result{}, err
	}

	myFinalizerName := "my-finalizer.example.com"
	if serviceinstance.ObjectMeta.DeletionTimestamp.IsZero() {
		// Inserting finalizer if missing from resource
		if err := r.insertFinalizerIfMissing(ctx, log, serviceinstance, myFinalizerName); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		// The object is being deleted.
		if err := r.removeFinalizerIfExists(ctx, log, serviceinstance, myFinalizerName); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager specifies how the controller is built to watch a CR and
// other resources that are owned and managed by that controller
func (r *ServiceInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&examplev1alpha1.ServiceInstance{}).
		Complete(r)
}

// removeFinalizerIfExists removes finalizer and updates resource if specified
// finalizer exists
func (r *ServiceInstanceReconciler) removeFinalizerIfExists(ctx context.Context, log logr.Logger, instance *examplev1alpha1.ServiceInstance, finalizerName string) error {
	if containsString(instance.GetFinalizers(), finalizerName) {
		log.Info("Removing external resources")
		if err := r.deleteExternalResources(instance); err != nil {
			log.Error(err, "Failed to remove external resources")
			return err
		}
		log.Info("Removed external resources")

		log.Info("Removing my finalizer.")
		instance.SetFinalizers(removeString(instance.GetFinalizers(), finalizerName))
		if err := r.Update(ctx, instance); err != nil {
			log.Error(err, "Failed to remove my finalizer")
			return err
		}
		log.Info("Removed my finalizer")
	}
	return nil
}

// insertFinalizerIfMissing inserts finalizer and updates resource if missing
func (r *ServiceInstanceReconciler) insertFinalizerIfMissing(ctx context.Context, log logr.Logger, instance *examplev1alpha1.ServiceInstance, finalizerName string) error {
	if !containsString(instance.GetFinalizers(), finalizerName) {
		log.Info("Inserting my finalizer")
		instance.SetFinalizers(append(instance.GetFinalizers(), finalizerName))
		if err := r.Update(context.Background(), instance); err != nil {
			log.Error(err, "Failed to insert my finalizer")
			return err
		}
		log.Info("Inserted my finalizer")
	}
	return nil
}

func (r *ServiceInstanceReconciler) deleteExternalResources(instance *examplev1alpha1.ServiceInstance) error {
	// Delete any external resources associated with the resource at hand.
	// Make it idempotent!
	return nil
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
