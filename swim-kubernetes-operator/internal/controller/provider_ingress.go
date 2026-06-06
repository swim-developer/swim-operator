package controller

import (
	"context"
	"fmt"

	appsv1alpha1 "github.com/swim-developer/swim-kubernetes-operator/api/v1alpha1"
	swimconst "github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/labels"
	commonreconciler "github.com/swim-developer/swim-operator-common/pkg/reconciler"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func k8sDnotamProviderIngress(cr *appsv1alpha1.SwimDigitalNotamProvider, clusterDomain string) *networkingv1.Ingress {
	l := labels.StandardLabels(cr.Name, "provider", cr.Name, sharedManagedByValue)
	host := cr.Spec.Provider.Ingress.Host
	if host == "" {
		host = fmt.Sprintf(swimconst.HostnameSuffix, cr.Name, cr.Namespace, clusterDomain)
	}
	pt := networkingv1.PathTypePrefix
	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-http", cr.Name),
			Namespace:   cr.Namespace,
			Labels:      l,
			Annotations: cr.Spec.Provider.Ingress.Annotations,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{{
				Host: host,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{{
							Path:     "/",
							PathType: &pt,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: cr.Name,
									Port: networkingv1.ServiceBackendPort{Number: 8080},
								},
							},
						}},
					},
				},
			}},
		},
	}
	if tls := resources.StrDefault(cr.Spec.Provider.Ingress.TLSSecretName, ""); tls != "" {
		ing.Spec.TLS = []networkingv1.IngressTLS{{Hosts: []string{host}, SecretName: tls}}
	}
	return ing
}

func k8sEd254ProviderIngress(cr *appsv1alpha1.SwimEd254Provider, host string) *networkingv1.Ingress {
	l := labels.StandardLabels(cr.Name, "provider", cr.Name, sharedManagedByValue)
	annotations := map[string]string{}
	if cr.Spec.Provider.Ingress.Annotations != nil {
		for k, v := range cr.Spec.Provider.Ingress.Annotations {
			annotations[k] = v
		}
	}
	pt := networkingv1.PathTypePrefix
	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-http", cr.Name),
			Namespace:   cr.Namespace,
			Labels:      l,
			Annotations: annotations,
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
									Name: cr.Name,
									Port: networkingv1.ServiceBackendPort{Number: 8080},
								},
							},
						}},
					},
				},
			}},
		},
	}
	if tls := cr.Spec.Provider.Ingress.TLSSecretName; tls != "" {
		ing.Spec.TLS = []networkingv1.IngressTLS{{Hosts: []string{host}, SecretName: tls}}
	}
	return ing
}

func reconcileK8sDnotamProviderIngress(ctx context.Context, c client.Client, scheme *runtime.Scheme, cr *appsv1alpha1.SwimDigitalNotamProvider, clusterDomain string) error {
	if !cr.Spec.Provider.Ingress.Enabled {
		return nil
	}
	return commonreconciler.ReconcileIngress(ctx, c, scheme, cr, k8sDnotamProviderIngress(cr, clusterDomain))
}

func reconcileK8sEd254ProviderIngress(ctx context.Context, c client.Client, scheme *runtime.Scheme, cr *appsv1alpha1.SwimEd254Provider, clusterDomain string) error {
	if !cr.Spec.Provider.Ingress.Enabled {
		return nil
	}
	host := cr.Spec.Provider.Ingress.Host
	if host == "" {
		host = fmt.Sprintf(swimconst.HostnameSuffix, cr.Name, cr.Namespace, clusterDomain)
	}
	return commonreconciler.ReconcileIngress(ctx, c, scheme, cr, k8sEd254ProviderIngress(cr, host))
}

func k8sFficeProviderIngress(cr *appsv1alpha1.SwimFficeProvider, host string) *networkingv1.Ingress {
	l := labels.StandardLabels(cr.Name, "provider", cr.Name, sharedManagedByValue)
	annotations := map[string]string{}
	if cr.Spec.Provider.Ingress.Annotations != nil {
		for k, v := range cr.Spec.Provider.Ingress.Annotations {
			annotations[k] = v
		}
	}
	pt := networkingv1.PathTypePrefix
	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-http", cr.Name),
			Namespace:   cr.Namespace,
			Labels:      l,
			Annotations: annotations,
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
									Name: cr.Name,
									Port: networkingv1.ServiceBackendPort{Number: 8080},
								},
							},
						}},
					},
				},
			}},
		},
	}
	if tls := cr.Spec.Provider.Ingress.TLSSecretName; tls != "" {
		ing.Spec.TLS = []networkingv1.IngressTLS{{Hosts: []string{host}, SecretName: tls}}
	}
	return ing
}

func reconcileK8sFficeProviderIngress(ctx context.Context, c client.Client, scheme *runtime.Scheme, cr *appsv1alpha1.SwimFficeProvider, clusterDomain string) error {
	if !cr.Spec.Provider.Ingress.Enabled {
		return nil
	}
	host := cr.Spec.Provider.Ingress.Host
	if host == "" {
		host = fmt.Sprintf(swimconst.HostnameSuffix, cr.Name, cr.Namespace, clusterDomain)
	}
	return commonreconciler.ReconcileIngress(ctx, c, scheme, cr, k8sFficeProviderIngress(cr, host))
}
