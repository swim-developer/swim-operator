package v1alpha1

type IngressSpec struct {
	// +optional
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`
	// Host is the primary ingress hostname (HTTPS for Consumer Validators, HTTP for Providers).
	// Auto-generated from CR name and namespace if empty.
	// +optional
	Host string `json:"host,omitempty"`
	// ArtemisHost is the hostname for the Artemis AMQP ingress (Providers only).
	// Defaults to an auto-generated hostname based on the Artemis service name if empty.
	// +optional
	ArtemisHost string `json:"artemisHost,omitempty"`
	// APIHost is the hostname for the mTLS Subscription Manager API ingress (Consumer Validators only).
	// Defaults to a hostname derived from Host by appending "-api" to the CR name segment if empty.
	// +optional
	APIHost string `json:"apiHost,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// +optional
	TLSSecretName string `json:"tlsSecretName,omitempty"`
}
