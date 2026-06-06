package controller

import (
	"context"

	appsv1alpha1 "github.com/swim-developer/swim-openshift-operator/api/v1alpha1"
	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/controller/cv"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *SwimDnotamConsumerValidatorReconciler) dnotamCVPhaseConfig(cr *appsv1alpha1.SwimDnotamConsumerValidator, req ctrl.Request) cv.CVPhaseConfig {
	return cv.CVPhaseConfig{
		Client:         r.Client,
		Scheme:         r.Scheme,
		Owner:          cr,
		Request:        req,
		FinalizerName:  constants.ConsumerValidatorFinalizerName,
		CRKind:         "SwimDnotamConsumerValidator",
		BuildParams:    dnotamCVBuildParams(cr),
		ManagedByLabel: sharedManagedByLabel,
		ManagedByValue: sharedManagedByValue,
		ResolveClusterDomain: func(ctx context.Context, specDomain, namespace string) string {
			return getOrDetectClusterDomain(ctx, r.Client, specDomain, namespace)
		},
		RemoveFinalizer: resources.MakeRemoveFinalizerFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimDnotamConsumerValidator { return &appsv1alpha1.SwimDnotamConsumerValidator{} },
			constants.ConsumerValidatorFinalizerName,
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

func (r *SwimDnotamConsumerValidatorReconciler) reconcileSwimDnotamCV(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimDnotamConsumerValidator) (ctrl.Result, error) {
	clusterDomain := getOrDetectClusterDomain(ctx, r.Client, cr.Spec.Global.ClusterDomain, cr.Namespace)
	_, _, appHost, appHTTPHost := cv.CVBuildHosts(cr.Name, cr.Namespace, clusterDomain)
	cfg := r.dnotamCVPhaseConfig(cr, req)
	cfg.ReconcileAppExposure = func(ctx context.Context) error {
		if err := reconcileRouteResource(ctx, r.Client, r.Scheme, cr, buildCVRoute(cfg.BuildParams, cfg.ManagedByValue, appHost, "https")); err != nil {
			return err
		}
		return reconcileRouteResource(ctx, r.Client, r.Scheme, cr, buildCVRoute(cfg.BuildParams, cfg.ManagedByValue, appHTTPHost, "http"))
	}
	if result, err := cv.HandleCVFinalization(ctx, cfg); err != nil || result.Requeue || result.RequeueAfter > 0 || cr.DeletionTimestamp != nil {
		return result, err
	}
	if result, err := cv.EnsureCVFinalizer(ctx, cfg); err != nil || result.Requeue || result.RequeueAfter > 0 {
		return result, err
	}
	if err := ensureCRLabels(ctx, r.Client, cr, ConsumerValidatorApp); err != nil {
		return ctrl.Result{}, err
	}
	return cv.ReconcileCV(ctx, cfg)
}

func (r *SwimEd254ConsumerValidatorReconciler) ed254CVPhaseConfig(cr *appsv1alpha1.SwimEd254ConsumerValidator, req ctrl.Request) cv.CVPhaseConfig {
	return cv.CVPhaseConfig{
		Client:         r.Client,
		Scheme:         r.Scheme,
		Owner:          cr,
		Request:        req,
		FinalizerName:  constants.Ed254ConsumerValidatorFinalizerName,
		CRKind:         "SwimEd254ConsumerValidator",
		BuildParams:    ed254CVBuildParams(cr),
		ManagedByLabel: sharedManagedByLabel,
		ManagedByValue: sharedManagedByValue,
		ResolveClusterDomain: func(ctx context.Context, specDomain, namespace string) string {
			return getOrDetectClusterDomain(ctx, r.Client, specDomain, namespace)
		},
		RemoveFinalizer: func(ctx context.Context) error {
			latest := &appsv1alpha1.SwimEd254ConsumerValidator{}
			if err := r.Get(ctx, req.NamespacedName, latest); err != nil {
				return client.IgnoreNotFound(err)
			}
			if controllerutil.ContainsFinalizer(latest, constants.Ed254ConsumerValidatorFinalizerName) {
				controllerutil.RemoveFinalizer(latest, constants.Ed254ConsumerValidatorFinalizerName)
				return r.Update(ctx, latest)
			}
			return nil
		},
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

func (r *SwimFficeConsumerValidatorReconciler) fficeCVPhaseConfig(cr *appsv1alpha1.SwimFficeConsumerValidator, req ctrl.Request) cv.CVPhaseConfig {
	return cv.CVPhaseConfig{
		Client:         r.Client,
		Scheme:         r.Scheme,
		Owner:          cr,
		Request:        req,
		FinalizerName:  constants.FficeConsumerValidatorFinalizerName,
		CRKind:         "SwimFficeConsumerValidator",
		BuildParams:    fficeCVBuildParams(cr),
		ManagedByLabel: sharedManagedByLabel,
		ManagedByValue: sharedManagedByValue,
		ResolveClusterDomain: func(ctx context.Context, specDomain, namespace string) string {
			return getOrDetectClusterDomain(ctx, r.Client, specDomain, namespace)
		},
		RemoveFinalizer: func(ctx context.Context) error {
			latest := &appsv1alpha1.SwimFficeConsumerValidator{}
			if err := r.Get(ctx, req.NamespacedName, latest); err != nil {
				return client.IgnoreNotFound(err)
			}
			if controllerutil.ContainsFinalizer(latest, constants.FficeConsumerValidatorFinalizerName) {
				controllerutil.RemoveFinalizer(latest, constants.FficeConsumerValidatorFinalizerName)
				return r.Update(ctx, latest)
			}
			return nil
		},
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

func (r *SwimFficeConsumerValidatorReconciler) reconcileSwimFficeCV(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimFficeConsumerValidator) (ctrl.Result, error) {
	clusterDomain := getOrDetectClusterDomain(ctx, r.Client, cr.Spec.Global.ClusterDomain, cr.Namespace)
	_, _, appHost, appHTTPHost := cv.CVBuildHosts(cr.Name, cr.Namespace, clusterDomain)
	cfg := r.fficeCVPhaseConfig(cr, req)
	cfg.ReconcileAppExposure = func(ctx context.Context) error {
		if err := reconcileRouteResource(ctx, r.Client, r.Scheme, cr, buildCVRoute(cfg.BuildParams, cfg.ManagedByValue, appHost, "https")); err != nil {
			return err
		}
		return reconcileRouteResource(ctx, r.Client, r.Scheme, cr, buildCVRoute(cfg.BuildParams, cfg.ManagedByValue, appHTTPHost, "http"))
	}
	if result, err := cv.HandleCVFinalization(ctx, cfg); err != nil || result.Requeue || result.RequeueAfter > 0 || cr.DeletionTimestamp != nil {
		return result, err
	}
	if result, err := cv.EnsureCVFinalizer(ctx, cfg); err != nil || result.Requeue || result.RequeueAfter > 0 {
		return result, err
	}
	if err := ensureCRLabels(ctx, r.Client, cr, ConsumerValidatorApp); err != nil {
		return ctrl.Result{}, err
	}
	return cv.ReconcileCV(ctx, cfg)
}

func (r *SwimEd254ConsumerValidatorReconciler) reconcileSwimEd254CV(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimEd254ConsumerValidator) (ctrl.Result, error) {
	clusterDomain := getOrDetectClusterDomain(ctx, r.Client, cr.Spec.Global.ClusterDomain, cr.Namespace)
	_, _, appHost, appHTTPHost := cv.CVBuildHosts(cr.Name, cr.Namespace, clusterDomain)
	cfg := r.ed254CVPhaseConfig(cr, req)
	cfg.ReconcileAppExposure = func(ctx context.Context) error {
		if err := reconcileRouteResource(ctx, r.Client, r.Scheme, cr, buildCVRoute(cfg.BuildParams, cfg.ManagedByValue, appHost, "https")); err != nil {
			return err
		}
		return reconcileRouteResource(ctx, r.Client, r.Scheme, cr, buildCVRoute(cfg.BuildParams, cfg.ManagedByValue, appHTTPHost, "http"))
	}
	if result, err := cv.HandleCVFinalization(ctx, cfg); err != nil || result.Requeue || result.RequeueAfter > 0 || cr.DeletionTimestamp != nil {
		return result, err
	}
	if result, err := cv.EnsureCVFinalizer(ctx, cfg); err != nil || result.Requeue || result.RequeueAfter > 0 {
		return result, err
	}
	if err := ensureCRLabels(ctx, r.Client, cr, ConsumerValidatorApp); err != nil {
		return ctrl.Result{}, err
	}
	return cv.ReconcileCV(ctx, cfg)
}
