package v1alpha1

type ProviderSpec struct {
	// +kubebuilder:validation:Required
	ProviderId string `json:"providerId"`
	// +kubebuilder:validation:Required
	SubscriptionManager SubscriptionManagerSpec `json:"subscriptionManager"`
	// +kubebuilder:validation:Required
	AmqpBroker AmqpBrokerSpec `json:"amqpBroker"`
}

type SubscriptionManagerSpec struct {
	// +kubebuilder:validation:Required
	URL string `json:"url"`
	// +optional
	TLS *TlsSpec `json:"tls,omitempty"`
	// +optional
	Resilience *ResilienceSpec `json:"resilience,omitempty"`
}

type AmqpBrokerSpec struct {
	// +kubebuilder:validation:Required
	Host string `json:"host"`
	// +kubebuilder:validation:Required
	Port int32 `json:"port"`
	// +optional
	// +kubebuilder:default=true
	SSLEnabled bool `json:"sslEnabled,omitempty"`
	// +optional
	Username string `json:"username,omitempty"`
	// +optional
	Password string `json:"password,omitempty"`
	// +optional
	TLS *TlsSpec `json:"tls,omitempty"`
}

type TlsSpec struct {
	// +optional
	TrustStorePath string `json:"trustStorePath,omitempty"`
	// +optional
	TrustStorePassword string `json:"trustStorePassword,omitempty"`
	// +optional
	KeyStorePath string `json:"keyStorePath,omitempty"`
	// +optional
	KeyStorePassword string `json:"keyStorePassword,omitempty"`
}

type ResilienceSpec struct {
	// +optional
	// +kubebuilder:default=5000
	ConnectTimeoutMs int32 `json:"connectTimeoutMs,omitempty"`
	// +optional
	// +kubebuilder:default=30000
	ReadTimeoutMs int32 `json:"readTimeoutMs,omitempty"`
	// +optional
	// +kubebuilder:default=3
	RetryMaxAttempts int32 `json:"retryMaxAttempts,omitempty"`
	// +optional
	// +kubebuilder:default=1000
	RetryDelayMs int64 `json:"retryDelayMs,omitempty"`
}

type DnotamSubscriptionSpec struct {
	// +kubebuilder:validation:Required
	Topic string `json:"topic"`
	// +optional
	QueueName string `json:"queueName,omitempty"`
	// +optional
	EventScenario []string `json:"eventScenario,omitempty"`
	// +optional
	AirportHeliport []string `json:"airportHeliport,omitempty"`
	// +optional
	Airspace []string `json:"airspace,omitempty"`
	// +optional
	EventSeries string `json:"eventSeries,omitempty"`
	// +optional
	Publisher string `json:"publisher,omitempty"`
	// +kubebuilder:validation:Required
	Provider string `json:"provider"`
	// +optional
	Description string `json:"description,omitempty"`
	// +optional
	Comment string `json:"comment,omitempty"`
}
