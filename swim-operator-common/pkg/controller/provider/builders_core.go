package provider

import (
	"fmt"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/labels"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildProviderServiceAccount(p ProviderBuildParams, managedBy string) *corev1.ServiceAccount {
	name := p.Name
	return resources.StandardServiceAccount(name, p.Namespace, labels.StandardLabels(name, "provider", name, managedBy))
}

func BuildProviderRole(p ProviderBuildParams, managedBy string) *rbacv1.Role {
	name := p.Name
	ns := p.Namespace
	lbl := labels.StandardLabels(name, "provider", name, managedBy)
	rules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"secrets", "configmaps"},
			Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
		},
	}
	rules = append(rules, p.Strategy.AdditionalRoleRules()...)
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-role", name),
			Namespace: ns,
			Labels:    lbl,
		},
		Rules: rules,
	}
}

func BuildProviderRoleBinding(p ProviderBuildParams, managedBy string) *rbacv1.RoleBinding {
	name := p.Name
	return resources.BuildRoleBinding(name, p.Namespace, name, fmt.Sprintf("%s-role", name), labels.StandardLabels(name, "provider", name, managedBy))
}

func BuildProviderAppServerCertificate(p ProviderBuildParams, managedBy string, clusterDomain string) *certmanagerv1.Certificate {
	name := p.Name
	ns := p.Namespace
	l := labels.StandardLabels(name, "provider", name, managedBy)
	issuerName := p.CertManager.IssuerName
	issuerKind := p.CertManager.IssuerKind
	hpv := fmt.Sprintf(constants.HostnameSuffix, name, ns, clusterDomain)
	mt := fmt.Sprintf(constants.MTLSHostnameSuffix, name, ns, clusterDomain)
	dns := []string{
		hpv,
		mt,
		fmt.Sprintf("%s.%s.svc.cluster.local", name, ns),
		fmt.Sprintf("%s.%s.svc", name, ns),
		fmt.Sprintf("%s.%s", name, ns),
		name,
		"localhost",
	}
	if ch := p.Strategy.Exposure().HTTPSEdgeHost; ch != "" && ch != hpv {
		dns = append([]string{ch}, dns...)
	}
	return resources.BuildCertificate(resources.CertificateParams{
		Name:       fmt.Sprintf("%s-server", name),
		Namespace:  ns,
		Labels:     l,
		SecretName: fmt.Sprintf(constants.ServerTLSSuffix, name),
		IssuerName: issuerName,
		IssuerKind: issuerKind,
		CommonName: name,
		DNSNames:   dns,
		Usages:     []certmanagerv1.KeyUsage{certmanagerv1.UsageServerAuth, certmanagerv1.UsageClientAuth},
	})
}

func BuildProviderAppService(p ProviderBuildParams, managedBy string) *corev1.Service {
	name := p.Name
	ns := p.Namespace
	l := labels.StandardLabels(name, "provider", name, managedBy)
	return resources.ServiceClusterIP(name, ns, l, map[string]string{"app": name}, []corev1.ServicePort{
		resources.ServicePortTCP("http", 8080, 8080),
		resources.ServicePortTCP("https", 8443, 8443),
		resources.ServicePortTCP("management", 9000, 9000),
		resources.ServicePortTCP("internal", 9080, 9080),
	})
}

func BuildProviderAppServiceMonitor(p ProviderBuildParams, managedBy string) *monitoringv1.ServiceMonitor {
	name := p.Name
	ns := p.Namespace
	return resources.BuildServiceMonitor(resources.ServiceMonitorParams{
		Name:        fmt.Sprintf("%s-metrics", name),
		Namespace:   ns,
		Labels:      labels.StandardLabels(name, "monitoring", name, managedBy),
		SelectorApp: name,
		PortName:    "management",
		Path:        "/q/metrics",
		Interval:    "30s",
	})
}

func BuildProviderHPA(p ProviderBuildParams, managedBy string) *autoscalingv2.HorizontalPodAutoscaler {
	minReplicas := p.HPA.MinReplicas
	if minReplicas == nil {
		def := int32(1)
		minReplicas = &def
	}
	return resources.BuildHPA(resources.HPAParams{
		Name:                           fmt.Sprintf("%s-hpa", p.Name),
		Namespace:                      p.Namespace,
		Labels:                         labels.StandardLabels(p.Name, "provider", p.Name, managedBy),
		TargetName:                     p.Name,
		MinReplicas:                    minReplicas,
		MaxReplicas:                    resources.Int32Default(p.HPA.MaxReplicas, 5),
		CPUUtilization:                 p.HPA.TargetCPUUtilizationPercentage,
		TargetCPUUtilizationPercentage: 70,
		ScaleUpStabilization:           60,
		ScaleDownStabilization:         300,
	})
}
