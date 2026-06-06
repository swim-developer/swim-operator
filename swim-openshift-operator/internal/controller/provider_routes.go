package controller

import (
	"context"
	"fmt"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/swim-developer/swim-operator-common/pkg/controller/provider"
	"github.com/swim-developer/swim-operator-common/pkg/labels"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ocpProviderAppRouteEdge(p provider.ProviderBuildParams, clusterDomain, managedBy string) *routev1.Route {
	name := p.Name
	ns := p.Namespace
	l := labels.StandardLabels(name, "provider", name, managedBy)
	ex := p.Strategy.Exposure()
	routeHost := fmt.Sprintf(HostnameSuffix, name, ns, clusterDomain)
	if ex.HTTPSEdgeHost != "" {
		routeHost = ex.HTTPSEdgeHost
	}
	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-http", name),
			Namespace: ns,
			Labels:    l,
		},
		Spec: routev1.RouteSpec{
			Host: routeHost,
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: name,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("http"),
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
			},
		},
	}
}

func ocpProviderAppRoutePassthrough(p provider.ProviderBuildParams, clusterDomain, managedBy string) *routev1.Route {
	name := p.Name
	ns := p.Namespace
	l := labels.StandardLabels(name, "provider", name, managedBy)
	ex := p.Strategy.Exposure()
	routeHost := fmt.Sprintf(MTLSHostnameSuffix, name, ns, clusterDomain)
	if ex.HTTPSPassthroughHost != "" {
		routeHost = ex.HTTPSPassthroughHost
	}
	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-https", name),
			Namespace: ns,
			Labels:    l,
		},
		Spec: routev1.RouteSpec{
			Host: routeHost,
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: name,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("https"),
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationPassthrough,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
			},
		},
	}
}

func reconcileProviderAppRoutes(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, p provider.ProviderBuildParams, clusterDomain, managedBy string) error {
	ex := p.Strategy.Exposure()
	if ex.HTTPEdgeEnabled {
		if err := reconcileRouteResource(ctx, c, scheme, owner, ocpProviderAppRouteEdge(p, clusterDomain, managedBy)); err != nil {
			return err
		}
	}
	if ex.HTTPSPassthroughEnabled {
		if err := reconcileRouteResource(ctx, c, scheme, owner, ocpProviderAppRoutePassthrough(p, clusterDomain, managedBy)); err != nil {
			return err
		}
	}
	return nil
}
