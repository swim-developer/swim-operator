package controller

import (
	"context"

	appsv1alpha1 "github.com/swim-developer/swim-openshift-operator/api/v1alpha1"
	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/controller/provider"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *SwimDigitalNotamProviderReconciler) dnotamProviderPhaseConfig(req ctrl.Request, cr *appsv1alpha1.SwimDigitalNotamProvider) provider.ProviderPhaseConfig {
	return provider.ProviderPhaseConfig{
		Client:         r.Client,
		Scheme:         r.Scheme,
		Owner:          cr,
		Request:        req,
		FinalizerName:  constants.ProviderFinalizerName,
		CRKind:         "SwimDigitalNotamProvider",
		BuildParams:    SwimDigitalNotamProviderToBuildParams(cr),
		ManagedByLabel: sharedManagedByLabel,
		ManagedByValue: sharedManagedByValue,
		ResolveClusterDomain: func(ctx context.Context, specDomain, namespace string) string {
			return getOrDetectClusterDomain(ctx, r.Client, specDomain, namespace)
		},
		RemoveFinalizer: resources.MakeRemoveFinalizerFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimDigitalNotamProvider { return &appsv1alpha1.SwimDigitalNotamProvider{} },
			constants.ProviderFinalizerName,
		),
		ApplyStatus: resources.MakeApplyStatusFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimDigitalNotamProvider { return &appsv1alpha1.SwimDigitalNotamProvider{} },
			func(o *appsv1alpha1.SwimDigitalNotamProvider) *[]metav1.Condition { return &o.Status.Conditions },
		),
		ReconcileAppExposure: func(ctx context.Context, clusterDomain string) error {
			return reconcileProviderAppRoutes(ctx, r.Client, r.Scheme, cr, SwimDigitalNotamProviderToBuildParams(cr), clusterDomain, sharedManagedByValue)
		},
	}
}

func (r *SwimDigitalNotamProviderReconciler) handleProviderFinalization(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimDigitalNotamProvider) (ctrl.Result, error) {
	return provider.HandleProviderFinalization(ctx, r.dnotamProviderPhaseConfig(req, cr))
}

func (r *SwimDigitalNotamProviderReconciler) ensureProviderFinalizer(ctx context.Context, cr *appsv1alpha1.SwimDigitalNotamProvider) (ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.EnsureProviderFinalizer(ctx, r.dnotamProviderPhaseConfig(req, cr))
}

func (r *SwimDigitalNotamProviderReconciler) validateProviderExternalKafka(ctx context.Context, cr *appsv1alpha1.SwimDigitalNotamProvider) (ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ValidateProviderExternalKafkaPhase(ctx, r.dnotamProviderPhaseConfig(req, cr))
}

func (r *SwimDigitalNotamProviderReconciler) reconcileProviderRBAC(ctx context.Context, cr *appsv1alpha1.SwimDigitalNotamProvider) error {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ReconcileProviderRBACPhase(ctx, r.dnotamProviderPhaseConfig(req, cr))
}

func (r *SwimDigitalNotamProviderReconciler) reconcileProviderPostgres(ctx context.Context, cr *appsv1alpha1.SwimDigitalNotamProvider) (ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ReconcileProviderPostgresPhase(ctx, r.dnotamProviderPhaseConfig(req, cr))
}

func (r *SwimDigitalNotamProviderReconciler) buildProviderArtemisHost(cr *appsv1alpha1.SwimDigitalNotamProvider, clusterDomain string) string {
	return provider.ProviderArtemisIngressHost(SwimDigitalNotamProviderToBuildParams(cr), clusterDomain)
}

func (r *SwimDigitalNotamProviderReconciler) reconcileProviderArtemis(ctx context.Context, cr *appsv1alpha1.SwimDigitalNotamProvider, artemisIngressHost string) (ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ReconcileProviderArtemisPhase(ctx, r.dnotamProviderPhaseConfig(req, cr), artemisIngressHost)
}

