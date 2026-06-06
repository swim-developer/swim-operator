package resources

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceMonitorParams struct {
	Name        string
	Namespace   string
	Labels      map[string]string
	MatchLabels map[string]string
	SelectorApp string
	PortName    string
	MetricsPath string
	Path        string
	Interval    string
}

func BuildServiceMonitor(p ServiceMonitorParams) *monitoringv1.ServiceMonitor {
	labels := make(map[string]string)
	for k, v := range p.Labels {
		labels[k] = v
	}
	labels["release"] = "prometheus-k8s"

	matchLabels := p.MatchLabels
	if matchLabels == nil && p.SelectorApp != "" {
		matchLabels = map[string]string{"app": p.SelectorApp}
	}

	metricsPath := p.MetricsPath
	if metricsPath == "" {
		metricsPath = p.Path
	}

	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{Name: p.Name, Namespace: p.Namespace, Labels: labels},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{MatchLabels: matchLabels},
			Endpoints: []monitoringv1.Endpoint{{
				Port:     StrDefault(p.PortName, "management"),
				Path:     StrDefault(metricsPath, "/q/metrics"),
				Interval: monitoringv1.Duration(StrDefault(p.Interval, "30s")),
			}},
		},
	}
}
