package provider

import (
	"fmt"

	"github.com/swim-developer/swim-operator-common/pkg/labels"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	corev1 "k8s.io/api/core/v1"
)

func BuildProviderAppConfigMap(p ProviderBuildParams, managedBy string, clusterDomain string) *corev1.ConfigMap {
	name := p.Name
	l := labels.StandardLabels(name, "provider", name, managedBy)
	data := p.Strategy.ConfigMapData(p, clusterDomain)
	applyOIDCTLSTrustStore(p, data)
	return resources.ConfigMap(fmt.Sprintf("%s-config", name), p.Namespace, l, data)
}

func BuildProviderAppSecret(p ProviderBuildParams, managedBy string) *corev1.Secret {
	l := labels.StandardLabels(p.Name, "provider", p.Name, managedBy)
	return resources.SecretStringData(fmt.Sprintf("%s-secret", p.Name), p.Namespace, l, p.Strategy.AppSecretData())
}

func BuildProviderAppOIDCSecret(p ProviderBuildParams, managedBy string) *corev1.Secret {
	l := labels.StandardLabels(p.Name, "provider", p.Name, managedBy)
	return resources.SecretStringData(fmt.Sprintf("%s-oidc-secret", p.Name), p.Namespace, l, p.Strategy.OIDCSecretData())
}

func applyOIDCTLSTrustStore(p ProviderBuildParams, data map[string]string) {
	if !p.ArtemisUpstream {
		return
	}
	data["QUARKUS_OIDC_TLS_CONFIGURATION_NAME"] = "oidc-client"
	data["QUARKUS_TLS_OIDC_CLIENT_TRUST_STORE_PEM_CERTS"] = "/certs/ca/ca.crt"
}
