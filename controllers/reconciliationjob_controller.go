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

const reconciliationJobFinalizerName = "fintech.io/reconciliationjob-finalizer"

// ReconciliationJobReconciler reconciles a ReconciliationJob object
type ReconciliationJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=fintech.io,resources=reconciliationjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fintech.io,resources=reconciliationjobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=fintech.io,resources=reconciliationjobs/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=fintech.io,resources=reconciliationjobs,verbs=get;list;watch;create;update;patch;delete
// nolint:dupl
func (r *ReconciliationJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	rj := &fintechv1alpha1.ReconciliationJob{}
	if err := r.Get(ctx, req.NamespacedName, rj); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch ReconciliationJob")
		return ctrl.Result{}, err
	}

	if rj.DeletionTimestamp != nil {
		if controllerutil.ContainsFinalizer(rj, reconciliationJobFinalizerName) {
			controllerutil.RemoveFinalizer(rj, reconciliationJobFinalizerName)
			if err := r.Update(ctx, rj); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(rj, reconciliationJobFinalizerName) {
		controllerutil.AddFinalizer(rj, reconciliationJobFinalizerName)
		if err := r.Update(ctx, rj); err != nil {
			return ctrl.Result{}, err
		}
	}

	deployment := &appsv1.Deployment{}
	deploymentName := types.NamespacedName{Name: rj.Name, Namespace: rj.Namespace}

	err := r.Get(ctx, deploymentName, deployment)
	if err != nil && apierrors.IsNotFound(err) {
		deployment = r.constructDeployment(rj)
		if err := controllerutil.SetControllerReference(rj, deployment, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, deployment); err != nil {
			log.Error(err, "unable to create Deployment for ReconciliationJob")
			return ctrl.Result{}, err
		}
		log.Info("created Deployment for ReconciliationJob")
	} else if err != nil {
		log.Error(err, "unable to get Deployment")
		return ctrl.Result{}, err
	}

	rj.Status.Phase = phaseRunning
	rj.Status.Replicas = *deployment.Spec.Replicas
	rj.Status.ReadyReplicas = deployment.Status.ReadyReplicas
	rj.Status.ObservedGeneration = rj.Generation

	if err := r.Status().Update(ctx, rj); err != nil {
		log.Error(err, "unable to update ReconciliationJob status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ReconciliationJobReconciler) constructDeployment(rj *fintechv1alpha1.ReconciliationJob) *appsv1.Deployment {
	replicas := int32(1)
	if rj.Spec.Replicas != nil {
		replicas = *rj.Spec.Replicas
	}

	labels := map[string]string{
		labelApp:      "reconciliation-job",
		labelInstance: rj.Name,
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rj.Name,
			Namespace: rj.Namespace,
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
							Name:  "reconciliation-job",
							Image: rj.Spec.Image,
							Env: []corev1.EnvVar{
								{
									Name:  "BATCH_SIZE",
									Value: string(rune(rj.Spec.BatchSize)),
								},
								{
									Name:  "CHECKPOINT_INTERVAL",
									Value: string(rune(rj.Spec.CheckpointInterval)),
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *ReconciliationJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fintechv1alpha1.ReconciliationJob{}).
		Owns(&appsv1.Deployment{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
