package v1alpha1

type SwimDigitalNotamConsumerSpec struct {
	// +optional
	Global GlobalSpec `json:"global,omitempty"`
	// +optional
	CertManager CertManagerSpec `json:"certManager,omitempty"`
	// +optional
	Kafka KafkaSpec `json:"kafka,omitempty"`
	// +optional
	Client ClientSpec `json:"client,omitempty"`
	// +optional
	HPA HPAConfig `json:"hpa,omitempty"`
}

type SwimEd254ConsumerSpec struct {
	// +optional
	Global GlobalSpec `json:"global,omitempty"`
	// +optional
	CertManager CertManagerSpec `json:"certManager,omitempty"`
	// +optional
	Kafka KafkaSpec `json:"kafka,omitempty"`
	// +optional
	Consumer Ed254ConsumerAppSpec `json:"consumer,omitempty"`
	// +optional
	HPA HPAConfig `json:"hpa,omitempty"`
}

type SwimDnotamConsumerValidatorSpec struct {
	// +optional
	ReplicaCount *int32 `json:"replicaCount,omitempty"`
	// +optional
	Global GlobalSpec `json:"global,omitempty"`
	// +optional
	AppConfig AppConfigSpec `json:"appConfig,omitempty"`
	// +optional
	MariaDB ConsumerValidatorMariaDBSpec `json:"mariadb,omitempty"`
	// +optional
	Artemis ArtemisSpec `json:"artemis,omitempty"`
	// +optional
	CertManager CertManagerSpec `json:"certManager,omitempty"`
	// +optional
	Image ImageSpec `json:"image,omitempty"`
	// +optional
	HPA HPAConfig `json:"hpa,omitempty"`
}

type SwimEd254ConsumerValidatorSpec struct {
	// +optional
	ReplicaCount *int32 `json:"replicaCount,omitempty"`
	// +optional
	Global GlobalSpec `json:"global,omitempty"`
	// +optional
	AppConfig AppConfigSpec `json:"appConfig,omitempty"`
	// +optional
	MariaDB ConsumerValidatorMariaDBSpec `json:"mariadb,omitempty"`
	// +optional
	Artemis ArtemisSpec `json:"artemis,omitempty"`
	// +optional
	CertManager CertManagerSpec `json:"certManager,omitempty"`
	// +optional
	Image ImageSpec `json:"image,omitempty"`
	// +optional
	HPA HPAConfig `json:"hpa,omitempty"`
}

type SwimDigitalNotamProviderBaseSpec struct {
	// +optional
	Global GlobalSpec `json:"global,omitempty"`
	// +optional
	CertManager CertManagerSpec `json:"certManager,omitempty"`
	// +optional
	Postgres ProviderPostgresSpec `json:"postgres,omitempty"`
	// +optional
	Artemis ProviderArtemisSpec `json:"artemis,omitempty"`
	// +optional
	Kafka KafkaSpec `json:"kafka,omitempty"`
	// +optional
	HPA HPAConfig `json:"hpa,omitempty"`
}

type SwimEd254ProviderBaseSpec struct {
	// +optional
	Global GlobalSpec `json:"global,omitempty"`
	// +optional
	CertManager CertManagerSpec `json:"certManager,omitempty"`
	// +optional
	Postgres Ed254PostgresSpec `json:"postgres,omitempty"`
	// +optional
	Artemis Ed254ArtemisSpec `json:"artemis,omitempty"`
	// +optional
	Kafka KafkaSpec `json:"kafka,omitempty"`
	// +optional
	HPA HPAConfig `json:"hpa,omitempty"`
}

type SwimFficeConsumerSpec struct {
	// +optional
	Global GlobalSpec `json:"global,omitempty"`
	// +optional
	CertManager CertManagerSpec `json:"certManager,omitempty"`
	// +optional
	Kafka KafkaSpec `json:"kafka,omitempty"`
	// +optional
	Consumer FficeConsumerAppSpec `json:"consumer,omitempty"`
	// +optional
	HPA HPAConfig `json:"hpa,omitempty"`
}

type SwimFficeConsumerValidatorSpec = SwimEd254ConsumerValidatorSpec

type SwimFficeProviderBaseSpec struct {
	// +optional
	Global GlobalSpec `json:"global,omitempty"`
	// +optional
	CertManager CertManagerSpec `json:"certManager,omitempty"`
	// +optional
	Postgres Ed254PostgresSpec `json:"postgres,omitempty"`
	// +optional
	Artemis Ed254ArtemisSpec `json:"artemis,omitempty"`
	// +optional
	Kafka KafkaSpec `json:"kafka,omitempty"`
	// +optional
	HPA HPAConfig `json:"hpa,omitempty"`
}

type ProviderValidatorBaseSpec struct {
	// +kubebuilder:validation:Required
	Keycloak ProviderValidatorKeycloakSpec `json:"keycloak"`
	// +kubebuilder:validation:Required
	ProviderAPIURLs string `json:"providerAPIURLs"`
	// +optional
	MariaDB ProviderValidatorMariaDBSpec `json:"mariadb,omitempty"`
	// +kubebuilder:validation:Required
	AMQP ProviderValidatorAMQPSpec `json:"amqp"`
	// +optional
	MTLS ProviderValidatorMTLSSpec `json:"mtls,omitempty"`
	// +optional
	HPA HPAConfig `json:"hpa,omitempty"`
}
