package controller

import (
	"context"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1alpha1 "github.com/swim-developer/swim-kubernetes-operator/api/v1alpha1"
	swimconst "github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/controller/cv"
	swimlabels "github.com/swim-developer/swim-operator-common/pkg/labels"
	commonreconciler "github.com/swim-developer/swim-operator-common/pkg/reconciler"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func buildFficeCVServiceMonitor(cr *appsv1alpha1.SwimFficeConsumerValidator) *monitoringv1.ServiceMonitor {
	lbl := swimlabels.StandardLabels(cr.Name, swimconst.ConsumerValidatorApp, cr.Name, sharedManagedByValue)
	return resources.BuildServiceMonitor(resources.ServiceMonitorParams{
		Name:        cr.Name,
		Namespace:   cr.Namespace,
		Labels:      lbl,
		MatchLabels: lbl,
		SelectorApp: cr.Name,
		PortName:    "http",
		MetricsPath: "/q/metrics",
	})
}

func buildDnotamCVServiceMonitor(cr *appsv1alpha1.SwimDnotamConsumerValidator) *monitoringv1.ServiceMonitor {
	lbl := swimlabels.StandardLabels(cr.Name, swimconst.ConsumerValidatorApp, cr.Name, sharedManagedByValue)
	return resources.BuildServiceMonitor(resources.ServiceMonitorParams{
		Name:        cr.Name,
		Namespace:   cr.Namespace,
		Labels:      lbl,
		MatchLabels: lbl,
		SelectorApp: cr.Name,
		PortName:    "http",
		MetricsPath: "/q/metrics",
	})
}

func buildEd254CVServiceMonitor(cr *appsv1alpha1.SwimEd254ConsumerValidator) *monitoringv1.ServiceMonitor {
	lbl := swimlabels.StandardLabels(cr.Name, swimconst.ConsumerValidatorApp, cr.Name, sharedManagedByValue)
	return resources.BuildServiceMonitor(resources.ServiceMonitorParams{
		Name:        cr.Name,
		Namespace:   cr.Namespace,
		Labels:      lbl,
		MatchLabels: lbl,
		SelectorApp: cr.Name,
		PortName:    "http",
		MetricsPath: "/q/metrics",
	})
}

func (r *SwimDnotamConsumerValidatorReconciler) dnotamCVPhaseConfig(cr *appsv1alpha1.SwimDnotamConsumerValidator, req ctrl.Request) cv.CVPhaseConfig {
	return cv.CVPhaseConfig{
		Client:         r.Client,
		Scheme:         r.Scheme,
		Owner:          cr,
		Request:        req,
		FinalizerName:  consumerValidatorFinalizerName,
		CRKind:         "SwimDnotamConsumerValidator",
		BuildParams:    dnotamK8sCVBuildParams(cr),
		ManagedByLabel: sharedManagedByLabel,
		ManagedByValue: sharedManagedByValue,
		ResolveClusterDomain: func(ctx context.Context, specDomain, _ string) string {
			return GetAppsDomain(specDomain)
		},
		RemoveFinalizer: resources.MakeRemoveFinalizerFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimDnotamConsumerValidator { return &appsv1alpha1.SwimDnotamConsumerValidator{} },
			consumerValidatorFinalizerName,
		),
		ApplyStatus: resources.MakeApplyStatusFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimDnotamConsumerValidator { return &appsv1alpha1.SwimDnotamConsumerValidator{} },
			func(o *appsv1alpha1.SwimDnotamConsumerValidator) *[]metav1.Condition { return &o.Status.Conditions },
		),
		FetchLatest: func(ctx context.Context) (client.Object, error) {
			o := &appsv1alpha1.SwimDnotamConsumerValidator{}
			err := r.Get(ctx, req.NamespacedName, o)
			return o, err
		},
	}
}

