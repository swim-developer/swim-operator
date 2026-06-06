package consumer

import (
	"context"
	"time"

	"github.com/swim-developer/swim-operator-common/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func ReconcileConsumerClientPhase(ctx context.Context, cfg ConsumerPhaseConfig, bundle ConsumerSecretsBundle) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	p := cfg.BuildParams
	mb := cfg.ManagedByValue
	desiredMongoSecret := BuildConsumerMongoSecret(p, mb)
	desiredClientConfigMap := BuildConsumerConfigMap(p, mb)
	if err := consumerReconcileConfigMap(ctx, cfg, desiredClientConfigMap); err != nil {
		return ctrl.Result{}, err
	}
	if res, err := consumerReconcileCertificate(ctx, cfg, BuildConsumerCertificate(p, mb)); err != nil || res.Requeue {
		return res, err
	}
	var clientHash string
	if p.Flavor == ConsumerFlavorDnotam {
		clientHash = resources.ComputeConfigHash(desiredClientConfigMap, desiredMongoSecret, bundle.Keystore, bundle.Providers, bundle.KafkaCreds)
	} else {
		clientHash = resources.ComputeConfigHash(desiredClientConfigMap, desiredMongoSecret, bundle.Providers, bundle.Keystore, bundle.KafkaCreds)
	}
	if err := consumerReconcileDeployment(ctx, cfg, BuildConsumerClientDeployment(p, mb, clientHash)); err != nil {
		return ctrl.Result{}, err
	}
	if err := consumerReconcileService(ctx, cfg, BuildConsumerClientService(p, mb)); err != nil {
		return ctrl.Result{}, err
	}
	if consumerServiceMonitorEnabled(p) {
		if err := consumerReconcileServiceMonitor(ctx, cfg, BuildConsumerServiceMonitor(p, mb)); err != nil {
			logger.Info("ServiceMonitor creation failed (observability stack may not be available)", "error", err)
		}
	}
	if err := consumerReconcileHPA(ctx, cfg, BuildConsumerHPA(p, mb)); err != nil {
		return ctrl.Result{}, err
	}
	clientReady, err := consumerIsClientReady(ctx, cfg)
	if err != nil {
		return ctrl.Result{}, err
	}
	if !clientReady {
		logger.Info("Client deployment not ready yet, requeuing in 10 seconds")
		applyConsumerCondition(ctx, cfg, "ClientReady", metav1.ConditionFalse, "Provisioning", "Waiting for client pods to be ready")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	applyConsumerCondition(ctx, cfg, "ClientReady", metav1.ConditionTrue, "Ready", "Client pods are ready and healthy")
	return ctrl.Result{}, nil
}

func consumerIsClientReady(ctx context.Context, cfg ConsumerPhaseConfig) (bool, error) {
	p := cfg.BuildParams
	deployment := &appsv1.Deployment{}
	err := cfg.Client.Get(ctx, client.ObjectKey{Name: p.Name, Namespace: p.Namespace}, deployment)
	if err != nil {
		return false, client.IgnoreNotFound(err)
	}
	if deployment.Status.Replicas == 0 {
		return false, nil
	}
	return deployment.Status.ReadyReplicas > 0 && deployment.Status.ReadyReplicas == deployment.Status.Replicas, nil
}
