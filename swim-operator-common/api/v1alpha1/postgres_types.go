package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type ProviderPostgresSpec struct {
	// +optional
	// +kubebuilder:default="registry.redhat.io/rhel9/postgresql-16:latest"
	Image string `json:"image,omitempty"`
	// +optional
	// +kubebuilder:default="5Gi"
	StorageSize string `json:"storageSize,omitempty"`
	// +optional
	// +kubebuilder:default="swim-dnotam"
	Database string `json:"database,omitempty"`
	// +optional
	// +kubebuilder:default="swim-provider"
	User string `json:"user,omitempty"`
	// +optional
	// +kubebuilder:default="swim-provider"
	Password string `json:"password,omitempty"`
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type Ed254PostgresSpec struct {
	// +optional
	// +kubebuilder:default="registry.redhat.io/rhel9/postgresql-16:latest"
	Image string `json:"image,omitempty"`
	// +optional
	// +kubebuilder:default="5Gi"
	StorageSize string `json:"storageSize,omitempty"`
	// +optional
	// +kubebuilder:default="swim-ed254"
	Database string `json:"database,omitempty"`
	// +optional
	// +kubebuilder:default="swim-ed254-provider"
	User string `json:"user,omitempty"`
	// +optional
	// +kubebuilder:default="swim-ed254-provider"
	Password string `json:"password,omitempty"`
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}
