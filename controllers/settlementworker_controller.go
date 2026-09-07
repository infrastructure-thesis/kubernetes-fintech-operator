/*
Copyright 2026.

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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	fintechv1alpha1 "github.com/infrastructure-thesis/kubernetes-fintech-operator/api/v1alpha1"
)

const settlementWorkerFinalizerName = "fintech.io/settlementworker-finalizer"

// SettlementWorkerReconciler reconciles a SettlementWorker object
type SettlementWorkerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=fintech.io,resources=settlementworkers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fintech.io,resources=settlementworkers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=fintech.io,resources=settlementworkers/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile implements the reconciliation loop
func (r *SettlementWorkerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Fetch the SettlementWorker
	sw := &fintechv1alpha1.SettlementWorker{}
	if err := r.Get(ctx, req.NamespacedName, sw); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch SettlementWorker")
		return ctrl.Result{}, err
	}

	// Handle deletion with finalizer
	if sw.DeletionTimestamp != nil {
		if controllerutil.ContainsFinalizer(sw, settlementWorkerFinalizerName) {
			// Cleanup logic here
			controllerutil.RemoveFinalizer(sw, settlementWorkerFinalizerName)
			if err := r.Update(ctx, sw); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(sw, settlementWorkerFinalizerName) {
		controllerutil.AddFinalizer(sw, settlementWorkerFinalizerName)
		if err := r.Update(ctx, sw); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Create or update Deployment
	deployment := &appsv1.Deployment{}
	deploymentName := types.NamespacedName{Name: sw.Name, Namespace: sw.Namespace}

	err := r.Get(ctx, deploymentName, deployment)
	if err != nil && apierrors.IsNotFound(err) {
		// Create new Deployment
		deployment = r.constructDeployment(sw)
		if err := controllerutil.SetControllerReference(sw, deployment, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, deployment); err != nil {
			log.Error(err, "unable to create Deployment for SettlementWorker")
			return ctrl.Result{}, err
		}
		log.Info("created Deployment for SettlementWorker")
	} else if err != nil {
		log.Error(err, "unable to get Deployment")
		return ctrl.Result{}, err
	}

	// Update status
	sw.Status.Phase = "Running"
	sw.Status.Replicas = *deployment.Spec.Replicas
	sw.Status.ReadyReplicas = deployment.Status.ReadyReplicas
	sw.Status.ObservedGeneration = sw.Generation

	if err := r.Status().Update(ctx, sw); err != nil {
		log.Error(err, "unable to update SettlementWorker status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *SettlementWorkerReconciler) constructDeployment(sw *fintechv1alpha1.SettlementWorker) *appsv1.Deployment {
	replicas := int32(1)
	if sw.Spec.Replicas != nil {
		replicas = *sw.Spec.Replicas
	}

	labels := map[string]string{
		"app":      "settlement-worker",
		"instance": sw.Name,
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sw.Name,
			Namespace: sw.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "settlement-worker",
							Image: sw.Spec.Image,
							Env: []corev1.EnvVar{
								{
									Name:  "SETTLEMENT_TYPE",
									Value: sw.Spec.SettlementType,
								},
								{
									Name:  "MAX_IN_FLIGHT",
									Value: string(rune(sw.Spec.MaxInFlightTransactions)),
								},
								{
									Name:  "SCHEDULING_PRIORITY",
									Value: sw.Spec.SchedulingPriority,
								},
							},
						},
					},
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager
func (r *SettlementWorkerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fintechv1alpha1.SettlementWorker{}).
		Owns(&appsv1.Deployment{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