func (r *SwimDigitalNotamProviderReconciler) reconcileProviderKafka(ctx context.Context, cr *appsv1alpha1.SwimDigitalNotamProvider) (ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ReconcileProviderKafkaPhase(ctx, r.dnotamProviderPhaseConfig(req, cr))
}

func (r *SwimDigitalNotamProviderReconciler) reconcileProviderAppConfig(ctx context.Context, cr *appsv1alpha1.SwimDigitalNotamProvider, clusterDomain string) (string, ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ReconcileProviderAppConfigPhase(ctx, r.dnotamProviderPhaseConfig(req, cr), clusterDomain)
}

func (r *SwimDigitalNotamProviderReconciler) reconcileProviderAppDeployment(ctx context.Context, cr *appsv1alpha1.SwimDigitalNotamProvider, clusterDomain, configHash string) error {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ReconcileProviderAppDeploymentPhase(ctx, r.dnotamProviderPhaseConfig(req, cr), clusterDomain, configHash)
}

func (r *SwimDigitalNotamProviderReconciler) checkProviderAppReadiness(ctx context.Context, cr *appsv1alpha1.SwimDigitalNotamProvider) (ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.CheckProviderAppReadinessPhase(ctx, r.dnotamProviderPhaseConfig(req, cr))
}

func (r *SwimDigitalNotamProviderReconciler) updateStatus(ctx context.Context, cr *appsv1alpha1.SwimDigitalNotamProvider, condType string, status metav1.ConditionStatus, reason, message string) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	_ = r.dnotamProviderPhaseConfig(req, cr).ApplyStatus(ctx, metav1.Condition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})
}

func (r *SwimEd254ProviderReconciler) ed254ProviderPhaseConfig(req ctrl.Request, cr *appsv1alpha1.SwimEd254Provider) provider.ProviderPhaseConfig {
	return provider.ProviderPhaseConfig{
		Client:         r.Client,
		Scheme:         r.Scheme,
		Owner:          cr,
		Request:        req,
		FinalizerName:  constants.Ed254ProviderFinalizerName,
		CRKind:         "SwimEd254Provider",
		BuildParams:    SwimEd254ProviderToBuildParams(cr),
		ManagedByLabel: sharedManagedByLabel,
		ManagedByValue: sharedManagedByValue,
		ResolveClusterDomain: func(ctx context.Context, specDomain, namespace string) string {
			return getOrDetectClusterDomain(ctx, r.Client, specDomain, namespace)
		},
		RemoveFinalizer: resources.MakeRemoveFinalizerFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimEd254Provider { return &appsv1alpha1.SwimEd254Provider{} },
			constants.Ed254ProviderFinalizerName,
		),
		ApplyStatus: resources.MakeApplyStatusFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimEd254Provider { return &appsv1alpha1.SwimEd254Provider{} },
			func(o *appsv1alpha1.SwimEd254Provider) *[]metav1.Condition { return &o.Status.Conditions },
		),
		ReconcileAppExposure: func(ctx context.Context, clusterDomain string) error {
			return reconcileProviderAppRoutes(ctx, r.Client, r.Scheme, cr, SwimEd254ProviderToBuildParams(cr), clusterDomain, sharedManagedByValue)
		},
	}
}

func (r *SwimEd254ProviderReconciler) handleEd254ProviderFinalization(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimEd254Provider) (ctrl.Result, error) {
	return provider.HandleProviderFinalization(ctx, r.ed254ProviderPhaseConfig(req, cr))
}

func (r *SwimEd254ProviderReconciler) ensureEd254ProviderFinalizer(ctx context.Context, cr *appsv1alpha1.SwimEd254Provider) (ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.EnsureProviderFinalizer(ctx, r.ed254ProviderPhaseConfig(req, cr))
}

func (r *SwimEd254ProviderReconciler) validateEd254ProviderExternalKafka(ctx context.Context, cr *appsv1alpha1.SwimEd254Provider) (ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ValidateProviderExternalKafkaPhase(ctx, r.ed254ProviderPhaseConfig(req, cr))
}

