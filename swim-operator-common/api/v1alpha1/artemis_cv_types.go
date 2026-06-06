package v1alpha1

type ArtemisSpec struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	Broker ArtemisBrokerSpec `json:"broker,omitempty"`
	// +optional
	Credentials ArtemisCredentialsSpec `json:"credentials,omitempty"`
	// +optional
	Acceptors ArtemisAcceptorsSpec `json:"acceptors,omitempty"`
	// +optional
	Console ArtemisConsoleSpec `json:"console,omitempty"`
	// +optional
	Persistence ArtemisPersistenceSpec `json:"persistence,omitempty"`
	// +optional
	BrokerProperties []string `json:"brokerProperties,omitempty"`
	// +optional
	CertManager ArtemisCertManagerSpec `json:"certManager,omitempty"`
}

type ArtemisBrokerSpec struct {
	// +optional
	Name string `json:"name,omitempty"`
	// +optional
	Image string `json:"image,omitempty"`
	// +optional
	Size int32 `json:"size,omitempty"`
}

type ArtemisCredentialsSpec struct {
	// +optional
	AdminUser string `json:"adminUser,omitempty"`
	// +optional
	AdminPassword string `json:"adminPassword,omitempty"`
	// +optional
	ClusterUser string `json:"clusterUser,omitempty"`
	// +optional
	ClusterPassword string `json:"clusterPassword,omitempty"`
}

type ArtemisAcceptorsSpec struct {
	// +optional
	Amqps ArtemisAmqpsSpec `json:"amqps,omitempty"`
	// +optional
	Amqp ArtemisAmqpSpec `json:"amqp,omitempty"`
}

type ArtemisAmqpsSpec struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	Port int32 `json:"port,omitempty"`
	// +optional
	VerifyHost *bool `json:"verifyHost,omitempty"`
	// +optional
	Expose *bool `json:"expose,omitempty"`
	// +optional
	ExposeMode string `json:"exposeMode,omitempty"`
	// +optional
	IngressHost string `json:"ingressHost,omitempty"`
	// +optional
	NeedClientAuth *bool `json:"needClientAuth,omitempty"`
}

type ArtemisAmqpSpec struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	Port int32 `json:"port,omitempty"`
	// +optional
	Expose *bool `json:"expose,omitempty"`
}

type ArtemisConsoleSpec struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	Expose *bool `json:"expose,omitempty"`
	// +optional
	SSLEnabled *bool `json:"sslEnabled,omitempty"`
}

type ArtemisPersistenceSpec struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	JournalType string `json:"journalType,omitempty"`
	// +optional
	RequireLogin *bool `json:"requireLogin,omitempty"`
}

type ArtemisCertManagerSpec struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	IssuerRef IssuerRefSpec `json:"issuerRef,omitempty"`
	// +optional
	CommonName string `json:"commonName,omitempty"`
	// +optional
	DNSNames []string `json:"dnsNames,omitempty"`
	// +optional
	KeystorePassword string `json:"keystorePassword,omitempty"`
}

type IssuerRefSpec struct {
	// +optional
	Name string `json:"name,omitempty"`
	// +optional
	Kind string `json:"kind,omitempty"`
}
