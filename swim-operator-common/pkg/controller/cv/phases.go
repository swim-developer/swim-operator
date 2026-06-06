package cv

import (
	"context"
	"fmt"
	"time"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/swim-developer/swim-operator-common/pkg/constants"
	commonreconciler "github.com/swim-developer/swim-operator-common/pkg/reconciler"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func applyCVCondition(ctx context.Context, cfg CVPhaseConfig, conditionType string, status metav1.ConditionStatus, reason, message string) {
	if cfg.ApplyStatus == nil {
		return
	}
	_ = cfg.ApplyStatus(ctx, metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})
}

func reconcileCVIngressAndNetworking(ctx context.Context, cfg CVPhaseConfig) (
	clusterDomain, artemisInternal, artemisIngress, appTLSHost, amqpInternal string,
	artemisManaged bool,
) {
	p := cfg.BuildParams
	resolve := cfg.ResolveClusterDomain
	if resolve == nil {
		resolve = func(_ context.Context, spec, _ string) string { return spec }
	}
	clusterDomain = resolve(ctx, p.Spec.Global.ClusterDomain, p.Namespace)
	var appHost string
	artemisInternal, artemisIngress, appHost, _ = CVBuildHosts(p.CRName, p.Namespace, clusterDomain)
	appTLSHost = appHost
	if p.Ingress.Enabled && p.Ingress.HostOverride != "" {
		appTLSHost = p.Ingress.HostOverride
	}
	if p.Ingress.Enabled {
		artemisIngress = cvArtemisIngressHost(p, artemisIngress)
	}
	amqpInternal = CVAmqpInternalHost(p.Spec, p.CRName, p.Namespace)
	artemisManaged = p.Spec.Artemis.Enabled == nil || *p.Spec.Artemis.Enabled
	return
}

func amqpInternalForConsumerConfigMap(artemisManaged bool, artemisInternal, amqpInternal string) string {
	if artemisManaged {
		return artemisInternal
	}
	return amqpInternal
}

func cvEarlyExit(r ctrl.Result, err error) bool {
	return err != nil || r.RequeueAfter > 0 || r.Requeue
}

