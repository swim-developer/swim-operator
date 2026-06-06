package controller

import (
	routev1 "github.com/openshift/api/route/v1"
	"github.com/swim-developer/swim-operator-common/pkg/controller/pv"
	appsv1alpha1 "github.com/swim-developer/swim-openshift-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func buildFficePVRoute(cr *appsv1alpha1.SwimFficeProviderValidator, clusterDomain, managedBy string) *routev1.Route {
	p := fficeOCPPVBuildParams(cr)
	host := pv.PVRouteHost(p, clusterDomain)
	labels := pv.StandardLabels(cr.Name, managedBy)
	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"haproxy.router.openshift.io/timeout": "300s",
			},
		},
		Spec: routev1.RouteSpec{
			Host: host,
			To: routev1.RouteTargetReference{
				Kind:   "Service",
				Name:   cr.Name,
				Weight: func() *int32 { w := int32(100); return &w }(),
			},
			Port: &routev1.RoutePort{TargetPort: intstr.FromString("http")},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
			},
		},
	}
}

func buildEd254PVRoute(cr *appsv1alpha1.SwimEd254ProviderValidator, clusterDomain, managedBy string) *routev1.Route {
	p := ed254OCPPVBuildParams(cr)
	host := pv.PVRouteHost(p, clusterDomain)
	labels := pv.StandardLabels(cr.Name, managedBy)
	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"haproxy.router.openshift.io/timeout": "300s",
			},
		},
		Spec: routev1.RouteSpec{
			Host: host,
			To: routev1.RouteTargetReference{
				Kind:   "Service",
				Name:   cr.Name,
				Weight: func() *int32 { w := int32(100); return &w }(),
			},
			Port: &routev1.RoutePort{TargetPort: intstr.FromString("http")},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
			},
		},
	}
}

func buildDnotamPVRoute(cr *appsv1alpha1.SwimDnotamProviderValidator, clusterDomain, managedBy string) *routev1.Route {
	p := dnotamOCPPVBuildParams(cr)
	host := pv.PVRouteHost(p, clusterDomain)
	labels := pv.StandardLabels(cr.Name, managedBy)
	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"haproxy.router.openshift.io/timeout": "300s",
			},
		},
		Spec: routev1.RouteSpec{
			Host: host,
			To: routev1.RouteTargetReference{
				Kind:   "Service",
				Name:   cr.Name,
				Weight: func() *int32 { w := int32(100); return &w }(),
			},
			Port: &routev1.RoutePort{TargetPort: intstr.FromString("http")},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
			},
		},
	}
}
