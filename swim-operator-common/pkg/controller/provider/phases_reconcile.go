package provider

import (
	"context"
	"fmt"
	"strings"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/swim-developer/swim-operator-common/pkg/constants"
	commonreconciler "github.com/swim-developer/swim-operator-common/pkg/reconciler"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func providerReconcileSecret(ctx context.Context, cfg ProviderPhaseConfig, desired *corev1.Secret) error {
	return commonreconciler.ReconcileSecret(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func providerReconcileUnmanagedSecret(ctx context.Context, cfg ProviderPhaseConfig, desired *corev1.Secret) error {
	current := &corev1.Secret{}
	err := cfg.Client.Get(ctx, client.ObjectKeyFromObject(desired), current)
	if errors.IsNotFound(err) {
		return cfg.Client.Create(ctx, desired)
	}
	return client.IgnoreNotFound(err)
}

func providerReconcileServiceMonitor(ctx context.Context, cfg ProviderPhaseConfig, desired *monitoringv1.ServiceMonitor) error {
	return commonreconciler.ReconcileServiceMonitor(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func providerReconcileHPA(ctx context.Context, cfg ProviderPhaseConfig, desired *autoscalingv2.HorizontalPodAutoscaler) error {
	return commonreconciler.ReconcileHPA(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func providerReconcileConfigMap(ctx context.Context, cfg ProviderPhaseConfig, desired *corev1.ConfigMap) error {
	return commonreconciler.ReconcileConfigMap(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func providerReconcileDeployment(ctx context.Context, cfg ProviderPhaseConfig, desired *appsv1.Deployment) error {
	return commonreconciler.ReconcileDeployment(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func providerReconcileService(ctx context.Context, cfg ProviderPhaseConfig, desired *corev1.Service) error {
	return commonreconciler.ReconcileService(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func providerReconcilePVC(ctx context.Context, cfg ProviderPhaseConfig, desired *corev1.PersistentVolumeClaim) error {
	return commonreconciler.ReconcilePVC(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func providerReconcileCertificate(ctx context.Context, cfg ProviderPhaseConfig, desired *certmanagerv1.Certificate) (ctrl.Result, error) {
	return commonreconciler.ReconcileCertificate(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func providerReconcileUnstructured(ctx context.Context, cfg ProviderPhaseConfig, desired *unstructured.Unstructured) error {
	return commonreconciler.ReconcileUnstructured(ctx, cfg.Client, cfg.Owner, desired)
}

func providerReconcileStatefulSet(ctx context.Context, cfg ProviderPhaseConfig, desired *appsv1.StatefulSet) error {
	return commonreconciler.ReconcileStatefulSet(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func providerReconcileServiceAccount(ctx context.Context, cfg ProviderPhaseConfig, desired *corev1.ServiceAccount) error {
	return commonreconciler.ReconcileServiceAccount(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func providerReconcileRole(ctx context.Context, cfg ProviderPhaseConfig, desired *rbacv1.Role) error {
	return commonreconciler.ReconcileRole(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func providerReconcileRoleBinding(ctx context.Context, cfg ProviderPhaseConfig, desired *rbacv1.RoleBinding) error {
	return commonreconciler.ReconcileRoleBinding(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

type ProviderArtemisSSLSecretInput struct {
	P                ProviderBuildParams
	ManagedBy        string
	CertSecretName   string
	TargetSecretName string
}

func ReconcileProviderArtemisSSLSecret(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, in ProviderArtemisSSLSecretInput) error {
	ap := providerArtemisParams(in.P, "")
	ap.Labels = artemisLabelsWithManagedBy(in.P, in.ManagedBy)
	keystorePassword := resources.StrDefault(ap.KeystorePassword, "changeit")
	return commonreconciler.ReconcileArtemisSSLSecretFromPEM(ctx, c, scheme, owner, commonreconciler.ArtemisSSLSecretFromPEMInput{
		CertSecretName:   in.CertSecretName,
		TargetSecretName: in.TargetSecretName,
		KeystorePassword: keystorePassword,
		Labels:           ap.Labels,
	})
}

func ReconcileSharedProviderUnstructured(ctx context.Context, cfg ProviderPhaseConfig, desired *unstructured.Unstructured) error {
	return commonreconciler.ReconcileSharedUnstructured(ctx, cfg.Client, desired, cfg.ManagedByLabel, cfg.ManagedByValue)
}

func ProviderArtemisIngressHost(p ProviderBuildParams, clusterDomain string) string {
	if p.ArtemisIngressHostOverride != "" {
		return p.ArtemisIngressHostOverride
	}
	artemisName := ProviderArtemisName(p)
	if appHost := providerAppHost(p); appHost != "" {
		return deriveArtemisHostFrom(artemisName, appHost)
	}
	return fmt.Sprintf(constants.HostnameSuffix, artemisName, p.Namespace, clusterDomain)
}

func providerAppHost(p ProviderBuildParams) string {
	return p.Strategy.Exposure().HTTPSEdgeHost
}

func deriveArtemisHostFrom(artemisName, appHost string) string {
	if idx := strings.Index(appHost, "."); idx != -1 {
		return artemisName + appHost[idx:]
	}
	return ""
}
