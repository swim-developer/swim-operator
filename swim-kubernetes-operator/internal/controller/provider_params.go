package controller

import (
	appsv1alpha1 "github.com/swim-developer/swim-kubernetes-operator/api/v1alpha1"
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	"github.com/swim-developer/swim-operator-common/pkg/controller/provider"
)

const upstreamPostgresImage = "docker.io/postgres:16"
const rhel9PostgresImage = "registry.redhat.io/rhel9/postgresql-16:latest"

func resolveKubernetesPostgres(spec commonapi.ProviderPostgresSpec) commonapi.ProviderPostgresSpec {
	if spec.Image == "" || spec.Image == rhel9PostgresImage {
		spec.Image = upstreamPostgresImage
	}
	return spec
}

func resolveKubernetesEd254Postgres(spec commonapi.Ed254PostgresSpec) commonapi.Ed254PostgresSpec {
	if spec.Image == "" || spec.Image == rhel9PostgresImage {
		spec.Image = upstreamPostgresImage
	}
	return spec
}

func SwimDigitalNotamProviderToBuildParams(cr *appsv1alpha1.SwimDigitalNotamProvider) provider.ProviderBuildParams {
	return provider.ProviderBuildParams{
		Name:                       cr.Name,
		Namespace:                  cr.Namespace,
		GlobalClusterDomain:        cr.Spec.Global.ClusterDomain,
		Kafka:                      cr.Spec.Kafka,
		CertManager:                cr.Spec.CertManager,
		ArtemisExposeMode:          "ingress",
		ArtemisIngressHostOverride: cr.Spec.Provider.Ingress.ArtemisHost,
		HPA:                        cr.Spec.HPA,
		PostgresUpstream:           true,
		ArtemisUpstream:            true,
		Strategy: provider.DnotamProviderStrategy{
			Payload: provider.DnotamProviderPayload{
				Postgres: resolveKubernetesPostgres(cr.Spec.Postgres),
				Artemis:  cr.Spec.Artemis,
				Provider: cr.Spec.Provider.ProviderAppBaseSpec,
				Exposure: provider.ProviderExposureSpec{
					HTTPEdgeEnabled: cr.Spec.Provider.Ingress.Enabled,
					HTTPSEdgeHost:   cr.Spec.Provider.Ingress.Host,
				},
			},
		},
	}
}

func SwimFficeProviderToBuildParams(cr *appsv1alpha1.SwimFficeProvider) provider.ProviderBuildParams {
	return provider.ProviderBuildParams{
		Name:                       cr.Name,
		Namespace:                  cr.Namespace,
		GlobalClusterDomain:        cr.Spec.Global.ClusterDomain,
		Kafka:                      cr.Spec.Kafka,
		CertManager:                cr.Spec.CertManager,
		ArtemisExposeMode:          "ingress",
		ArtemisIngressHostOverride: cr.Spec.Provider.Ingress.ArtemisHost,
		HPA:                        cr.Spec.HPA,
		PostgresUpstream:           true,
		ArtemisUpstream:            true,
		Strategy: provider.FficeProviderStrategy{
			Payload: provider.FficeProviderPayload{
				Postgres: resolveKubernetesEd254Postgres(cr.Spec.Postgres),
				Artemis:  cr.Spec.Artemis,
				Provider: cr.Spec.Provider.FficeProviderAppBaseSpec,
				Exposure: provider.ProviderExposureSpec{
					HTTPEdgeEnabled: cr.Spec.Provider.Ingress.Enabled,
					HTTPSEdgeHost:   cr.Spec.Provider.Ingress.Host,
				},
			},
		},
	}
}

func SwimEd254ProviderToBuildParams(cr *appsv1alpha1.SwimEd254Provider) provider.ProviderBuildParams {
	return provider.ProviderBuildParams{
		Name:                       cr.Name,
		Namespace:                  cr.Namespace,
		GlobalClusterDomain:        cr.Spec.Global.ClusterDomain,
		Kafka:                      cr.Spec.Kafka,
		CertManager:                cr.Spec.CertManager,
		ArtemisExposeMode:          "ingress",
		ArtemisIngressHostOverride: cr.Spec.Provider.Ingress.ArtemisHost,
		HPA:                        cr.Spec.HPA,
		PostgresUpstream:           true,
		ArtemisUpstream:            true,
		Strategy: provider.Ed254ProviderStrategy{
			Payload: provider.Ed254ProviderPayload{
				Postgres: resolveKubernetesEd254Postgres(cr.Spec.Postgres),
				Artemis:  cr.Spec.Artemis,
				Provider: cr.Spec.Provider.Ed254ProviderAppBaseSpec,
				Exposure: provider.ProviderExposureSpec{
					HTTPEdgeEnabled: cr.Spec.Provider.Ingress.Enabled,
					HTTPSEdgeHost:   cr.Spec.Provider.Ingress.Host,
				},
			},
		},
	}
}