func reconcileCVArtemisOrExternalConfigMap(ctx context.Context, cfg CVPhaseConfig, mb string, artemisManaged bool, artemisInternal, artemisIngress, amqpInternal string) (ctrl.Result, error) {
	p := cfg.BuildParams
	if artemisManaged {
		return reconcileCVManagedArtemis(ctx, cfg, mb, artemisInternal, artemisIngress)
	}
	if err := cvReconcileConfigMap(ctx, cfg, BuildCVConfigMap(p, mb, amqpInternal)); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func ReconcileCV(ctx context.Context, cfg CVPhaseConfig) (ctrl.Result, error) {
	p := cfg.BuildParams
	mb := cfg.ManagedByValue
	_, artemisInternal, artemisIngress, appTLSHost, amqpInternal, artemisManaged := reconcileCVIngressAndNetworking(ctx, cfg)

	if err := reconcileCVRBAC(ctx, cfg, mb); err != nil {
		return ctrl.Result{}, err
	}
	if err := reconcileCVSecrets(ctx, cfg, mb, artemisManaged); err != nil {
		return ctrl.Result{}, err
	}
	if r, err := reconcileCVMariaDB(ctx, cfg, mb); cvEarlyExit(r, err) {
		return r, err
	}
	if r, err := reconcileCVArtemisOrExternalConfigMap(ctx, cfg, mb, artemisManaged, artemisInternal, artemisIngress, amqpInternal); cvEarlyExit(r, err) {
		return r, err
	}
	if r, err := reconcileCVAppCertificates(ctx, cfg, mb, appTLSHost, artemisManaged); cvEarlyExit(r, err) {
		return r, err
	}
	cmApp := BuildCVConfigMap(p, mb, amqpInternalForConsumerConfigMap(artemisManaged, artemisInternal, amqpInternal))
	configHash := commonreconciler.ComputeConfigHash(cmApp)
	if err := reconcileCVApp(ctx, cfg, mb, configHash); err != nil {
		return ctrl.Result{}, err
	}
	if cfg.ReconcileAppExposure != nil {
		if err := cfg.ReconcileAppExposure(ctx); err != nil {
			return ctrl.Result{}, err
		}
	}
	if r, err := checkCVReadiness(ctx, cfg, artemisManaged); cvEarlyExit(r, err) {
		return r, err
	}
	applyCVCondition(ctx, cfg, "Available", metav1.ConditionTrue, "Reconciled", "All resources created successfully")
	return ctrl.Result{}, nil
}

func reconcileCVRBAC(ctx context.Context, cfg CVPhaseConfig, managedBy string) error {
	p := cfg.BuildParams
	if err := cvReconcileServiceAccount(ctx, cfg, BuildCVServiceAccount(p, managedBy)); err != nil {
		return err
	}
	if err := cvReconcileRole(ctx, cfg, BuildCVRole(p, managedBy)); err != nil {
		return err
	}
	return cvReconcileRoleBinding(ctx, cfg, BuildCVRoleBinding(p, managedBy))
}

func reconcileCVSecrets(ctx context.Context, cfg CVPhaseConfig, managedBy string, artemisManaged bool) error {
	p := cfg.BuildParams
	if artemisManaged {
		if err := cvReconcileSecret(ctx, cfg, BuildCVArtemisCredentialsSecret(p, managedBy)); err != nil {
			return err
		}
		if err := cvReconcileSecret(ctx, cfg, BuildCVArtemisKeystoreSecret(p, managedBy)); err != nil {
			return err
		}
	}
	if err := cvReconcileSecret(ctx, cfg, BuildCVAmqpSecret(p, managedBy)); err != nil {
		return err
	}
	if p.Spec.MariaDB.ExistingSecret == "" {
		if err := cvReconcileSecret(ctx, cfg, BuildCVMariaDBSecret(p, managedBy)); err != nil {
			return err
		}
	}
	return nil
}

func reconcileCVMariaDB(ctx context.Context, cfg CVPhaseConfig, managedBy string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	p := cfg.BuildParams
	if err := cvReconcileService(ctx, cfg, BuildCVMariaDBService(p, managedBy)); err != nil {
		return ctrl.Result{}, err
	}
	if err := cvReconcileStatefulSet(ctx, cfg, BuildCVMariaDBStatefulSet(p, managedBy)); err != nil {
		return ctrl.Result{}, err
	}
	if !IsCVPodReady(ctx, cfg.Client, p.Namespace, map[string]string{"app": MariaDBServiceName(p.CRName)}) {
		logger.Info("MariaDB not ready yet, requeuing in 10 seconds")
		applyCVCondition(ctx, cfg, "MariaDBReady", metav1.ConditionFalse, "Provisioning", "Waiting for MariaDB to be ready")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	applyCVCondition(ctx, cfg, "MariaDBReady", metav1.ConditionTrue, "Ready", "MariaDB is ready")
	return ctrl.Result{}, nil
}

func reconcileCVManagedArtemis(ctx context.Context, cfg CVPhaseConfig, managedBy string, artemisInternalHost, artemisIngressHost string) (ctrl.Result, error) {
	p := cfg.BuildParams
	if err := cvReconcileConfigMap(ctx, cfg, BuildCVConfigMap(p, managedBy, artemisInternalHost)); err != nil {
		return ctrl.Result{}, err
	}
	if res, err := cvReconcileCertificate(ctx, cfg, BuildCVArtemisCertificate(p, managedBy, artemisIngressHost)); err != nil || res.Requeue || res.RequeueAfter > 0 {
		return res, err
	}
	if err := ReconcileCVArtemisSSLSecret(ctx, cfg.Client, cfg.Scheme, cfg.Owner, CVArtemisSSLReconcileParams{
		Spec:      p.Spec,
		CRName:    p.CRName,
		Namespace: p.Namespace,
		ManagedBy: managedBy,
	}); err != nil {
		return ctrl.Result{}, err
	}
	if err := cvReconcileUnstructured(ctx, cfg, BuildCVArtemisBroker(p, managedBy, artemisIngressHost)); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func reconcileCVAppCertificates(ctx context.Context, cfg CVPhaseConfig, managedBy string, appTLSHost string, artemisManaged bool) (ctrl.Result, error) {
	p := cfg.BuildParams
	if res, err := cvReconcileCertificate(ctx, cfg, BuildCVServerCertificate(p, managedBy, appTLSHost)); err != nil || res.Requeue || res.RequeueAfter > 0 {
		return res, err
	}
	if res, err := cvReconcileCertificate(ctx, cfg, BuildCVClientCertificate(p, managedBy)); err != nil || res.Requeue || res.RequeueAfter > 0 {
		return res, err
	}
	if artemisManaged {
		if res, err := cvReconcileCertificate(ctx, cfg, BuildCVMTLSCertificate(p, managedBy)); err != nil || res.Requeue || res.RequeueAfter > 0 {
			return res, err
		}
	}
	return ctrl.Result{}, nil
}

func reconcileCVApp(ctx context.Context, cfg CVPhaseConfig, managedBy string, configHash string) error {
	p := cfg.BuildParams
	if err := cvReconcileService(ctx, cfg, BuildCVAppService(p, managedBy)); err != nil {
		return err
	}
	if err := cvReconcileDeployment(ctx, cfg, BuildCVDeployment(p, managedBy, configHash)); err != nil {
		return err
	}
	return cvReconcileHPA(ctx, cfg, BuildCVHPA(p, managedBy))
}

func checkCVReadiness(ctx context.Context, cfg CVPhaseConfig, artemisManaged bool) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	p := cfg.BuildParams
	if artemisManaged {
		artemisName := fmt.Sprintf(constants.ArtemisSuffix, p.CRName)
		if !IsCVPodReady(ctx, cfg.Client, p.Namespace, map[string]string{"ActiveMQArtemis": artemisName}) {
			logger.Info("Artemis not ready yet, requeuing in 10 seconds")
			applyCVCondition(ctx, cfg, "ArtemisReady", metav1.ConditionFalse, "Provisioning", "Waiting for Artemis to be ready")
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}
		applyCVCondition(ctx, cfg, "ArtemisReady", metav1.ConditionTrue, "Ready", "Artemis is ready")
	} else {
		applyCVCondition(ctx, cfg, "ArtemisReady", metav1.ConditionTrue, "Skipped", "Artemis not managed by operator")
	}
	deployment := &appsv1.Deployment{}
	if err := cfg.Client.Get(ctx, client.ObjectKey{Name: p.CRName, Namespace: p.Namespace}, deployment); err != nil {
		logger.Info("App not ready yet, requeuing in 10 seconds")
		applyCVCondition(ctx, cfg, "AppReady", metav1.ConditionFalse, "Provisioning", "Waiting for app to be ready")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	appReady := deployment.Status.ReadyReplicas > 0 && deployment.Status.ReadyReplicas == deployment.Status.Replicas
	if !appReady {
		logger.Info("App not ready yet, requeuing in 10 seconds")
		applyCVCondition(ctx, cfg, "AppReady", metav1.ConditionFalse, "Provisioning", "Waiting for app to be ready")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	applyCVCondition(ctx, cfg, "AppReady", metav1.ConditionTrue, "Ready", "App is ready")
	return ctrl.Result{}, nil
}

func cvReconcileSecret(ctx context.Context, cfg CVPhaseConfig, desired *corev1.Secret) error {
	return commonreconciler.ReconcileSecret(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func cvReconcileConfigMap(ctx context.Context, cfg CVPhaseConfig, desired *corev1.ConfigMap) error {
	return commonreconciler.ReconcileConfigMap(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func cvReconcileService(ctx context.Context, cfg CVPhaseConfig, desired *corev1.Service) error {
	return commonreconciler.ReconcileService(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func cvReconcileDeployment(ctx context.Context, cfg CVPhaseConfig, desired *appsv1.Deployment) error {
	return commonreconciler.ReconcileDeployment(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func cvReconcileStatefulSet(ctx context.Context, cfg CVPhaseConfig, desired *appsv1.StatefulSet) error {
	return commonreconciler.ReconcileStatefulSet(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func cvReconcileCertificate(ctx context.Context, cfg CVPhaseConfig, desired *certmanagerv1.Certificate) (ctrl.Result, error) {
	return commonreconciler.ReconcileCertificate(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func cvReconcileUnstructured(ctx context.Context, cfg CVPhaseConfig, desired *unstructured.Unstructured) error {
	return commonreconciler.ReconcileUnstructured(ctx, cfg.Client, cfg.Owner, desired)
}

func cvReconcileHPA(ctx context.Context, cfg CVPhaseConfig, desired *autoscalingv2.HorizontalPodAutoscaler) error {
	return commonreconciler.ReconcileHPA(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func cvReconcileServiceAccount(ctx context.Context, cfg CVPhaseConfig, desired *corev1.ServiceAccount) error {
	return commonreconciler.ReconcileServiceAccount(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func cvReconcileRole(ctx context.Context, cfg CVPhaseConfig, desired *rbacv1.Role) error {
	return commonreconciler.ReconcileRole(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func cvReconcileRoleBinding(ctx context.Context, cfg CVPhaseConfig, desired *rbacv1.RoleBinding) error {
	return commonreconciler.ReconcileRoleBinding(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}
