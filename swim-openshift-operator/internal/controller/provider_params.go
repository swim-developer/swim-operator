package controller

import (
	appsv1alpha1 "github.com/swim-developer/swim-openshift-operator/api/v1alpha1"
	"github.com/swim-developer/swim-operator-common/pkg/controller/provider"
)

func providerExposureFromRoutes(r appsv1alpha1.ProviderRoutesSpec) provider.ProviderExposureSpec {
	return provider.ProviderExposureSpec{
		HTTPEdgeEnabled:         r.HTTPSEdge,
		HTTPSPassthroughEnabled: r.HTTPSPassthrough,
		HTTPSEdgeHost:           r.HTTPSEdgeHost,
		HTTPSPassthroughHost:    r.HTTPSPassthroughHost,
	}
}

func SwimDigitalNotamProviderToBuildParams(cr *appsv1alpha1.SwimDigitalNotamProvider) provider.ProviderBuildParams {
	return provider.ProviderBuildParams{
		Name:                cr.Name,
		Namespace:           cr.Namespace,
		GlobalClusterDomain: cr.Spec.Global.ClusterDomain,
		Kafka:               cr.Spec.Kafka,
		CertManager:         cr.Spec.CertManager,
		ArtemisExposeMode:   "route",
		HPA:                 cr.Spec.HPA,
		Strategy: provider.DnotamProviderStrategy{
			Payload: provider.DnotamProviderPayload{
				Postgres: cr.Spec.Postgres,
				Artemis:  cr.Spec.Artemis,
				Provider: cr.Spec.Provider.ProviderAppBaseSpec,
				Exposure: providerExposureFromRoutes(cr.Spec.Provider.Routes),
			},
		},
	}
}

func SwimEd254ProviderToBuildParams(cr *appsv1alpha1.SwimEd254Provider) provider.ProviderBuildParams {
	return provider.ProviderBuildParams{
		Name:                cr.Name,
		Namespace:           cr.Namespace,
		GlobalClusterDomain: cr.Spec.Global.ClusterDomain,
		Kafka:               cr.Spec.Kafka,
		CertManager:         cr.Spec.CertManager,
		ArtemisExposeMode:   "route",
		HPA:                 cr.Spec.HPA,
		Strategy: provider.Ed254ProviderStrategy{
			Payload: provider.Ed254ProviderPayload{
				Postgres: cr.Spec.Postgres,
				Artemis:  cr.Spec.Artemis,
				Provider: cr.Spec.Provider.Ed254ProviderAppBaseSpec,
				Exposure: providerExposureFromRoutes(cr.Spec.Provider.Routes),
			},
		},
	}
}

func SwimFficeProviderToBuildParams(cr *appsv1alpha1.SwimFficeProvider) provider.ProviderBuildParams {
	return provider.ProviderBuildParams{
		Name:                cr.Name,
		Namespace:           cr.Namespace,
		GlobalClusterDomain: cr.Spec.Global.ClusterDomain,
		Kafka:               cr.Spec.Kafka,
		CertManager:         cr.Spec.CertManager,
		ArtemisExposeMode:   "route",
		HPA:                 cr.Spec.HPA,
		Strategy: provider.FficeProviderStrategy{
			Payload: provider.FficeProviderPayload{
				Postgres: cr.Spec.Postgres,
				Artemis:  cr.Spec.Artemis,
				Provider: cr.Spec.Provider.FficeProviderAppBaseSpec,
				Exposure: providerExposureFromRoutes(cr.Spec.Provider.Routes),
			},
		},
	}
}
