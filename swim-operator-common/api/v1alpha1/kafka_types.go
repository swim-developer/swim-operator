package v1alpha1

// +kubebuilder:validation:XValidation:rule="self.deploymentMode != 'external' || (has(self.bootstrapServers) && size(self.bootstrapServers) > 0)",message="bootstrapServers is required when deploymentMode is external"
// +kubebuilder:validation:XValidation:rule="self.deploymentMode != 'external' || (has(self.username) && size(self.username) > 0)",message="username is required when deploymentMode is external"
// +kubebuilder:validation:XValidation:rule="self.deploymentMode != 'external' || (has(self.password) && size(self.password) > 0)",message="password is required when deploymentMode is external"
type KafkaSpec struct {
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`
	// +kubebuilder:default="managed"
	// +kubebuilder:validation:Enum=managed;external
	DeploymentMode string `json:"deploymentMode,omitempty"`
	// +kubebuilder:default="1Gi"
	StorageSize string `json:"storageSize,omitempty"`
	// +kubebuilder:default=1
	Replicas int32 `json:"replicas,omitempty"`
	// +kubebuilder:default="4.1.0"
	Version string `json:"version,omitempty"`
	// +optional
	BootstrapServers string `json:"bootstrapServers,omitempty"`
	// +optional
	Username string `json:"username,omitempty"`
	// +optional
	Password string `json:"password,omitempty"`
	// +optional
	// +kubebuilder:default=false
	KafkaConsoleEnabled bool `json:"kafkaConsoleEnabled,omitempty"`
}
