package pv

import (
	"context"
	"time"

	"github.com/swim-developer/swim-operator-common/pkg/helpers"
	commonreconciler "github.com/swim-developer/swim-operator-common/pkg/reconciler"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func applyPVCondition(ctx context.Context, cfg PVPhaseConfig, conditionType string, status metav1.ConditionStatus, reason, message string) {
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

func ReconcilePV(ctx context.Context, cfg PVPhaseConfig) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	p := cfg.BuildParams
	mb := cfg.ManagedByValue
	if err := reconcilePVRBAC(ctx, cfg, mb); err != nil {
		return ctrl.Result{}, err
	}
	if r, err := reconcilePVMariaDB(ctx, cfg, mb); err != nil || r.RequeueAfter > 0 {
		return r, err
	}
	if cfg.ReconcilePreAppTLS != nil {
		if r, err := cfg.ReconcilePreAppTLS(ctx); err != nil || r.RequeueAfter > 0 {
			return r, err
		}
	}
	if r, err := reconcilePVApplication(ctx, cfg, mb); err != nil || r.RequeueAfter > 0 {
		return r, err
	}
	if cfg.ReconcileAppExposure != nil {
		if err := cfg.ReconcileAppExposure(ctx); err != nil {
			return ctrl.Result{}, err
		}
	}
	if r, err := checkPVAppReadiness(ctx, cfg); err != nil || r.RequeueAfter > 0 {
		return r, err
	}
	applyPVCondition(ctx, cfg, "Available", metav1.ConditionTrue, "Reconciled", "All resources created successfully")
	logger.Info("Provider validator reconcile complete", "name", p.CRName, "namespace", p.Namespace)
	return ctrl.Result{}, nil
}

func reconcilePVRBAC(ctx context.Context, cfg PVPhaseConfig, managedBy string) error {
	p := cfg.BuildParams
	if err := pvReconcileServiceAccount(ctx, cfg, BuildPVServiceAccount(p, managedBy)); err != nil {
		return err
	}
	if err := pvReconcileRole(ctx, cfg, BuildPVRole(p, managedBy)); err != nil {
		return err
	}
	return pvReconcileRoleBinding(ctx, cfg, BuildPVRoleBinding(p, managedBy))
}

func reconcilePVMariaDB(ctx context.Context, cfg PVPhaseConfig, managedBy string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	p := cfg.BuildParams
	if err := pvReconcilePVC(ctx, cfg, BuildPVAppPVC(p, managedBy)); err != nil {
		return ctrl.Result{}, err
	}
	if p.Spec.MariaDB.ExistingSecret == "" {
		if err := pvReconcileSecret(ctx, cfg, BuildPVMariaDBSecret(p, managedBy)); err != nil {
			return ctrl.Result{}, err
		}
	}
	if err := pvReconcileService(ctx, cfg, BuildPVMariaDBService(p, managedBy)); err != nil {
		return ctrl.Result{}, err
	}
	if err := pvReconcileStatefulSet(ctx, cfg, BuildPVMariaDBStatefulSet(p, managedBy)); err != nil {
		return ctrl.Result{}, err
	}
	if !isPVMariaDBReady(ctx, cfg.Client, p.Namespace, p.CRName) {
		logger.Info("MariaDB not ready yet, requeuing in 10 seconds")
		applyPVCondition(ctx, cfg, "MariaDBReady", metav1.ConditionFalse, "Provisioning", "Waiting for MariaDB to be ready")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	applyPVCondition(ctx, cfg, "MariaDBReady", metav1.ConditionTrue, "Ready", "MariaDB is ready")
	return ctrl.Result{}, nil
}

func reconcilePVApplication(ctx context.Context, cfg PVPhaseConfig, managedBy string) (ctrl.Result, error) {
	p := cfg.BuildParams
	desiredCM := BuildPVConfigMap(p, managedBy)
	if err := pvReconcileConfigMap(ctx, cfg, desiredCM); err != nil {
		return ctrl.Result{}, err
	}
	configHash := commonreconciler.ComputeConfigHash(desiredCM)
	if err := pvReconcileDeployment(ctx, cfg, BuildPVDeployment(p, managedBy, configHash)); err != nil {
		return ctrl.Result{}, err
	}
	if err := pvReconcileService(ctx, cfg, BuildPVService(p, managedBy)); err != nil {
		return ctrl.Result{}, err
	}
	if err := pvReconcileHPA(ctx, cfg, BuildPVHPA(p, managedBy)); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func checkPVAppReadiness(ctx context.Context, cfg PVPhaseConfig) (ctrl.Result, error) {
	p := cfg.BuildParams
	if !isPVAppDeploymentReady(ctx, cfg.Client, p.Namespace, p.CRName) {
		applyPVCondition(ctx, cfg, "ProviderValidatorReady", metav1.ConditionFalse, "Provisioning", "Waiting for ProviderValidator app to be ready")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	applyPVCondition(ctx, cfg, "ProviderValidatorReady", metav1.ConditionTrue, "Ready", "ProviderValidator is ready")
	return ctrl.Result{}, nil
}

func isPVMariaDBReady(ctx context.Context, c client.Client, namespace, crName string) bool {
	sts := &appsv1.StatefulSet{}
	key := client.ObjectKey{Name: MariaDBServiceName(crName), Namespace: namespace}
	if err := c.Get(ctx, key, sts); err != nil {
		return false
	}
	if sts.Spec.Replicas == nil {
		return false
	}
	return sts.Status.ReadyReplicas > 0 && sts.Status.ReadyReplicas == *sts.Spec.Replicas
}

func isPVAppDeploymentReady(ctx context.Context, c client.Client, namespace, crName string) bool {
	deployment := &appsv1.Deployment{}
	if err := c.Get(ctx, client.ObjectKey{Name: crName, Namespace: namespace}, deployment); err != nil {
		return false
	}
	if deployment.Spec.Replicas == nil {
		return false
	}
	return deployment.Status.ReadyReplicas > 0 && deployment.Status.ReadyReplicas == *deployment.Spec.Replicas
}

func IsPVPodReady(ctx context.Context, c client.Client, namespace string, lbls map[string]string) bool {
	return helpers.IsPodReady(ctx, c, namespace, lbls)
}

func pvReconcileSecret(ctx context.Context, cfg PVPhaseConfig, desired *corev1.Secret) error {
	return commonreconciler.ReconcileSecret(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func pvReconcileConfigMap(ctx context.Context, cfg PVPhaseConfig, desired *corev1.ConfigMap) error {
	return commonreconciler.ReconcileConfigMap(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func pvReconcileService(ctx context.Context, cfg PVPhaseConfig, desired *corev1.Service) error {
	return commonreconciler.ReconcileService(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func pvReconcileDeployment(ctx context.Context, cfg PVPhaseConfig, desired *appsv1.Deployment) error {
	return commonreconciler.ReconcileDeployment(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func pvReconcileHPA(ctx context.Context, cfg PVPhaseConfig, desired *autoscalingv2.HorizontalPodAutoscaler) error {
	return commonreconciler.ReconcileHPA(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func pvReconcilePVC(ctx context.Context, cfg PVPhaseConfig, desired *corev1.PersistentVolumeClaim) error {
	return commonreconciler.ReconcilePVC(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func pvReconcileServiceAccount(ctx context.Context, cfg PVPhaseConfig, desired *corev1.ServiceAccount) error {
	return commonreconciler.ReconcileServiceAccount(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func pvReconcileRole(ctx context.Context, cfg PVPhaseConfig, desired *rbacv1.Role) error {
	return commonreconciler.ReconcileRole(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func pvReconcileRoleBinding(ctx context.Context, cfg PVPhaseConfig, desired *rbacv1.RoleBinding) error {
	return commonreconciler.ReconcileRoleBinding(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}

func pvReconcileStatefulSet(ctx context.Context, cfg PVPhaseConfig, desired *appsv1.StatefulSet) error {
	return commonreconciler.ReconcileStatefulSet(ctx, cfg.Client, cfg.Scheme, cfg.Owner, desired)
}