func (r *SwimDnotamConsumerValidatorReconciler) dnotamCVIngressExposure(ctx context.Context, cr *appsv1alpha1.SwimDnotamConsumerValidator, cfg cv.CVPhaseConfig, cvHost, clusterDomain string) error {
	if !cr.Spec.Ingress.Enabled {
		return nil
	}
	httpsHost := cvHost
	if cr.Spec.Ingress.Host != "" {
		httpsHost = cr.Spec.Ingress.Host
	}
	if err := commonreconciler.ReconcileIngress(ctx, r.Client, r.Scheme, cr, cv.BuildCVIngressHTTPS(cfg.BuildParams, cfg.ManagedByValue, httpsHost)); err != nil {
		return err
	}
	apiHost := cv.CVAPIIngressHost(cfg.BuildParams, clusterDomain)
	return commonreconciler.ReconcileIngress(ctx, r.Client, r.Scheme, cr, cv.BuildCVIngressAPI(cfg.BuildParams, cfg.ManagedByValue, apiHost))
}

func (r *SwimDnotamConsumerValidatorReconciler) maybeReconcileDnotamCVServiceMonitor(ctx context.Context, cr *appsv1alpha1.SwimDnotamConsumerValidator) {
	if !cr.Spec.Observability.ServiceMonitorEnabled {
		return
	}
	if err := commonreconciler.ReconcileServiceMonitor(ctx, r.Client, r.Scheme, cr, buildDnotamCVServiceMonitor(cr)); err != nil {
		log.FromContext(ctx).V(1).Info("ServiceMonitor reconcile failed", "error", err)
	}
}

func (r *SwimDnotamConsumerValidatorReconciler) reconcileSwimDnotamCV(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimDnotamConsumerValidator) (ctrl.Result, error) {
	clusterDomain := GetAppsDomain(cr.Spec.Global.ClusterDomain)
	_, _, cvHost, _ := cv.CVBuildHosts(cr.Name, cr.Namespace, clusterDomain)
	cfg := r.dnotamCVPhaseConfig(cr, req)
	cfg.ReconcileAppExposure = func(ctx context.Context) error {
		return r.dnotamCVIngressExposure(ctx, cr, cfg, cvHost, clusterDomain)
	}
	if result, err := cv.HandleCVFinalization(ctx, cfg); err != nil {
		return result, err
	} else if result.Requeue || result.RequeueAfter > 0 || cr.DeletionTimestamp != nil {
		return result, nil
	}
	if result, err := cv.EnsureCVFinalizer(ctx, cfg); err != nil {
		return result, err
	} else if result.Requeue || result.RequeueAfter > 0 {
		return result, nil
	}
	if err := swimlabels.EnsureCRLabels(ctx, r.Client, cr, swimconst.ConsumerValidatorApp, sharedManagedByValue); err != nil {
		return ctrl.Result{}, err
	}
	result, err := cv.ReconcileCV(ctx, cfg)
	if err != nil {
		return result, err
	}
	if result.Requeue || result.RequeueAfter > 0 {
		return result, nil
	}
	r.maybeReconcileDnotamCVServiceMonitor(ctx, cr)
	return ctrl.Result{}, nil
}

func (r *SwimEd254ConsumerValidatorReconciler) ed254CVPhaseConfig(cr *appsv1alpha1.SwimEd254ConsumerValidator, req ctrl.Request) cv.CVPhaseConfig {
	return cv.CVPhaseConfig{
		Client:         r.Client,
		Scheme:         r.Scheme,
		Owner:          cr,
		Request:        req,
		FinalizerName:  ed254ConsumerValidatorFinalizerName,
		CRKind:         "SwimEd254ConsumerValidator",
		BuildParams:    ed254K8sCVBuildParams(cr),
		ManagedByLabel: sharedManagedByLabel,
		ManagedByValue: sharedManagedByValue,
		ResolveClusterDomain: func(ctx context.Context, specDomain, namespace string) string {
			return getOrDetectClusterDomain(ctx, r.Client, specDomain, namespace)
		},
		RemoveFinalizer: resources.MakeRemoveFinalizerFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimEd254ConsumerValidator { return &appsv1alpha1.SwimEd254ConsumerValidator{} },
			ed254ConsumerValidatorFinalizerName,
		),
		ApplyStatus: resources.MakeApplyStatusFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimEd254ConsumerValidator { return &appsv1alpha1.SwimEd254ConsumerValidator{} },
			func(o *appsv1alpha1.SwimEd254ConsumerValidator) *[]metav1.Condition { return &o.Status.Conditions },
		),
		FetchLatest: func(ctx context.Context) (client.Object, error) {
			o := &appsv1alpha1.SwimEd254ConsumerValidator{}
			err := r.Get(ctx, req.NamespacedName, o)
			return o, err
		},
	}
}

