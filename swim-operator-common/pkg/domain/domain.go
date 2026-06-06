package domain

import "fmt"

const (
	DefaultClusterDomain    = "cluster.local"
	DefaultAppsDomain       = "apps.cluster.local"
	DefaultServiceDomainFmt = "%s.%s.svc.%s"
)

func GetClusterDomain(specDomain string) string {
	if specDomain != "" {
		return specDomain
	}
	return DefaultClusterDomain
}

func GetAppsDomain(specDomain string) string {
	if specDomain != "" {
		return "apps." + specDomain
	}
	return DefaultAppsDomain
}

func ServiceFQDN(name, namespace, clusterDomain string) string {
	d := GetClusterDomain(clusterDomain)
	return fmt.Sprintf(DefaultServiceDomainFmt, name, namespace, d)
}

func ServiceShortDNS(name, namespace string) string {
	return fmt.Sprintf("%s.%s.svc", name, namespace)
}
