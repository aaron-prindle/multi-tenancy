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

package tenant

import (
	"context"
	"fmt"

	tenancyv1alpha1 "github.com/aaron-prindle/tenant-crd/pkg/apis/tenancy/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Tenant Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTenant{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("tenant-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Tenant
	err = c.Watch(&source.Kind{Type: &tenancyv1alpha1.Tenant{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by Tenant - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tenancyv1alpha1.Tenant{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileTenant{}

// ReconcileTenant reconciles a Tenant object
type ReconcileTenant struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Tenant object and makes changes based on the state read
// and what is in the Tenant.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tenancy.k8s.io,resources=tenants,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenancy.k8s.io,resources=tenants/status,verbs=get;update;patch
func (r *ReconcileTenant) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Tenant instance
	instance := &tenancyv1alpha1.Tenant{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			fmt.Println("1")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		fmt.Println("2")
		return reconcile.Result{}, err
	}

	// TODO(user): Change this to be the object type created by your controller
	// Define the desired Deployment object
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: instance.Name + "-namespace",
			// Namespace: instance.Namespace,
		},
		Spec: corev1.NamespaceSpec{},
	}

	// if err := controllerutil.SetControllerReference(instance, namespace, r.scheme); err != nil {
	// 	return reconcile.Result{}, err
	// }
	var rolebindings []*rbacv1.RoleBinding
	fmt.Println("======")
	fmt.Println(len(instance.Spec.Users))
	fmt.Println("======")

	// TODO(aaron-prindle) check that Users are present as validation
	// can remove omitempty
	for _, user := range instance.Spec.Users {
		for _, role := range user.Roles {
			rolebinding := &rbacv1.RoleBinding{
				// this is ok because we know exactly how we want to be serialized
				TypeMeta: metav1.TypeMeta{APIVersion: rbacv1.SchemeGroupVersion.String(), Kind: "Role"},
				ObjectMeta: metav1.ObjectMeta{
					Name:      instance.Name + "-role",
					Namespace: instance.Namespace,
				},
				Subjects: []rbacv1.Subject{
					rbacv1.Subject{
						Kind:     "User",
						Name:     user.Name,
						APIGroup: "rbac.authorization.k8s.io",
					},
				},
				RoleRef: rbacv1.RoleRef{
					Kind:     "Role",
					Name:     role,
					APIGroup: "rbac.authorization.k8s.io",
				},
			}
			rolebindings = append(rolebindings, rolebinding)
		}
	}

	errNamespace := controllerutil.SetControllerReference(instance, namespace, r.scheme)
	if errNamespace != nil {
		for _, rolebinding := range rolebindings {
			errRoleBinding := controllerutil.SetControllerReference(instance, rolebinding, r.scheme)
			if errRoleBinding != nil {
				goto test
			}
		}
		fmt.Println("3")
		return reconcile.Result{}, err
	}
test:

	// TODO(user): Change this for the object type created by your controller
	// Check if the Deployment already exists
	// found := &appsv1.Deployment{}
	found := &corev1.Namespace{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: namespace.Name, Namespace: namespace.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Namespace", "namespace", namespace.Namespace, "name", namespace.Name)
		err = r.Create(context.TODO(), namespace)
		fmt.Println("4")
		fmt.Println(err)
		// return reconcile.Result{}, err
	} // else if err != nil {
	// 	fmt.Println("4")
	// 	return reconcile.Result{}, err
	// }

	fmt.Println("======")
	fmt.Println(len(rolebindings))
	fmt.Println("======")

	// TODO(user): Change this for the object type created by your controller
	// Check if the Deployment already exists
	foundRoleBinding := &rbacv1.RoleBinding{}
	for _, rolebinding := range rolebindings {
		err = r.Get(context.TODO(), types.NamespacedName{Name: rolebinding.Name, Namespace: rolebinding.Namespace}, foundRoleBinding)
		if err != nil && errors.IsNotFound(err) {
			log.Info("Creating RoleBinding", "namespace", rolebinding.Namespace, "name", rolebinding.Name)
			err = r.Create(context.TODO(), rolebinding)
			fmt.Println("5")
			fmt.Println(err)
			return reconcile.Result{}, err
		} // else if err != nil {
		// 	fmt.Println("3")
		// 	return reconcile.Result{}, err
		// }
	}

	fmt.Println("6")
	return reconcile.Result{}, nil
}
