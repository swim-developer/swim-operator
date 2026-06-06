package cv

import (
	"fmt"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/labels"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	corev1 "k8s.io/api/core/v1"
)

func BuildCVConfigMap(p CVBuildParams, managedBy string, amqpHost string) *corev1.ConfigMap {
	lbl := labels.StandardLabels(p.CRName, constants.ConsumerValidatorApp, p.CRName, managedBy)
	httpPort := resources.Int32Default(p.Spec.AppConfig.Quarkus.HTTPPort, 8080)
	sslPort := resources.Int32Default(p.Spec.AppConfig.Quarkus.SSLPort, 8443)
	logLevel := resources.StrDefault(p.Spec.AppConfig.Quarkus.LogLevel, "INFO")
	eventEnabled := resources.StrDefault(p.Spec.AppConfig.EventGenerator.Enabled, "true")
	eventSchedule := resources.StrDefault(p.Spec.AppConfig.EventGenerator.Schedule, "0 */1 * * * ?")
	eventsPath := resources.StrDefault(p.Spec.AppConfig.EventGenerator.EventsPath, "/opt/events")
	amqpPort := resources.Int32Default(p.Spec.AppConfig.Amqp.Port, 5672)
	mariadbPort := resources.Int32Default(p.Spec.MariaDB.Port, 3306)
	mariadbDatabase := resources.StrDefault(p.Spec.MariaDB.Database, p.DefaultDatabase)
	if p.Flavor == CVFlavorEd254 {
		exceptionsPath := resources.StrDefault(p.Spec.AppConfig.EventGenerator.ExceptionsPath, "/opt/exceptions")
		data := map[string]string{
			"QUARKUS_HTTP_PORT":                                  fmt.Sprintf("%d", httpPort),
			"QUARKUS_HTTP_SSL_PORT":                              fmt.Sprintf("%d", sslPort),
			"QUARKUS_HTTP_SSL_CERTIFICATE_FILES":                 "/certs/server/tls.crt",
			"QUARKUS_HTTP_SSL_CERTIFICATE_KEY_FILES":             "/certs/server/tls.key",
			"QUARKUS_HTTP_SSL_CERTIFICATE_TRUST_STORE_FILE":      "/certs/ca/ca.crt",
			"QUARKUS_HTTP_SSL_CERTIFICATE_TRUST_STORE_FILE_TYPE": "PEM",
			"QUARKUS_LOG_LEVEL":               logLevel,
			"EVENT_GENERATOR_ENABLED":         eventEnabled,
			"EVENT_GENERATOR_SCHEDULE":        eventSchedule,
			"EVENT_GENERATOR_EVENTS_PATH":     eventsPath,
			"EVENT_GENERATOR_EXCEPTIONS_PATH": exceptionsPath,
			"AMQP_BROKER_HOST":                amqpHost,
			"AMQP_BROKER_PORT":                fmt.Sprintf("%d", amqpPort),
			"MARIADB_HOST":                    MariaDBServiceName(p.CRName),
			"MARIADB_PORT":                    fmt.Sprintf("%d", mariadbPort),
			"MARIADB_DATABASE":                mariadbDatabase,
		}
		return resources.ConfigMap(fmt.Sprintf("%s-config", p.CRName), p.Namespace, lbl, data)
	}
	data := resources.BuildDnotamConsumerValidatorConfigMapData(resources.DnotamConsumerValidatorConfigMapParams{
		HTTPPort:        httpPort,
		SSLPort:         sslPort,
		LogLevel:        logLevel,
		EventEnabled:    eventEnabled,
		EventSchedule:   eventSchedule,
		EventsPath:      eventsPath,
		AmqpHost:        amqpHost,
		AmqpPort:        amqpPort,
		MariaDBHost:     MariaDBServiceName(p.CRName),
		MariaDBPort:     mariadbPort,
		MariaDBDatabase: mariadbDatabase,
	})
	return resources.ConfigMap(fmt.Sprintf("%s-config", p.CRName), p.Namespace, lbl, data)
}

func BuildCVServerCertificate(p CVBuildParams, managedBy string, host string) *certmanagerv1.Certificate {
	lbl := labels.StandardLabels(p.CRName, constants.ConsumerValidatorApp, p.CRName, managedBy)
	return resources.BuildServerCertificate(p.CRName, p.Namespace, p.CRName, lbl, p.Spec.CertManager.IssuerName, p.Spec.CertManager.IssuerKind, host)
}

func BuildCVClientCertificate(p CVBuildParams, managedBy string) *certmanagerv1.Certificate {
	lbl := labels.StandardLabels(p.CRName, constants.ConsumerValidatorApp, p.CRName, managedBy)
	return resources.BuildClientCertificate(p.CRName, p.Namespace, lbl, p.Spec.CertManager.IssuerName, p.Spec.CertManager.IssuerKind)
}

func BuildCVMTLSCertificate(p CVBuildParams, managedBy string) *certmanagerv1.Certificate {
	lbl := labels.StandardLabels(p.CRName, constants.ConsumerValidatorApp, p.CRName, managedBy)
	ks := fmt.Sprintf("%s-artemis-keystore-password", p.CRName)
	return resources.BuildMTLSCertificate(p.CRName, p.Namespace, lbl, p.Spec.CertManager.IssuerName, p.Spec.CertManager.IssuerKind, ks)
}
