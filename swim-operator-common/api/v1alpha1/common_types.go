package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type GlobalSpec struct {
	// +optional
	// +kubebuilder:default=""
	ClusterDomain string `json:"clusterDomain,omitempty"`
}

type CertManagerSpec struct {
	// +optional
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`
	// +optional
	// +kubebuilder:default="swim-ca-issuer"
	IssuerName string `json:"issuerName,omitempty"`
	// +optional
	// +kubebuilder:default="ClusterIssuer"
	IssuerKind string `json:"issuerKind,omitempty"`
}

type ImageSpec struct {
	// +optional
	Repository string `json:"repository,omitempty"`
	// +optional
	Tag string `json:"tag,omitempty"`
	// +optional
	// +kubebuilder:default=Always
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`
}

type ProbeConfig struct {
	// +optional
	InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty"`
	// +optional
	PeriodSeconds int32 `json:"periodSeconds,omitempty"`
	// +optional
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty"`
	// +optional
	FailureThreshold int32 `json:"failureThreshold,omitempty"`
}

type ObservabilitySpec struct {
	// +optional
	// +kubebuilder:default=false
	OpenTelemetryEnabled bool `json:"openTelemetryEnabled,omitempty"`
	// +optional
	// +kubebuilder:default=""
	OtelEndpoint string `json:"otelEndpoint,omitempty"`
	// +optional
	// +kubebuilder:default=""
	OtelHeaders string `json:"otelHeaders,omitempty"`
	// +optional
	// +kubebuilder:default=false
	PrometheusEnabled bool `json:"prometheusEnabled,omitempty"`
	// +optional
	// +kubebuilder:default=false
	ServiceMonitorEnabled bool `json:"serviceMonitorEnabled,omitempty"`
}

type DatabaseSpec struct {
	// +optional
	Host string `json:"host,omitempty"`
	// +optional
	Port int32 `json:"port,omitempty"`
	// +optional
	Database string `json:"database,omitempty"`
	// +optional
	Username string `json:"username,omitempty"`
	// +optional
	Password string `json:"password,omitempty"`
	// +optional
	ExistingSecret string `json:"existingSecret,omitempty"`
	// +optional
	// +kubebuilder:default="1Gi"
	StorageSize string `json:"storageSize,omitempty"`
}

type HPAConfig struct {
	// +optional
	Enabled bool `json:"enabled,omitempty"`
	// +optional
	// +kubebuilder:default=1
	MinReplicas *int32 `json:"minReplicas,omitempty"`
	// +kubebuilder:validation:Required
	// +kubebuilder:default=1
	MaxReplicas int32 `json:"maxReplicas"`
	// +optional
	// +kubebuilder:default=80
	TargetCPUUtilizationPercentage *int32 `json:"targetCPUUtilizationPercentage,omitempty"`
}
