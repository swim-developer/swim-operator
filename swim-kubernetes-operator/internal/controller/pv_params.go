package controller

import (
	"context"

	appsv1alpha1 "github.com/swim-developer/swim-kubernetes-operator/api/v1alpha1"
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	"github.com/swim-developer/swim-operator-common/pkg/controller/pv"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func dnotamK8sPVBuildParams(cr *appsv1alpha1.SwimDnotamProviderValidator) pv.PVBuildParams {
	return pv.PVBuildParams{
		CRName:    cr.Name,
		Namespace: cr.Namespace,
		CertManager: commonapi.CertManagerSpec{
			Enabled:    cr.Spec.CertManager.Enabled,
			IssuerName: cr.Spec.CertManager.IssuerName,
			IssuerKind: cr.Spec.CertManager.IssuerKind,
		},
		Spec: commonapi.ProviderValidatorBaseSpec{
			Keycloak:        cr.Spec.Keycloak,
			ProviderAPIURLs: cr.Spec.ProviderAPIURLs,
			MariaDB:         cr.Spec.MariaDB,
			AMQP:            cr.Spec.AMQP,
			MTLS:            cr.Spec.MTLS,
			HPA:             cr.Spec.HPA,
		},
		DefaultImage:         "quay.io/masales/swim-dnotam-provider-validator:latest",
		IngressEnabled:       cr.Spec.Ingress.Enabled,
		IngressHost:          cr.Spec.Ingress.Host,
		IngressTLSSecretName: cr.Spec.Ingress.TLSSecretName,
		IngressAnnotations:   cr.Spec.Ingress.Annotations,
	}
}

func fficeK8sPVBuildParams(cr *appsv1alpha1.SwimFficeProviderValidator) pv.PVBuildParams {
	return pv.PVBuildParams{
		CRName:    cr.Name,
		Namespace: cr.Namespace,
		CertManager: commonapi.CertManagerSpec{
			Enabled:    cr.Spec.CertManager.Enabled,
			IssuerName: cr.Spec.CertManager.IssuerName,
			IssuerKind: cr.Spec.CertManager.IssuerKind,
		},
		Spec: commonapi.ProviderValidatorBaseSpec{
			Keycloak:        cr.Spec.Keycloak,
			ProviderAPIURLs: cr.Spec.ProviderAPIURLs,
			MariaDB:         cr.Spec.MariaDB,
			AMQP:            cr.Spec.AMQP,
			MTLS:            cr.Spec.MTLS,
			HPA:             cr.Spec.HPA,
		},
		DefaultImage:         "quay.io/masales/swim-ffice-provider-validator:latest",
		IngressEnabled:       cr.Spec.Ingress.Enabled,
		IngressHost:          cr.Spec.Ingress.Host,
		IngressTLSSecretName: cr.Spec.Ingress.TLSSecretName,
		IngressAnnotations:   cr.Spec.Ingress.Annotations,
	}
}

func (r *SwimFficeProviderValidatorReconciler) fficePVPhaseConfig(cr *appsv1alpha1.SwimFficeProviderValidator, req ctrl.Request) pv.PVPhaseConfig {
	return pv.PVPhaseConfig{
		Client:         r.Client,
		Scheme:         r.Scheme,
		Owner:          cr,
		Request:        req,
		BuildParams:    fficeK8sPVBuildParams(cr),
		ManagedByLabel: sharedManagedByLabel,
		ManagedByValue: sharedManagedByValue,
		ResolveClusterDomain: func(ctx context.Context, specDomain, _ string) string {
			return GetAppsDomain(specDomain)
		},
		ApplyStatus: resources.MakeApplyStatusFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimFficeProviderValidator { return &appsv1alpha1.SwimFficeProviderValidator{} },
			func(o *appsv1alpha1.SwimFficeProviderValidator) *[]metav1.Condition { return &o.Status.Conditions },
		),
	}
}

func ed254K8sPVBuildParams(cr *appsv1alpha1.SwimEd254ProviderValidator) pv.PVBuildParams {
	return pv.PVBuildParams{
		CRName:    cr.Name,
		Namespace: cr.Namespace,
		CertManager: commonapi.CertManagerSpec{
			Enabled:    cr.Spec.CertManager.Enabled,
			IssuerName: cr.Spec.CertManager.IssuerName,
			IssuerKind: cr.Spec.CertManager.IssuerKind,
		},
		Spec: commonapi.ProviderValidatorBaseSpec{
			Keycloak:        cr.Spec.Keycloak,
			ProviderAPIURLs: cr.Spec.ProviderAPIURLs,
			MariaDB:         cr.Spec.MariaDB,
			AMQP:            cr.Spec.AMQP,
			MTLS:            cr.Spec.MTLS,
			HPA:             cr.Spec.HPA,
		},
		DefaultImage:         "quay.io/masales/swim-ed254-provider-validator:latest",
		IngressEnabled:       cr.Spec.Ingress.Enabled,
		IngressHost:          cr.Spec.Ingress.Host,
		IngressTLSSecretName: cr.Spec.Ingress.TLSSecretName,
		IngressAnnotations:   cr.Spec.Ingress.Annotations,
	}
}

func (r *SwimEd254ProviderValidatorReconciler) ed254PVPhaseConfig(cr *appsv1alpha1.SwimEd254ProviderValidator, req ctrl.Request) pv.PVPhaseConfig {
	return pv.PVPhaseConfig{
		Client:         r.Client,
		Scheme:         r.Scheme,
		Owner:          cr,
		Request:        req,
		BuildParams:    ed254K8sPVBuildParams(cr),
		ManagedByLabel: sharedManagedByLabel,
		ManagedByValue: sharedManagedByValue,
		ResolveClusterDomain: func(ctx context.Context, specDomain, _ string) string {
			return GetAppsDomain(specDomain)
		},
		ApplyStatus: resources.MakeApplyStatusFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimEd254ProviderValidator { return &appsv1alpha1.SwimEd254ProviderValidator{} },
			func(o *appsv1alpha1.SwimEd254ProviderValidator) *[]metav1.Condition { return &o.Status.Conditions },
		),
	}
}

func (r *SwimDnotamProviderValidatorReconciler) pvPhaseConfig(cr *appsv1alpha1.SwimDnotamProviderValidator, req ctrl.Request) pv.PVPhaseConfig {
	return pv.PVPhaseConfig{
		Client:         r.Client,
		Scheme:         r.Scheme,
		Owner:          cr,
		Request:        req,
		BuildParams:    dnotamK8sPVBuildParams(cr),
		ManagedByLabel: sharedManagedByLabel,
		ManagedByValue: sharedManagedByValue,
		ResolveClusterDomain: func(ctx context.Context, specDomain, _ string) string {
			return GetAppsDomain(specDomain)
		},
		ApplyStatus: resources.MakeApplyStatusFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimDnotamProviderValidator { return &appsv1alpha1.SwimDnotamProviderValidator{} },
			func(o *appsv1alpha1.SwimDnotamProviderValidator) *[]metav1.Condition { return &o.Status.Conditions },
		),
	}
}