func (r *SwimEd254ConsumerValidatorReconciler) ed254CVIngressExposure(ctx context.Context, cr *appsv1alpha1.SwimEd254ConsumerValidator, cfg cv.CVPhaseConfig, cvHost, clusterDomain string) error {
	if !cr.Spec.Ingress.Enabled {
		return nil
	}
	httpsHost := cvHost
	if cr.Spec.Ingress.Host != "" {
		httpsHost = cr.Spec.Ingress.Host
	}
	if err := commonreconciler.ReconcileIngress(ctx, r.Client, r.Scheme, cr, cv.BuildCVIngressHTTPS(cfg.BuildParams, cfg.ManagedByValue, httpsHost)); err != nil {
		return err
	}
	apiHost := cv.CVAPIIngressHost(cfg.BuildParams, clusterDomain)
	return commonreconciler.ReconcileIngress(ctx, r.Client, r.Scheme, cr, cv.BuildCVIngressAPI(cfg.BuildParams, cfg.ManagedByValue, apiHost))
}

func (r *SwimEd254ConsumerValidatorReconciler) maybeReconcileEd254CVServiceMonitor(ctx context.Context, cr *appsv1alpha1.SwimEd254ConsumerValidator) {
	if !cr.Spec.Observability.ServiceMonitorEnabled {
		return
	}
	if err := commonreconciler.ReconcileServiceMonitor(ctx, r.Client, r.Scheme, cr, buildEd254CVServiceMonitor(cr)); err != nil {
		log.FromContext(ctx).V(1).Info("ServiceMonitor reconcile failed", "error", err)
	}
}

func (r *SwimFficeConsumerValidatorReconciler) fficeCVPhaseConfig(cr *appsv1alpha1.SwimFficeConsumerValidator, req ctrl.Request) cv.CVPhaseConfig {
	return cv.CVPhaseConfig{
		Client:         r.Client,
		Scheme:         r.Scheme,
		Owner:          cr,
		Request:        req,
		FinalizerName:  fficeConsumerValidatorFinalizerName,
		CRKind:         "SwimFficeConsumerValidator",
		BuildParams:    fficeK8sCVBuildParams(cr),
		ManagedByLabel: sharedManagedByLabel,
		ManagedByValue: sharedManagedByValue,
		ResolveClusterDomain: func(ctx context.Context, specDomain, namespace string) string {
			return getOrDetectClusterDomain(ctx, r.Client, specDomain, namespace)
		},
		RemoveFinalizer: resources.MakeRemoveFinalizerFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimFficeConsumerValidator { return &appsv1alpha1.SwimFficeConsumerValidator{} },
			fficeConsumerValidatorFinalizerName,
		),
		ApplyStatus: resources.MakeApplyStatusFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimFficeConsumerValidator { return &appsv1alpha1.SwimFficeConsumerValidator{} },
			func(o *appsv1alpha1.SwimFficeConsumerValidator) *[]metav1.Condition { return &o.Status.Conditions },
		),
		FetchLatest: func(ctx context.Context) (client.Object, error) {
			o := &appsv1alpha1.SwimFficeConsumerValidator{}
			err := r.Get(ctx, req.NamespacedName, o)
			return o, err
		},
	}
}

func (r *SwimFficeConsumerValidatorReconciler) fficeCVIngressExposure(ctx context.Context, cr *appsv1alpha1.SwimFficeConsumerValidator, cfg cv.CVPhaseConfig, cvHost, clusterDomain string) error {
	if !cr.Spec.Ingress.Enabled {
		return nil
	}
	httpsHost := cvHost
	if cr.Spec.Ingress.Host != "" {
		httpsHost = cr.Spec.Ingress.Host
	}
	if err := commonreconciler.ReconcileIngress(ctx, r.Client, r.Scheme, cr, cv.BuildCVIngressHTTPS(cfg.BuildParams, cfg.ManagedByValue, httpsHost)); err != nil {
		return err
	}
	apiHost := cv.CVAPIIngressHost(cfg.BuildParams, clusterDomain)
	return commonreconciler.ReconcileIngress(ctx, r.Client, r.Scheme, cr, cv.BuildCVIngressAPI(cfg.BuildParams, cfg.ManagedByValue, apiHost))
}

