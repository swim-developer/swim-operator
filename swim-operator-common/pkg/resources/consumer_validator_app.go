package resources

import "fmt"

type DnotamConsumerValidatorConfigMapParams struct {
	HTTPPort         int32
	SSLPort          int32
	LogLevel         string
	EventEnabled     string
	EventSchedule    string
	EventsPath       string
	AmqpHost         string
	AmqpPort         int32
	MariaDBHost      string
	MariaDBPort      int32
	MariaDBDatabase  string
}

func BuildDnotamConsumerValidatorConfigMapData(p DnotamConsumerValidatorConfigMapParams) map[string]string {
	return map[string]string{
		"QUARKUS_HTTP_PORT":                                  fmt.Sprintf("%d", p.HTTPPort),
		"QUARKUS_HTTP_SSL_PORT":                              fmt.Sprintf("%d", p.SSLPort),
		"QUARKUS_HTTP_SSL_CERTIFICATE_FILES":                 "/certs/server/tls.crt",
		"QUARKUS_HTTP_SSL_CERTIFICATE_KEY_FILES":             "/certs/server/tls.key",
		"QUARKUS_HTTP_SSL_CERTIFICATE_TRUST_STORE_FILE":      "/certs/ca/ca.crt",
		"QUARKUS_HTTP_SSL_CERTIFICATE_TRUST_STORE_FILE_TYPE": "PEM",
		"QUARKUS_LOG_LEVEL":                                  p.LogLevel,
		"EVENT_GENERATOR_ENABLED":                            p.EventEnabled,
		"EVENT_GENERATOR_SCHEDULE":                           p.EventSchedule,
		"EVENT_GENERATOR_EVENTS_PATH":                        p.EventsPath,
		"AMQP_BROKER_HOST":                                   p.AmqpHost,
		"AMQP_BROKER_PORT":                                   fmt.Sprintf("%d", p.AmqpPort),
		"MARIADB_HOST":                                       p.MariaDBHost,
		"MARIADB_PORT":                                       fmt.Sprintf("%d", p.MariaDBPort),
		"MARIADB_DATABASE":                                   p.MariaDBDatabase,
	}
}
