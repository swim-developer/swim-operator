package provider

import (
	"context"
	"fmt"
	"strings"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/swim-developer/swim-operator-common/pkg/domain"
	"github.com/swim-developer/swim-operator-common/pkg/labels"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func ProviderArtemisName(p ProviderBuildParams) string {
	return resources.DefaultArtemisName(p.Strategy.ArtemisSpecName(), p.Name)
}

func providerArtemisParams(p ProviderBuildParams, ingressHost string) resources.ArtemisProviderParams {
	ap := p.Strategy.ArtemisBaseParams(p, ingressHost)
	ap.Image = artemisImage(p)
	ap.InitImage = artemisInitImage(p)
	return ap
}

func artemisImage(p ProviderBuildParams) string {
	if p.ArtemisUpstream {
		return ""
	}
	return "registry.redhat.io/amq7/amq-broker-rhel9:7.13.2"
}

func artemisInitImage(p ProviderBuildParams) string {
	if p.ArtemisUpstream {
		return ""
	}
	return "quay.io/swim-developer/amq-broker-init-swim:7.13.2"
}

func artemisLabelsWithManagedBy(p ProviderBuildParams, managedBy string) map[string]string {
	an := ProviderArtemisName(p)
	return labels.StandardLabels(an, "artemis", p.Name, managedBy)
}

func BuildProviderArtemisCredentialsSecret(p ProviderBuildParams, managedBy string) *corev1.Secret {
	ap := providerArtemisParams(p, "")
	ap.Labels = artemisLabelsWithManagedBy(p, managedBy)
	return resources.BuildProviderArtemisCredentialsSecret(ap)
}

func BuildProviderArtemisKeystoreSecret(p ProviderBuildParams, managedBy string) *corev1.Secret {
	ap := providerArtemisParams(p, "")
	ap.Labels = artemisLabelsWithManagedBy(p, managedBy)
	return resources.BuildProviderArtemisKeystoreSecret(ap)
}

func BuildProviderArtemisOIDCSecret(p ProviderBuildParams, managedBy string) *corev1.Secret {
	return p.Strategy.ArtemisOIDCSecret(p, managedBy)
}

func BuildProviderArtemisAddressBPSecret(p ProviderBuildParams, managedBy string) *corev1.Secret {
	return p.Strategy.ArtemisAddressBPSecret(p, managedBy)
}

func BuildProviderArtemisSecurityBPSecret(p ProviderBuildParams, managedBy string) *corev1.Secret {
	return p.Strategy.ArtemisSecurityBPSecret(p, managedBy)
}

func BuildProviderArtemisCertificate(p ProviderBuildParams, managedBy string, ingressHost string) *certmanagerv1.Certificate {
	ap := providerArtemisParams(p, ingressHost)
	ap.Labels = artemisLabelsWithManagedBy(p, managedBy)
	return resources.BuildProviderArtemisCertificate(ap)
}

func BuildProviderArtemisBroker(p ProviderBuildParams, managedBy string, ingressHost string) *unstructured.Unstructured {
	ap := providerArtemisParams(p, ingressHost)
	ap.Labels = artemisLabelsWithManagedBy(p, managedBy)
	return resources.BuildProviderArtemisBroker(ap)
}

func BuildProviderArtemisJMXService(p ProviderBuildParams, managedBy string) *corev1.Service {
	ap := providerArtemisParams(p, "")
	ap.Labels = artemisLabelsWithManagedBy(p, managedBy)
	return resources.BuildProviderArtemisJMXService(ap.ArtemisName, ap.Namespace, ap.Labels, ap.JMXPort)
}

func BuildProviderKafkaConsole(ctx context.Context, p ProviderBuildParams, managedBy string, resolve func(context.Context, string, string) string) *unstructured.Unstructured {
	res := resolve
	if res == nil {
		res = func(_ context.Context, spec, _ string) string { return spec }
	}
	clusterDomain := res(ctx, p.GlobalClusterDomain, p.Namespace)
	hostname := fmt.Sprintf("kafka-ui-%s.%s", p.Namespace, domain.GetAppsDomain(strings.TrimPrefix(clusterDomain, "apps.")))
	return resources.BuildKafkaConsole(p.Namespace, hostname, labels.StandardLabels("kafka-ui", "kafka", p.Name, managedBy))
}
