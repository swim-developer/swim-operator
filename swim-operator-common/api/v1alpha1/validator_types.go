package v1alpha1

type ProviderValidatorKeycloakSpec struct {
	// +kubebuilder:validation:Required
	URL string `json:"url"`
	// +kubebuilder:validation:Required
	Realm string `json:"realm"`
	// +kubebuilder:validation:Required
	ClientID string `json:"clientId"`
}

type ProviderValidatorMTLSSpec struct {
	// +optional
	Enabled bool `json:"enabled,omitempty"`
	// +optional
	CertsSecretName string `json:"certsSecretName,omitempty"`
	// +optional
	PasswordsSecretName string `json:"passwordsSecretName,omitempty"`
}

type ProviderValidatorAMQPSpec struct {
	// +kubebuilder:validation:Required
	Host string `json:"host"`
	// +optional
	Port int32 `json:"port,omitempty"`
}