func (r *SwimEd254ProviderReconciler) reconcileEd254ProviderRBAC(ctx context.Context, cr *appsv1alpha1.SwimEd254Provider) error {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ReconcileProviderRBACPhase(ctx, r.ed254ProviderPhaseConfig(req, cr))
}

func (r *SwimEd254ProviderReconciler) reconcileEd254ProviderPostgres(ctx context.Context, cr *appsv1alpha1.SwimEd254Provider) (ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ReconcileProviderPostgresPhase(ctx, r.ed254ProviderPhaseConfig(req, cr))
}

func (r *SwimEd254ProviderReconciler) buildEd254ProviderArtemisHost(cr *appsv1alpha1.SwimEd254Provider, clusterDomain string) string {
	return provider.ProviderArtemisIngressHost(SwimEd254ProviderToBuildParams(cr), clusterDomain)
}

func (r *SwimEd254ProviderReconciler) reconcileEd254ProviderArtemis(ctx context.Context, cr *appsv1alpha1.SwimEd254Provider, artemisIngressHost string) (ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ReconcileProviderArtemisPhase(ctx, r.ed254ProviderPhaseConfig(req, cr), artemisIngressHost)
}

func (r *SwimEd254ProviderReconciler) reconcileEd254ProviderKafka(ctx context.Context, cr *appsv1alpha1.SwimEd254Provider) (ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ReconcileProviderKafkaPhase(ctx, r.ed254ProviderPhaseConfig(req, cr))
}

func (r *SwimEd254ProviderReconciler) reconcileEd254ProviderAppConfig(ctx context.Context, cr *appsv1alpha1.SwimEd254Provider, clusterDomain string) (string, ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ReconcileProviderAppConfigPhase(ctx, r.ed254ProviderPhaseConfig(req, cr), clusterDomain)
}

func (r *SwimEd254ProviderReconciler) reconcileEd254ProviderAppDeployment(ctx context.Context, cr *appsv1alpha1.SwimEd254Provider, clusterDomain, configHash string) error {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ReconcileProviderAppDeploymentPhase(ctx, r.ed254ProviderPhaseConfig(req, cr), clusterDomain, configHash)
}

func (r *SwimEd254ProviderReconciler) checkEd254ProviderAppReadiness(ctx context.Context, cr *appsv1alpha1.SwimEd254Provider) (ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.CheckProviderAppReadinessPhase(ctx, r.ed254ProviderPhaseConfig(req, cr))
}

func (r *SwimEd254ProviderReconciler) updateStatus(ctx context.Context, cr *appsv1alpha1.SwimEd254Provider, condType string, status metav1.ConditionStatus, reason, message string) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	_ = r.ed254ProviderPhaseConfig(req, cr).ApplyStatus(ctx, metav1.Condition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})
}

func (r *SwimFficeProviderReconciler) fficeProviderPhaseConfig(req ctrl.Request, cr *appsv1alpha1.SwimFficeProvider) provider.ProviderPhaseConfig {
	return provider.ProviderPhaseConfig{
		Client:         r.Client,
		Scheme:         r.Scheme,
		Owner:          cr,
		Request:        req,
		FinalizerName:  constants.FficeProviderFinalizerName,
		CRKind:         "SwimFficeProvider",
		BuildParams:    SwimFficeProviderToBuildParams(cr),
		ManagedByLabel: sharedManagedByLabel,
		ManagedByValue: sharedManagedByValue,
		ResolveClusterDomain: func(ctx context.Context, specDomain, namespace string) string {
			return getOrDetectClusterDomain(ctx, r.Client, specDomain, namespace)
		},
		RemoveFinalizer: resources.MakeRemoveFinalizerFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimFficeProvider { return &appsv1alpha1.SwimFficeProvider{} },
			constants.FficeProviderFinalizerName,
		),
		ApplyStatus: resources.MakeApplyStatusFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimFficeProvider { return &appsv1alpha1.SwimFficeProvider{} },
			func(o *appsv1alpha1.SwimFficeProvider) *[]metav1.Condition { return &o.Status.Conditions },
		),
		ReconcileAppExposure: func(ctx context.Context, clusterDomain string) error {
			return reconcileProviderAppRoutes(ctx, r.Client, r.Scheme, cr, SwimFficeProviderToBuildParams(cr), clusterDomain, sharedManagedByValue)
		},
	}
}

