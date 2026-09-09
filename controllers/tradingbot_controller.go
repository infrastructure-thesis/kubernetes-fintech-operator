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
	"fmt"

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

const tradingBotFinalizerName = "fintech.io/tradingbot-finalizer"

// TradingBotReconciler reconciles a TradingBot object
type TradingBotReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=fintech.io,resources=tradingbots,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fintech.io,resources=tradingbots/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=fintech.io,resources=tradingbots/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=fintech.io,resources=tradingbots,verbs=get;list;watch;create;update;patch;delete
// nolint:dupl
func (r *TradingBotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	tb := &fintechv1alpha1.TradingBot{}
	if err := r.Get(ctx, req.NamespacedName, tb); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch TradingBot")
		return ctrl.Result{}, err
	}

	if tb.DeletionTimestamp != nil {
		if controllerutil.ContainsFinalizer(tb, tradingBotFinalizerName) {
			controllerutil.RemoveFinalizer(tb, tradingBotFinalizerName)
			if err := r.Update(ctx, tb); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(tb, tradingBotFinalizerName) {
		controllerutil.AddFinalizer(tb, tradingBotFinalizerName)
		if err := r.Update(ctx, tb); err != nil {
			return ctrl.Result{}, err
		}
	}

	deployment := &appsv1.Deployment{}
	deploymentName := types.NamespacedName{Name: tb.Name, Namespace: tb.Namespace}

	err := r.Get(ctx, deploymentName, deployment)
	if err != nil && apierrors.IsNotFound(err) {
		deployment = r.constructTradingBotDeployment(tb)
		if err := controllerutil.SetControllerReference(tb, deployment, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, deployment); err != nil {
			log.Error(err, "unable to create Deployment for TradingBot")
			return ctrl.Result{}, err
		}
		log.Info("created Deployment for TradingBot")
	} else if err != nil {
		log.Error(err, "unable to get Deployment")
		return ctrl.Result{}, err
	}

	tb.Status.Phase = phaseRunning
	tb.Status.Replicas = *deployment.Spec.Replicas
	tb.Status.ReadyReplicas = deployment.Status.ReadyReplicas
	tb.Status.ObservedGeneration = tb.Generation

	if err := r.Status().Update(ctx, tb); err != nil {
		log.Error(err, "unable to update TradingBot status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *TradingBotReconciler) constructTradingBotDeployment(tb *fintechv1alpha1.TradingBot) *appsv1.Deployment {
	replicas := int32(1)
	if tb.Spec.Replicas != nil {
		replicas = *tb.Spec.Replicas
	}

	labels := map[string]string{
		labelApp:      "trading-bot",
		labelInstance: tb.Name,
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tb.Name,
			Namespace: tb.Namespace,
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
							Name:  "trading-bot",
							Image: tb.Spec.Image,
							Env: []corev1.EnvVar{
								{
									Name:  "TRADING_PAIR",
									Value: tb.Spec.TradingPair,
								},
								{
									Name:  "CPU_PINNING",
									Value: fmt.Sprintf("%v", tb.Spec.CPUPinning),
								},
								{
									Name:  "NODE_AFFINITY",
									Value: tb.Spec.NodeAffinity,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *TradingBotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fintechv1alpha1.TradingBot{}).
		Owns(&appsv1.Deployment{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
