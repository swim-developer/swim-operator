package controller

import (
	"context"

	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	"github.com/swim-developer/swim-operator-common/pkg/controller/pv"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	appsv1alpha1 "github.com/swim-developer/swim-openshift-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func dnotamOCPPVBuildParams(cr *appsv1alpha1.SwimDnotamProviderValidator) pv.PVBuildParams {
	return pv.PVBuildParams{
		CRName:      cr.Name,
		Namespace:   cr.Namespace,
		CertManager: commonapi.CertManagerSpec{},
		Spec: commonapi.ProviderValidatorBaseSpec{
			Keycloak:        cr.Spec.Keycloak,
			ProviderAPIURLs: cr.Spec.ProviderAPIURLs,
			MariaDB:         cr.Spec.MariaDB,
			AMQP:            cr.Spec.AMQP,
			MTLS:            cr.Spec.MTLS,
			HPA:             cr.Spec.HPA,
		},
		DefaultImage:         "quay.io/masales/swim-dnotam-provider-validator:latest",
		IngressEnabled:       false,
		IngressHost:          "",
		IngressTLSSecretName: "",
		IngressAnnotations:   nil,
		RouteHost:            cr.Spec.Route.Host,
	}
}

func fficeOCPPVBuildParams(cr *appsv1alpha1.SwimFficeProviderValidator) pv.PVBuildParams {
	return pv.PVBuildParams{
		CRName:      cr.Name,
		Namespace:   cr.Namespace,
		CertManager: commonapi.CertManagerSpec{},
		Spec: commonapi.ProviderValidatorBaseSpec{
			Keycloak:        cr.Spec.Keycloak,
			ProviderAPIURLs: cr.Spec.ProviderAPIURLs,
			MariaDB:         cr.Spec.MariaDB,
			AMQP:            cr.Spec.AMQP,
			MTLS:            cr.Spec.MTLS,
			HPA:             cr.Spec.HPA,
		},
		DefaultImage:         "quay.io/masales/swim-ffice-provider-validator:latest",
		IngressEnabled:       false,
		IngressHost:          "",
		IngressTLSSecretName: "",
		IngressAnnotations:   nil,
		RouteHost:            cr.Spec.Route.Host,
	}
}

func (r *SwimFficeProviderValidatorReconciler) fficePVPhaseConfig(cr *appsv1alpha1.SwimFficeProviderValidator, req ctrl.Request) pv.PVPhaseConfig {
	return pv.PVPhaseConfig{
		Client:         r.Client,
		Scheme:         r.Scheme,
		Owner:          cr,
		Request:        req,
		BuildParams:    fficeOCPPVBuildParams(cr),
		ManagedByLabel: sharedManagedByLabel,
		ManagedByValue: sharedManagedByValue,
		ResolveClusterDomain: func(ctx context.Context, specDomain, namespace string) string {
			return getOrDetectClusterDomain(ctx, r.Client, specDomain, namespace)
		},
		ApplyStatus: resources.MakeApplyStatusFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimFficeProviderValidator { return &appsv1alpha1.SwimFficeProviderValidator{} },
			func(o *appsv1alpha1.SwimFficeProviderValidator) *[]metav1.Condition { return &o.Status.Conditions },
		),
		ReconcilePreAppTLS: nil,
	}
}

func ed254OCPPVBuildParams(cr *appsv1alpha1.SwimEd254ProviderValidator) pv.PVBuildParams {
	return pv.PVBuildParams{
		CRName:      cr.Name,
		Namespace:   cr.Namespace,
		CertManager: commonapi.CertManagerSpec{},
		Spec: commonapi.ProviderValidatorBaseSpec{
			Keycloak:        cr.Spec.Keycloak,
			ProviderAPIURLs: cr.Spec.ProviderAPIURLs,
			MariaDB:         cr.Spec.MariaDB,
			AMQP:            cr.Spec.AMQP,
			MTLS:            cr.Spec.MTLS,
			HPA:             cr.Spec.HPA,
		},
		DefaultImage:         "quay.io/masales/swim-ed254-provider-validator:latest",
		IngressEnabled:       false,
		IngressHost:          "",
		IngressTLSSecretName: "",
		IngressAnnotations:   nil,
		RouteHost:            cr.Spec.Route.Host,
	}
}

func (r *SwimEd254ProviderValidatorReconciler) ed254PVPhaseConfig(cr *appsv1alpha1.SwimEd254ProviderValidator, req ctrl.Request) pv.PVPhaseConfig {
	return pv.PVPhaseConfig{
		Client:         r.Client,
		Scheme:         r.Scheme,
		Owner:          cr,
		Request:        req,
		BuildParams:    ed254OCPPVBuildParams(cr),
		ManagedByLabel: sharedManagedByLabel,
		ManagedByValue: sharedManagedByValue,
		ResolveClusterDomain: func(ctx context.Context, specDomain, namespace string) string {
			return getOrDetectClusterDomain(ctx, r.Client, specDomain, namespace)
		},
		ApplyStatus: resources.MakeApplyStatusFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimEd254ProviderValidator { return &appsv1alpha1.SwimEd254ProviderValidator{} },
			func(o *appsv1alpha1.SwimEd254ProviderValidator) *[]metav1.Condition { return &o.Status.Conditions },
		),
		ReconcilePreAppTLS: nil,
	}
}

func (r *SwimDnotamProviderValidatorReconciler) pvPhaseConfig(cr *appsv1alpha1.SwimDnotamProviderValidator, req ctrl.Request) pv.PVPhaseConfig {
	return pv.PVPhaseConfig{
		Client:         r.Client,
		Scheme:         r.Scheme,
		Owner:          cr,
		Request:        req,
		BuildParams:    dnotamOCPPVBuildParams(cr),
		ManagedByLabel: sharedManagedByLabel,
		ManagedByValue: sharedManagedByValue,
		ResolveClusterDomain: func(ctx context.Context, specDomain, namespace string) string {
			return getOrDetectClusterDomain(ctx, r.Client, specDomain, namespace)
		},
		ApplyStatus: resources.MakeApplyStatusFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimDnotamProviderValidator { return &appsv1alpha1.SwimDnotamProviderValidator{} },
			func(o *appsv1alpha1.SwimDnotamProviderValidator) *[]metav1.Condition { return &o.Status.Conditions },
		),
		ReconcilePreAppTLS: nil,
	}
}