func (r *SwimFficeConsumerValidatorReconciler) maybeReconcileFficeCVServiceMonitor(ctx context.Context, cr *appsv1alpha1.SwimFficeConsumerValidator) {
	if !cr.Spec.Observability.ServiceMonitorEnabled {
		return
	}
	if err := commonreconciler.ReconcileServiceMonitor(ctx, r.Client, r.Scheme, cr, buildFficeCVServiceMonitor(cr)); err != nil {
		log.FromContext(ctx).V(1).Info("ServiceMonitor reconcile failed", "error", err)
	}
}

func (r *SwimFficeConsumerValidatorReconciler) reconcileSwimFficeCV(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimFficeConsumerValidator) (ctrl.Result, error) {
	clusterDomain := getOrDetectClusterDomain(ctx, r.Client, cr.Spec.Global.ClusterDomain, cr.Namespace)
	_, _, cvHost, _ := cv.CVBuildHosts(cr.Name, cr.Namespace, clusterDomain)
	cfg := r.fficeCVPhaseConfig(cr, req)
	cfg.ReconcileAppExposure = func(ctx context.Context) error {
		return r.fficeCVIngressExposure(ctx, cr, cfg, cvHost, clusterDomain)
	}
	if result, err := cv.HandleCVFinalization(ctx, cfg); err != nil {
		return result, err
	} else if result.Requeue || result.RequeueAfter > 0 || cr.DeletionTimestamp != nil {
		return result, nil
	}
	if result, err := cv.EnsureCVFinalizer(ctx, cfg); err != nil {
		return result, err
	} else if result.Requeue || result.RequeueAfter > 0 {
		return result, nil
	}
	if err := swimlabels.EnsureCRLabels(ctx, r.Client, cr, swimconst.ConsumerValidatorApp, sharedManagedByValue); err != nil {
		return ctrl.Result{}, err
	}
	result, err := cv.ReconcileCV(ctx, cfg)
	if err != nil {
		return result, err
	}
	if result.Requeue || result.RequeueAfter > 0 {
		return result, nil
	}
	r.maybeReconcileFficeCVServiceMonitor(ctx, cr)
	return ctrl.Result{}, nil
}

func (r *SwimEd254ConsumerValidatorReconciler) reconcileSwimEd254CV(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimEd254ConsumerValidator) (ctrl.Result, error) {
	clusterDomain := getOrDetectClusterDomain(ctx, r.Client, cr.Spec.Global.ClusterDomain, cr.Namespace)
	_, _, cvHost, _ := cv.CVBuildHosts(cr.Name, cr.Namespace, clusterDomain)
	cfg := r.ed254CVPhaseConfig(cr, req)
	cfg.ReconcileAppExposure = func(ctx context.Context) error {
		return r.ed254CVIngressExposure(ctx, cr, cfg, cvHost, clusterDomain)
	}
	if result, err := cv.HandleCVFinalization(ctx, cfg); err != nil {
		return result, err
	} else if result.Requeue || result.RequeueAfter > 0 || cr.DeletionTimestamp != nil {
		return result, nil
	}
	if result, err := cv.EnsureCVFinalizer(ctx, cfg); err != nil {
		return result, err
	} else if result.Requeue || result.RequeueAfter > 0 {
		return result, nil
	}
	if err := swimlabels.EnsureCRLabels(ctx, r.Client, cr, swimconst.ConsumerValidatorApp, sharedManagedByValue); err != nil {
		return ctrl.Result{}, err
	}
	result, err := cv.ReconcileCV(ctx, cfg)
	if err != nil {
		return result, err
	}
	if result.Requeue || result.RequeueAfter > 0 {
		return result, nil
	}
	r.maybeReconcileEd254CVServiceMonitor(ctx, cr)
	return ctrl.Result{}, nil
}
