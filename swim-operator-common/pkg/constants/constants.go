package constants

const (
	SharedManagedByLabel = "apps.swim-developer.github.io/managed-by"

	DnotamConsumerFinalizerName         = "apps.swim-developer.github.io/finalizer"
	Ed254ConsumerFinalizerName          = "apps.swim-developer.github.io/ed254-consumer-finalizer"
	ProviderFinalizerName               = "apps.swim-developer.github.io/provider-finalizer"
	Ed254ProviderFinalizerName          = "apps.swim-developer.github.io/ed254-provider-finalizer"
	ConsumerValidatorFinalizerName      = "apps.swim-developer.github.io/consumer-validator-finalizer"
	Ed254ConsumerValidatorFinalizerName = "apps.swim-developer.github.io/ed254-consumer-validator-finalizer"

	FficeConsumerFinalizerName          = "apps.swim-developer.github.io/ffice-consumer-finalizer"
	FficeProviderFinalizerName          = "apps.swim-developer.github.io/ffice-provider-finalizer"
	FficeConsumerValidatorFinalizerName = "apps.swim-developer.github.io/ffice-consumer-validator-finalizer"
	FficeProviderValidatorFinalizerName = "apps.swim-developer.github.io/ffice-provider-validator-finalizer"
	Ed254ProviderValidatorFinalizerName = "apps.swim-developer.github.io/ed254-provider-validator-finalizer"

	KafkaAPIVersion = "kafka.strimzi.io/v1beta2"
	KafkaGroup      = "kafka.strimzi.io"

	ErrUnableToCreateController = "unable to create controller"
	ErrFailedToWriteOutput      = "failed to write to output: %w"
	InfoFoundActiveSWIMCR       = "Found active SWIM CR"

	ArtemisSuffix            = "%s-artemis"
	KeystorePasswordSuffix   = "%s-keystore-password"
	MongoDBCredentialsSuffix = "%s-mongodb-credentials"
	MongoDBSuffix            = "%s-mongodb"
	MongoDBDataSuffix        = "%s-mongodb-data"
	ServerTLSSuffix          = "%s-server-tls"
	HostnameSuffix           = "%s-%s.%s"
	MTLSHostnameSuffix       = "%s-mtls-%s.%s"
	SSOJAASConfigSuffix      = "%s-sso-jaas-config"
	PostgresSuffix           = "%s-postgres"
	PostgresSecretSuffix     = "%s-postgres-secret"
	SSLSecretSuffix          = "%s-ssl-secret"

	DatabasePasswordKey = "database-password"
	DatabaseUserKey     = "database-user"

	ConsumerValidatorApp = "consumer-validator"
	ProviderValidatorApp = "provider-validator"

	MTLSCertsVolume      = "mtls-certs"
	CurlMetricsContainer = "curl-metrics"
)
