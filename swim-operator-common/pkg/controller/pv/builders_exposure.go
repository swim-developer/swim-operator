package pv

import (
	"fmt"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PVRouteHost(p PVBuildParams, clusterDomain string) string {
	if p.RouteHost != "" {
		return p.RouteHost
	}
	return fmt.Sprintf(constants.HostnameSuffix, p.CRName, p.Namespace, clusterDomain)
}

func PVIngressHost(p PVBuildParams, clusterDomain string) string {
	if p.IngressHost != "" {
		return p.IngressHost
	}
	return fmt.Sprintf(constants.HostnameSuffix, p.CRName, p.Namespace, clusterDomain)
}

func BuildPVIngress(p PVBuildParams, managedBy, host string) *networkingv1.Ingress {
	pt := networkingv1.PathTypePrefix
	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        p.CRName,
			Namespace:   p.Namespace,
			Labels:      StandardLabels(p.CRName, managedBy),
			Annotations: p.IngressAnnotations,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{{
				Host: host,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{{
							Path: "/", PathType: &pt,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: p.CRName,
									Port: networkingv1.ServiceBackendPort{Number: 8080},
								},
							},
						}},
					},
				},
			}},
		},
	}
	if p.IngressTLSSecretName != "" {
		ing.Spec.TLS = []networkingv1.IngressTLS{{Hosts: []string{host}, SecretName: p.IngressTLSSecretName}}
	}
	return ing
}
