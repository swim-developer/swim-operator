package v1alpha1

type AppConfigSpec struct {
	// +optional
	Quarkus QuarkusSpec `json:"quarkus,omitempty"`
	// +optional
	EventGenerator EventGeneratorSpec `json:"eventGenerator,omitempty"`
	// +optional
	Amqp AmqpSpec `json:"amqp,omitempty"`
}

type QuarkusSpec struct {
	// +optional
	// +kubebuilder:default=8080
	HTTPPort int32 `json:"httpPort,omitempty"`
	// +optional
	// +kubebuilder:default=8443
	SSLPort int32 `json:"sslPort,omitempty"`
	// +optional
	LogLevel string `json:"logLevel,omitempty"`
}

type EventGeneratorSpec struct {
	// +optional
	Enabled string `json:"enabled,omitempty"`
	// +optional
	Schedule string `json:"schedule,omitempty"`
	// +optional
	EventsPath string `json:"eventsPath,omitempty"`
	// +optional
	ExceptionsPath string `json:"exceptionsPath,omitempty"`
}

type AmqpSpec struct {
	// +optional
	Host string `json:"host,omitempty"`
	// +optional
	// +kubebuilder:default=5672
	Port int32 `json:"port,omitempty"`
	// +optional
	Username string `json:"username,omitempty"`
	// +optional
	Password string `json:"password,omitempty"`
	// +optional
	ExistingSecret string `json:"existingSecret,omitempty"`
}