func (r *SwimFficeProviderReconciler) handleFficeProviderFinalization(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimFficeProvider) (ctrl.Result, error) {
	return provider.HandleProviderFinalization(ctx, r.fficeProviderPhaseConfig(req, cr))
}

func (r *SwimFficeProviderReconciler) ensureFficeProviderFinalizer(ctx context.Context, cr *appsv1alpha1.SwimFficeProvider) (ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.EnsureProviderFinalizer(ctx, r.fficeProviderPhaseConfig(req, cr))
}

func (r *SwimFficeProviderReconciler) validateFficeProviderExternalKafka(ctx context.Context, cr *appsv1alpha1.SwimFficeProvider) (ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ValidateProviderExternalKafkaPhase(ctx, r.fficeProviderPhaseConfig(req, cr))
}

func (r *SwimFficeProviderReconciler) reconcileFficeProviderRBAC(ctx context.Context, cr *appsv1alpha1.SwimFficeProvider) error {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ReconcileProviderRBACPhase(ctx, r.fficeProviderPhaseConfig(req, cr))
}

func (r *SwimFficeProviderReconciler) reconcileFficeProviderPostgres(ctx context.Context, cr *appsv1alpha1.SwimFficeProvider) (ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ReconcileProviderPostgresPhase(ctx, r.fficeProviderPhaseConfig(req, cr))
}

func (r *SwimFficeProviderReconciler) buildFficeProviderArtemisHost(cr *appsv1alpha1.SwimFficeProvider, clusterDomain string) string {
	return provider.ProviderArtemisIngressHost(SwimFficeProviderToBuildParams(cr), clusterDomain)
}

func (r *SwimFficeProviderReconciler) reconcileFficeProviderArtemis(ctx context.Context, cr *appsv1alpha1.SwimFficeProvider, artemisIngressHost string) (ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ReconcileProviderArtemisPhase(ctx, r.fficeProviderPhaseConfig(req, cr), artemisIngressHost)
}

func (r *SwimFficeProviderReconciler) reconcileFficeProviderKafka(ctx context.Context, cr *appsv1alpha1.SwimFficeProvider) (ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ReconcileProviderKafkaPhase(ctx, r.fficeProviderPhaseConfig(req, cr))
}

func (r *SwimFficeProviderReconciler) reconcileFficeProviderAppConfig(ctx context.Context, cr *appsv1alpha1.SwimFficeProvider, clusterDomain string) (string, ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ReconcileProviderAppConfigPhase(ctx, r.fficeProviderPhaseConfig(req, cr), clusterDomain)
}

func (r *SwimFficeProviderReconciler) reconcileFficeProviderAppDeployment(ctx context.Context, cr *appsv1alpha1.SwimFficeProvider, clusterDomain, configHash string) error {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.ReconcileProviderAppDeploymentPhase(ctx, r.fficeProviderPhaseConfig(req, cr), clusterDomain, configHash)
}

func (r *SwimFficeProviderReconciler) checkFficeProviderAppReadiness(ctx context.Context, cr *appsv1alpha1.SwimFficeProvider) (ctrl.Result, error) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	return provider.CheckProviderAppReadinessPhase(ctx, r.fficeProviderPhaseConfig(req, cr))
}

func (r *SwimFficeProviderReconciler) updateFficeProviderStatus(ctx context.Context, cr *appsv1alpha1.SwimFficeProvider, condType string, status metav1.ConditionStatus, reason, message string) {
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(cr)}
	_ = r.fficeProviderPhaseConfig(req, cr).ApplyStatus(ctx, metav1.Condition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})
}
