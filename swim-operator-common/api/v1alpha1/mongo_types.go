package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type MongoSpec struct {
	// +optional
	// +kubebuilder:default="1Gi"
	StorageSize string `json:"storageSize,omitempty"`
	// +kubebuilder:default="swim"
	User string `json:"user"`
	// +kubebuilder:default="swim"
	Password string `json:"password"`
	// +kubebuilder:default="swim-dnotam"
	Database string `json:"database"`
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type Ed254MongoSpec struct {
	// +optional
	// +kubebuilder:default="1Gi"
	StorageSize string `json:"storageSize,omitempty"`
	// +kubebuilder:default="swim"
	User string `json:"user"`
	// +kubebuilder:default="swim"
	Password string `json:"password"`
	// +kubebuilder:default="swim-ed254"
	Database string `json:"database"`
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}
