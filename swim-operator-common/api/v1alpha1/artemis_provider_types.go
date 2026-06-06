package v1alpha1

type ProviderArtemisSpec struct {
	// +optional
	Name string `json:"name,omitempty"`
	// +optional
	// +kubebuilder:default=1
	Size int32 `json:"size,omitempty"`
	// +optional
	// +kubebuilder:default="admin"
	AdminUser string `json:"adminUser,omitempty"`
	// +optional
	// +kubebuilder:default="admin"
	AdminPassword string `json:"adminPassword,omitempty"`
	// +optional
	// +kubebuilder:default="changeit"
	KeystorePassword string `json:"keystorePassword,omitempty"`
	// +optional
	Storage ProviderArtemisStorageSpec `json:"storage,omitempty"`
	// +optional
	Acceptors ProviderArtemisAcceptorsSpec `json:"acceptors,omitempty"`
	// +optional
	Console ProviderArtemisConsoleSpec `json:"console,omitempty"`
	// +optional
	JMX ArtemisJMXSpec `json:"jmx,omitempty"`
	// +optional
	BrokerProperties []string `json:"brokerProperties,omitempty"`
	// +kubebuilder:validation:Required
	OIDC ArtemisOIDCSpec `json:"oidc"`
}

type ProviderArtemisStorageSpec struct {
	// +optional
	// +kubebuilder:default="5Gi"
	// +kubebuilder:validation:Pattern=`^([5-9]|[1-9][0-9]+)Gi$`
	Size string `json:"size,omitempty"`
	// +optional
	StorageClassName string `json:"storageClassName,omitempty"`
}

type ProviderArtemisAcceptorsSpec struct {
	// +optional
	AMQPSPort int32 `json:"amqpsPort,omitempty"`
	// +optional
	AMQPPort int32 `json:"amqpPort,omitempty"`
	// +optional
	IngressHost string `json:"ingressHost,omitempty"`
	// +kubebuilder:validation:Required
	// +kubebuilder:default=false
	VerifyHost bool `json:"verifyHost"`
}

type ProviderArtemisConsoleSpec struct {
	// +optional
	SSLEnabled *bool `json:"sslEnabled,omitempty"`
	// +optional
	Expose bool `json:"expose,omitempty"`
}

type ArtemisJMXSpec struct {
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`
	// +kubebuilder:default=1099
	Port int32 `json:"port,omitempty"`
}

type ArtemisOIDCSpec struct {
	// +optional
	Enabled bool `json:"enabled,omitempty"`
	// +kubebuilder:validation:Required
	AuthServerUrl string `json:"authServerUrl"`
	// +kubebuilder:validation:Required
	Realm string `json:"realm"`
	// +kubebuilder:validation:Required
	// +kubebuilder:default="amq-broker"
	ClientId string `json:"clientId"`
	// +optional
	ClientSecret string `json:"clientSecret,omitempty"`
}

type ProviderOIDCSpec struct {
	// +kubebuilder:validation:Required
	AuthServerUrl string `json:"authServerUrl"`
	// +kubebuilder:validation:Required
	ClientId string `json:"clientId"`
	// +kubebuilder:validation:Required
	ClientSecret string `json:"clientSecret"`
}

type Ed254ArtemisSpec struct {
	// +kubebuilder:default="swim-artemis"
	Name string `json:"name"`
	// +optional
	// +kubebuilder:default=1
	Size int32 `json:"size,omitempty"`
	// +kubebuilder:default="admin"
	AdminUser string `json:"adminUser"`
	// +kubebuilder:default="admin"
	AdminPassword string `json:"adminPassword"`
	// +optional
	// +kubebuilder:default="changeit"
	KeystorePassword string `json:"keystorePassword,omitempty"`
	// +optional
	Storage ProviderArtemisStorageSpec `json:"storage,omitempty"`
	// +optional
	Acceptors ProviderArtemisAcceptorsSpec `json:"acceptors,omitempty"`
	// +optional
	Console ProviderArtemisConsoleSpec `json:"console,omitempty"`
	// +optional
	JMX ArtemisJMXSpec `json:"jmx,omitempty"`
	// +optional
	BrokerProperties []string `json:"brokerProperties,omitempty"`
	// +kubebuilder:validation:Required
	OIDC ArtemisOIDCSpec `json:"oidc"`
}
