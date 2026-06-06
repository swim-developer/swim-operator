package consumer

import (
	"context"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	commonreconciler "github.com/swim-developer/swim-operator-common/pkg/reconciler"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
)

func ReconcileConsumerRBACPhase(ctx context.Context, cfg ConsumerPhaseConfig) error {
	p := cfg.BuildParams
	mb := cfg.ManagedByValue
	if err := consumerReconcileServiceAccount(ctx, cfg, BuildConsumerServiceAccount(p, mb)); err != nil {
		return err
	}
	if err := consumerReconcileRole(ctx, cfg, BuildConsumerRole(p, mb)); err != nil {
		return err
	}
	if err := consumerReconcileRoleBinding(ctx, cfg, BuildConsumerRoleBinding(p, mb)); err != nil {
		return err
	}
	return nil
}

func ReconcileConsumerSecretsBundle(ctx context.Context, cfg ConsumerPhaseConfig) ConsumerSecretsBundle {
	p := cfg.BuildParams
	mb := cfg.ManagedByValue
	bundle := ConsumerSecretsBundle{
		Keystore:   BuildConsumerKeystorePasswordSecret(p, mb),
		Providers:  BuildConsumerProvidersSecret(p, mb),
		KafkaCreds: BuildConsumerKafkaCredentialsSecret(p, mb),
	}
	_ = consumerReconcileSecret(ctx, cfg, bundle.Keystore)
	_ = consumerReconcileSecret(ctx, cfg, bundle.Providers)
	_ = consumerReconcileSecret(ctx, cfg, bundle.KafkaCreds)
	return bundle
}

func applyConsumerCondition(ctx context.Context, cfg ConsumerPhaseConfig, conditionType string, status metav1.ConditionStatus, reason, message string) {
	if cfg.ApplyStatus == nil {
		return
	}
	cond := metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	}
	_ = cfg.ApplyStatus(ctx, cond)
}

func consumerReconcileSecret(ctx context.Context, cfg ConsumerPhaseConfig, desired *corev1.Secret) error {
	return commonreconciler.ReconcileSecret(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func consumerReconcileServiceAccount(ctx context.Context, cfg ConsumerPhaseConfig, desired *corev1.ServiceAccount) error {
	return commonreconciler.ReconcileServiceAccount(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func consumerReconcileRole(ctx context.Context, cfg ConsumerPhaseConfig, desired *rbacv1.Role) error {
	return commonreconciler.ReconcileRole(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func consumerReconcileRoleBinding(ctx context.Context, cfg ConsumerPhaseConfig, desired *rbacv1.RoleBinding) error {
	return commonreconciler.ReconcileRoleBinding(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func consumerReconcileConfigMap(ctx context.Context, cfg ConsumerPhaseConfig, desired *corev1.ConfigMap) error {
	return commonreconciler.ReconcileConfigMap(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func consumerReconcileDeployment(ctx context.Context, cfg ConsumerPhaseConfig, desired *appsv1.Deployment) error {
	return commonreconciler.ReconcileDeployment(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func consumerReconcileService(ctx context.Context, cfg ConsumerPhaseConfig, desired *corev1.Service) error {
	return commonreconciler.ReconcileService(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func consumerReconcilePVC(ctx context.Context, cfg ConsumerPhaseConfig, desired *corev1.PersistentVolumeClaim) error {
	return commonreconciler.ReconcilePVC(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func consumerReconcileCertificate(ctx context.Context, cfg ConsumerPhaseConfig, desired *certmanagerv1.Certificate) (ctrl.Result, error) {
	return commonreconciler.ReconcileCertificate(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func consumerReconcileUnstructured(ctx context.Context, cfg ConsumerPhaseConfig, desired *unstructured.Unstructured) error {
	return commonreconciler.ReconcileUnstructured(ctx, cfg.Client, cfg.Owner, desired)
}

func consumerReconcileServiceMonitor(ctx context.Context, cfg ConsumerPhaseConfig, desired *monitoringv1.ServiceMonitor) error {
	return commonreconciler.ReconcileServiceMonitor(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func consumerReconcileHPA(ctx context.Context, cfg ConsumerPhaseConfig, desired *autoscalingv2.HorizontalPodAutoscaler) error {
	return commonreconciler.ReconcileHPA(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

