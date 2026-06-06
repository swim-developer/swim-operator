package cv

import (
	"fmt"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/labels"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func buildCVIngress(p CVBuildParams, managedBy string, ingressName, host string, port int32, tlsSecret string) *networkingv1.Ingress {
	pt := networkingv1.PathTypePrefix
	lbl := labels.StandardLabels(p.CRName, constants.ConsumerValidatorApp, p.CRName, managedBy)
	ann := map[string]string{}
	if p.Ingress.Annotations != nil {
		for k, v := range p.Ingress.Annotations {
			ann[k] = v
		}
	}
	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Name: ingressName, Namespace: p.Namespace, Labels: lbl, Annotations: ann},
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
									Port: networkingv1.ServiceBackendPort{Number: port},
								},
							},
						}},
					},
				},
			}},
		},
	}
	if tlsSecret != "" {
		ing.Spec.TLS = []networkingv1.IngressTLS{{Hosts: []string{host}, SecretName: tlsSecret}}
	}
	return ing
}

func BuildCVIngressHTTPS(p CVBuildParams, managedBy string, host string) *networkingv1.Ingress {
	httpPort := resources.Int32Default(p.Spec.AppConfig.Quarkus.HTTPPort, 8080)
	name := fmt.Sprintf("%s-https", p.CRName)
	tls := p.Ingress.TLSSecretName
	if tls == "" {
		tls = fmt.Sprintf(constants.ServerTLSSuffix, p.CRName)
	}
	return buildCVIngress(p, managedBy, name, host, httpPort, tls)
}

func BuildCVIngressAPI(p CVBuildParams, managedBy string, host string) *networkingv1.Ingress {
	sslPort := resources.Int32Default(p.Spec.AppConfig.Quarkus.SSLPort, 8443)
	name := fmt.Sprintf("%s-api", p.CRName)
	tls := p.Ingress.TLSSecretName
	if tls == "" {
		tls = fmt.Sprintf(constants.ServerTLSSuffix, p.CRName)
	}
	ing := buildCVIngress(p, managedBy, name, host, sslPort, tls)
	ing.Annotations["nginx.ingress.kubernetes.io/ssl-passthrough"] = "true"
	return ing
}

