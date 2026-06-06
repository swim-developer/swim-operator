package controller

import (
	"fmt"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/controller/cv"
	"github.com/swim-developer/swim-operator-common/pkg/labels"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func buildCVRoute(p cv.CVBuildParams, managedBy string, host string, portName string) *routev1.Route {
	lbl := labels.StandardLabels(p.CRName, constants.ConsumerValidatorApp, p.CRName, managedBy)
	targetPort := intstr.FromString(portName)
	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-%s", p.CRName, portName), Namespace: p.Namespace, Labels: lbl},
		Spec: routev1.RouteSpec{
			Host: host,
			To:   routev1.RouteTargetReference{Kind: "Service", Name: p.CRName},
			Port: &routev1.RoutePort{TargetPort: targetPort},
		},
	}
	if portName == "https" {
		route.Spec.TLS = &routev1.TLSConfig{Termination: routev1.TLSTerminationPassthrough, InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect}
	} else {
		route.Spec.TLS = &routev1.TLSConfig{Termination: routev1.TLSTerminationEdge, InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect}
	}
	return route
}
