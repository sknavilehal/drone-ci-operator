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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	dronev1alpha1 "github.com/sknavilehal/drone-operator/api/v1alpha1"
)

const (
	droneServerImage = "drone/drone:2"
	droneRunnerImage = "drone/drone-runner-docker:1"
)

// DroneServerReconciler reconciles a DroneServer object
type DroneServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=drone.ocean.dev,resources=droneservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=drone.ocean.dev,resources=droneservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=drone.ocean.dev,resources=droneservers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DroneServer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *DroneServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the DroneServer instance
	ds := &dronev1alpha1.DroneServer{}
	err := r.Get(ctx, req.NamespacedName, ds)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Resource not found. Return and don't requeue.
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// Ensure the DroneServer Deployment exists
	dep := r.deploymentForDroneServer(ds)
	if err := r.applyDeployment(ctx, dep); err != nil {
		log.Error(err, "Failed to apply DroneServer deployment")
		return ctrl.Result{}, err
	}

	// Ensure the DroneRunner Deployment exists
	runnerDep := r.deploymentForDroneRunner(ds)
	if err := r.applyDeployment(ctx, runnerDep); err != nil {
		log.Error(err, "Failed to apply DroneRunner deployment")
		return ctrl.Result{}, err
	}

	// Ensure the DroneServer Service exists
	svc := r.serviceForDroneServer(ds)
	if err := r.applyService(ctx, svc); err != nil {
		log.Error(err, "Failed to apply DroneServer service")
		return ctrl.Result{}, err
	}

	// Ensure the DroneRunner Service exists
	runnerSvc := r.serviceForDroneRunner(ds)
	if err := r.applyService(ctx, runnerSvc); err != nil {
		log.Error(err, "Failed to apply DroneRunner service")
		return ctrl.Result{}, err
	}

	// Optionally: update status here

	return ctrl.Result{}, nil
}

func (r *DroneServerReconciler) deploymentForDroneServer(ds *dronev1alpha1.DroneServer) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ds.Name + "-deployment",
			Namespace: ds.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func() *int32 { var i int32 = 1; return &i }(),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": ds.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": ds.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "drone-server",
							Image: droneServerImage,
							Env: []corev1.EnvVar{
								{
									Name: "DRONE_GITHUB_CLIENT_ID",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: ds.Spec.GithubClientIDSecret.Name,
											},
											Key: ds.Spec.GithubClientIDSecret.Key,
										},
									},
								},
								{
									Name: "DRONE_GITHUB_CLIENT_SECRET",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: ds.Spec.GithubClientSecretSecret.Name,
											},
											Key: ds.Spec.GithubClientSecretSecret.Key,
										},
									},
								},
								{
									Name: "DRONE_RPC_SECRET",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: ds.Spec.SharedSecret.Name,
											},
											Key: "DRONE_RPC_SECRET",
										},
									},
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

}

func (r *DroneServerReconciler) deploymentForDroneRunner(ds *dronev1alpha1.DroneServer) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ds.Name + "-runner",
			Namespace: ds.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func() *int32 { var i int32 = 1; return &i }(),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": ds.Name + "-runner"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": ds.Name + "-runner"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "drone-runner",
							Image: droneRunnerImage,
							Env: []corev1.EnvVar{
								{
									Name: "DRONE_RPC_SECRET",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: ds.Spec.SharedSecret.Name,
											},
											Key: "DRONE_RPC_SECRET",
										},
									},
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 3000,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *DroneServerReconciler) serviceForDroneServer(ds *dronev1alpha1.DroneServer) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ds.Name,
			Namespace: ds.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": ds.Name},
			Ports: []corev1.ServicePort{
				{
					Port: 80,
				},
			},
		},
	}
}

func (r *DroneServerReconciler) serviceForDroneRunner(ds *dronev1alpha1.DroneServer) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ds.Name + "-runner",
			Namespace: ds.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": ds.Name + "-runner"},
			Ports: []corev1.ServicePort{
				{
					Port: 3000,
				},
			},
		},
	}
}

func (r *DroneServerReconciler) applyDeployment(ctx context.Context, dep *appsv1.Deployment) error {
	found := &appsv1.Deployment{}
	err := r.Get(ctx, client.ObjectKey{Name: dep.Name, Namespace: dep.Namespace}, found)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return r.Create(ctx, dep)
		}
		return err
	}

	// Update the found deployment
	found.Spec = dep.Spec
	return r.Update(ctx, found)
}

func (r *DroneServerReconciler) applyService(ctx context.Context, svc *corev1.Service) error {
	found := &corev1.Service{}
	err := r.Get(ctx, client.ObjectKey{Name: svc.Name, Namespace: svc.Namespace}, found)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return r.Create(ctx, svc)
		}
		return err
	}

	// Update the found service
	found.Spec = svc.Spec
	return r.Update(ctx, found)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DroneServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dronev1alpha1.DroneServer{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Named("droneserver").
		Complete(r)
}
