package v1alpha1

type ConsumerValidatorMariaDBSpec struct {
	// +optional
	// +kubebuilder:default=3306
	Port int32 `json:"port,omitempty"`
	// +optional
	// +kubebuilder:default="swim_consumer_validator"
	Database string `json:"database,omitempty"`
	// +optional
	// +kubebuilder:default="swim"
	Username string `json:"username,omitempty"`
	// +optional
	Password string `json:"password,omitempty"`
	// +optional
	ExistingSecret string `json:"existingSecret,omitempty"`
}

type ProviderValidatorMariaDBSpec struct {
	// +optional
	Host string `json:"host,omitempty"`
	// +optional
	// +kubebuilder:default=3306
	Port int32 `json:"port,omitempty"`
	// +optional
	// +kubebuilder:default="swim_provider_validator"
	Database string `json:"database,omitempty"`
	// +optional
	// +kubebuilder:default="swim"
	Username string `json:"username,omitempty"`
	// +optional
	Password string `json:"password,omitempty"`
	// +optional
	ExistingSecret string `json:"existingSecret,omitempty"`
}
