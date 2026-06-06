package controller

import (
	"context"

	"github.com/swim-developer/swim-operator-common/pkg/domain"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultClusterDomain    = domain.DefaultClusterDomain
	DefaultAppsDomain       = domain.DefaultAppsDomain
	DefaultServiceDomainFmt = domain.DefaultServiceDomainFmt
)

var (
	GetClusterDomain = domain.GetClusterDomain
	GetAppsDomain    = domain.GetAppsDomain
	ServiceFQDN      = domain.ServiceFQDN
	ServiceShortDNS  = domain.ServiceShortDNS
)

func getOrDetectClusterDomain(_ context.Context, _ client.Client, specDomain, _ string) string {
	return GetClusterDomain(specDomain)
}
