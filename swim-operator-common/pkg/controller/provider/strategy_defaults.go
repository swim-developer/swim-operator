package provider

const (
	defaultPostgresEnvPort       = "5432"
	defaultQuarkusManagementPort = "8080"
	kafkaBootstrapLocalFmt       = "kafka-kafka-bootstrap.%s.svc.cluster.local:9092"
	kafkaPatternEnabled          = "true"
	kafkaPatternDisabled         = "false"
	defaultBrokerAdminCredential = "admin"
	amqpBrokerHDLSHostFmt        = "%s-hdls-svc.%s.svc.cluster.local"
	artemisJMXBrokerURLFmt       = "service:jmx:rmi:///jndi/rmi://%s-jmx-svc.%s.svc.cluster.local:1099/jmxrmi"
)
