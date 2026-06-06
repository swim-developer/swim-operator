package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type ProviderAppBaseSpec struct {
	// +optional
	// +kubebuilder:default="quay.io/masales/swim-dnotam-provider:latest"
	Image string `json:"image,omitempty"`
	// +optional
	// +kubebuilder:default=1
	Replicas int32 `json:"replicas,omitempty"`
	// +optional
	// +kubebuilder:default="INFO"
	LogLevel string `json:"logLevel,omitempty"`
	// +kubebuilder:validation:Required
	// +kubebuilder:default=false
	ConsumeFromClientTopics bool `json:"consumeFromClientTopics"`
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// +optional
	Observability ObservabilitySpec `json:"observability,omitempty"`
	// +kubebuilder:validation:Required
	OIDC ProviderOIDCSpec `json:"oidc"`
}

type FficeProviderAppBaseSpec struct {
	// +optional
	// +kubebuilder:default="quay.io/masales/swim-ffice-provider:latest"
	Image string `json:"image,omitempty"`
	// +optional
	// +kubebuilder:default=1
	Replicas int32 `json:"replicas,omitempty"`
	// +optional
	// +kubebuilder:default="INFO"
	LogLevel string `json:"logLevel,omitempty"`
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// +optional
	Observability ObservabilitySpec `json:"observability,omitempty"`
	// +kubebuilder:validation:Required
	OIDC ProviderOIDCSpec `json:"oidc"`
	// +optional
	// +kubebuilder:default=15
	HeartbeatIntervalSeconds int32 `json:"heartbeatIntervalSeconds,omitempty"`
	// +optional
	SwimTopics []string `json:"swimTopics,omitempty"`
}

type Ed254ProviderAppBaseSpec struct {
	// +optional
	// +kubebuilder:default="quay.io/masales/swim-ed254-provider:latest"
	Image string `json:"image,omitempty"`
	// +optional
	// +kubebuilder:default=1
	Replicas int32 `json:"replicas,omitempty"`
	// +optional
	// +kubebuilder:default="INFO"
	LogLevel string `json:"logLevel,omitempty"`
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// +optional
	Observability ObservabilitySpec `json:"observability,omitempty"`
	// +kubebuilder:validation:Required
	OIDC ProviderOIDCSpec `json:"oidc"`
	// +optional
	Aerodromes []string `json:"aerodromes,omitempty"`
	// +optional
	// +kubebuilder:default=30
	HeartbeatIntervalSeconds int32 `json:"heartbeatIntervalSeconds,omitempty"`
	// +optional
	// +kubebuilder:default=100
	MaxConsumers int32 `json:"maxConsumers,omitempty"`
}
