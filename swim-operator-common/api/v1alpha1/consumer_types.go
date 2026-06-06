package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type ClientSpec struct {
	// +optional
	// +kubebuilder:default="quay.io/masales/swim-dnotam-consumer:latest"
	Image string `json:"image,omitempty"`
	// +optional
	// +kubebuilder:default=1
	Replicas int32 `json:"replicas,omitempty"`
	// +optional
	Mongo MongoSpec `json:"mongo,omitempty"`
	// +optional
	Config ClientConfigSpec `json:"config,omitempty"`
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// +optional
	Probe ProbeConfig `json:"probe,omitempty"`
}

type ClientConfigSpec struct {
	// +optional
	Providers []ProviderSpec `json:"providers,omitempty"`
	// +optional
	DnotamSubscriptions []DnotamSubscriptionSpec `json:"dnotamSubscriptions,omitempty"`
	// +optional
	// +kubebuilder:default="true"
	SwimValidationEnabled string `json:"swimValidationEnabled,omitempty"`
	// +optional
	// +kubebuilder:default="false"
	SwimValidationFailOnNullBody string `json:"swimValidationFailOnNullBody,omitempty"`
	// +optional
	// +kubebuilder:default="true"
	DnotamDeleteAndRecreate string `json:"dnotamDeleteAndRecreate,omitempty"`
	// +optional
	Observability ObservabilitySpec `json:"observability,omitempty"`
}

type Ed254ConsumerAppSpec struct {
	// +optional
	// +kubebuilder:default="quay.io/masales/swim-ed254-consumer:latest"
	Image string `json:"image,omitempty"`
	// +optional
	// +kubebuilder:default=1
	Replicas int32 `json:"replicas,omitempty"`
	// +optional
	Mongo Ed254MongoSpec `json:"mongo,omitempty"`
	// +optional
	Config Ed254ConsumerConfigSpec `json:"config,omitempty"`
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// +optional
	Probe ProbeConfig `json:"probe,omitempty"`
}

type FficeConsumerAppSpec struct {
	// +optional
	// +kubebuilder:default="quay.io/masales/swim-ffice-consumer:latest"
	Image string `json:"image,omitempty"`
	// +optional
	// +kubebuilder:default=1
	Replicas int32 `json:"replicas,omitempty"`
	// +optional
	Mongo Ed254MongoSpec `json:"mongo,omitempty"`
	// +optional
	Config FficeConsumerConfigSpec `json:"config,omitempty"`
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// +optional
	Probe ProbeConfig `json:"probe,omitempty"`
}

type FficeConsumerConfigSpec struct {
	// +optional
	Providers []ProviderSpec `json:"providers,omitempty"`
	// +optional
	FficeSubscriptions []FficeSubscriptionSpec `json:"fficeSubscriptions,omitempty"`
	// +optional
	// +kubebuilder:default="true"
	SwimValidationEnabled string `json:"swimValidationEnabled,omitempty"`
	// +optional
	Observability ObservabilitySpec `json:"observability,omitempty"`
	// +optional
	// +kubebuilder:default=90
	HeartbeatTimeoutSeconds int32 `json:"heartbeatTimeoutSeconds,omitempty"`
}

type Ed254ConsumerConfigSpec struct {
	// +optional
	Providers []ProviderSpec `json:"providers,omitempty"`
	// +optional
	Ed254Subscriptions []Ed254SubscriptionSpec `json:"ed254Subscriptions,omitempty"`
	// +optional
	// +kubebuilder:default="true"
	SwimValidationEnabled string `json:"swimValidationEnabled,omitempty"`
	// +optional
	Observability ObservabilitySpec `json:"observability,omitempty"`
	// +optional
	// +kubebuilder:default=90
	HeartbeatTimeoutSeconds int32 `json:"heartbeatTimeoutSeconds,omitempty"`
}
