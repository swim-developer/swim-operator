package provider

import (
	"fmt"
	"strings"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/labels"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultSwimEd254DB        = "swim-ed254"
	ed254KafkaTopicAll        = "ed254-events-all-topic"
	ed254KafkaTopicArrivalSeq = "ed254-arrival-sequence-topic"
	ed254DlqTopic             = "ed254-dlq-topic"
	ed254AmqpPortStr          = "5672"
	ed254AddressBPSecretFmt   = "%s-ed254-address-bp"
	ed254SecurityBPSecretFmt  = "%s-ed254-security-bp"
	defaultLogLevelINFO       = "INFO"
)

type Ed254ProviderStrategy struct {
	Payload Ed254ProviderPayload
}

func (s Ed254ProviderStrategy) CRKind() string { return "SwimEd254Provider" }

func (s Ed254ProviderStrategy) ArtemisBrokerCleanupPrefix() string { return "ed254" }

func (s Ed254ProviderStrategy) ArtemisSpecName() string { return s.Payload.Artemis.Name }

func (s Ed254ProviderStrategy) ServiceMonitorEnabled() bool {
	return s.Payload.Provider.Observability.ServiceMonitorEnabled
}

func (s Ed254ProviderStrategy) ReadyMessage() string { return "ED-254 Provider app is ready" }

func (s Ed254ProviderStrategy) NotReadyMessage() string { return "ED-254 Provider app not ready" }

func (s Ed254ProviderStrategy) AdditionalRoleRules() []rbacv1.PolicyRule { return nil }

func (s Ed254ProviderStrategy) Exposure() ProviderExposureSpec { return s.Payload.Exposure }

func (s Ed254ProviderStrategy) ConfigMapData(p ProviderBuildParams, clusterDomain string) map[string]string {
	cr := s.Payload
	artemisName := resources.DefaultArtemisName(cr.Artemis.Name, p.Name)
	postgresDatabase := resources.StrDefault(cr.Postgres.Database, defaultSwimEd254DB)
	kafkaBootstrapServers := fmt.Sprintf(kafkaBootstrapLocalFmt, p.Namespace)
	if p.Kafka.DeploymentMode == "external" && p.Kafka.BootstrapServers != "" {
		kafkaBootstrapServers = p.Kafka.BootstrapServers
	}
	aerodromes := ""
	if len(cr.Provider.Aerodromes) > 0 {
		aerodromes = strings.Join(cr.Provider.Aerodromes, ",")
	}
	heartbeatInterval := resources.Int32Default(cr.Provider.HeartbeatIntervalSeconds, 30)
	pgName := fmt.Sprintf(constants.PostgresSuffix, p.Name)
	return map[string]string{
		"POSTGRES_HOST":              resources.PostgresHostFQDN(pgName, p.Namespace),
		"POSTGRES_PORT":              defaultPostgresEnvPort,
		"POSTGRES_DB":                postgresDatabase,
		"AMQP_HOST":                  fmt.Sprintf(amqpBrokerHDLSHostFmt, artemisName, p.Namespace),
		"AMQP_PORT":                  ed254AmqpPortStr,
		"ARTEMIS_BROKER_NAME":        artemisName,
		"ARTEMIS_JMX_URL":            fmt.Sprintf(artemisJMXBrokerURLFmt, artemisName, p.Namespace),
		"KAFKA_BOOTSTRAP_SERVERS":    kafkaBootstrapServers,
		"KAFKA_TOPIC":                ed254KafkaTopicAll,
		"KAFKA_PATTERN":              kafkaPatternDisabled,
		"KAFKA_GROUP_ID":             p.Name,
		"AERODROMES":                 aerodromes,
		"HEARTBEAT_INTERVAL_SECONDS": fmt.Sprintf("%d", heartbeatInterval),
		"QUARKUS_HTTP_PORT":          defaultQuarkusManagementPort,
		"LOG_LEVEL":                  resources.StrDefault(cr.Provider.LogLevel, "INFO"),
		"OTEL_ENABLED":               fmt.Sprintf("%t", cr.Provider.Observability.OpenTelemetryEnabled),
		"OTEL_SDK_DISABLED":          fmt.Sprintf("%t", !cr.Provider.Observability.OpenTelemetryEnabled),
		"OTEL_ENDPOINT":              cr.Provider.Observability.OtelEndpoint,
		"OTEL_HEADERS":               resources.StrDefault(cr.Provider.Observability.OtelHeaders, ""),
		"PROMETHEUS_ENABLED":         fmt.Sprintf("%t", cr.Provider.Observability.PrometheusEnabled),
		"OPENAPI_SERVERS":            openapiServersValue(cr.Exposure, p.Name, p.Namespace, clusterDomain),
		"K8S_SECRET_ADDRESS":         fmt.Sprintf(ed254AddressBPSecretFmt, artemisName),
		"K8S_SECRET_SECURITY":        fmt.Sprintf(ed254SecurityBPSecretFmt, artemisName),
	}
}

func (s Ed254ProviderStrategy) AppSecretData() map[string]string {
	cr := s.Payload
	return map[string]string{
		"POSTGRES_USER":     resources.StrDefault(cr.Postgres.User, defaultSwimEd254DB),
		"POSTGRES_PASSWORD": resources.StrDefault(cr.Postgres.Password, defaultSwimEd254DB),
		"AMQP_USERNAME":     resources.StrDefault(cr.Artemis.AdminUser, defaultBrokerAdminCredential),
		"AMQP_PASSWORD":     resources.StrDefault(cr.Artemis.AdminPassword, defaultBrokerAdminCredential),
	}
}

func (s Ed254ProviderStrategy) OIDCSecretData() map[string]string {
	o := s.Payload.Provider.OIDC
	return map[string]string{
		"OIDC_AUTH_SERVER_URL": o.AuthServerUrl,
		"OIDC_CLIENT_ID":       o.ClientId,
		"OIDC_CLIENT_SECRET":   o.ClientSecret,
	}
}

func (s Ed254ProviderStrategy) AppImage() string {
	return resources.StrDefault(s.Payload.Provider.Image, "quay.io/masales/swim-ed254-provider:latest")
}

func (s Ed254ProviderStrategy) AppReplicas() int32 {
	return resources.Int32Default(s.Payload.Provider.Replicas, 1)
}

func (s Ed254ProviderStrategy) AppResources() corev1.ResourceRequirements {
	return resources.ResourcesOrDefault(s.Payload.Provider.Resources, "512Mi", "500m", "1Gi", "2")
}

func (s Ed254ProviderStrategy) PostgresParams(p ProviderBuildParams, managedBy string) resources.PostgresParams {
	cr := s.Payload
	name := fmt.Sprintf(constants.PostgresSuffix, p.Name)
	return resources.PostgresParams{
		Name:               name,
		Namespace:          p.Namespace,
		Labels:             labels.StandardLabels(name, "postgres", p.Name, managedBy),
		Image:              cr.Postgres.Image,
		StorageSize:        cr.Postgres.StorageSize,
		Database:           resources.StrDefault(cr.Postgres.Database, defaultSwimEd254DB),
		User:               resources.StrDefault(cr.Postgres.User, defaultSwimEd254DB),
		Password:           resources.StrDefault(cr.Postgres.Password, defaultSwimEd254DB),
		Resources:          cr.Postgres.Resources,
		ServiceAccountName: p.Name,
		SecretName:         fmt.Sprintf(constants.PostgresSecretSuffix, p.Name),
		PVCName:            fmt.Sprintf("%s-postgres-pvc", p.Name),
	}
}

func (s Ed254ProviderStrategy) ArtemisBaseParams(p ProviderBuildParams, ingressHost string) resources.ArtemisProviderParams {
	cr := s.Payload
	artemisName := resources.DefaultArtemisName(cr.Artemis.Name, p.Name)
	lbl := labels.StandardLabels(artemisName, "artemis", p.Name, "")
	consoleExpose := true
	if !cr.Artemis.Console.Expose {
		consoleExpose = false
	}
	consoleSSL := true
	if cr.Artemis.Console.SSLEnabled != nil {
		consoleSSL = *cr.Artemis.Console.SSLEnabled
	}
	return resources.ArtemisProviderParams{
		ArtemisName:      artemisName,
		Namespace:        p.Namespace,
		Labels:           lbl,
		AdminUser:        cr.Artemis.AdminUser,
		AdminPassword:    cr.Artemis.AdminPassword,
		KeystorePassword: cr.Artemis.KeystorePassword,
		IssuerName:       p.CertManager.IssuerName,
		IssuerKind:       p.CertManager.IssuerKind,
		IngressHost:      ingressHost,
		ExposeMode:       p.ArtemisExposeMode,
		Size:             cr.Artemis.Size,
		VerifyHost:       cr.Artemis.Acceptors.VerifyHost,
		AMQPSPort:        cr.Artemis.Acceptors.AMQPSPort,
		AMQPPort:         cr.Artemis.Acceptors.AMQPPort,
		BrokerProperties: cr.Artemis.BrokerProperties,
		StorageSize:      cr.Artemis.Storage.Size,
		StorageClassName: cr.Artemis.Storage.StorageClassName,
		ConsoleExpose:    consoleExpose,
		ConsoleSSL:       consoleSSL,
		ExtraMounts: []string{
			fmt.Sprintf(constants.SSOJAASConfigSuffix, artemisName),
			fmt.Sprintf(ed254AddressBPSecretFmt, artemisName),
			fmt.Sprintf(ed254SecurityBPSecretFmt, artemisName),
		},
		JMXEnabled: cr.Artemis.JMX.Enabled,
		JMXPort:    cr.Artemis.JMX.Port,
	}
}

func (s Ed254ProviderStrategy) ArtemisOIDCSecret(p ProviderBuildParams, managedBy string) *corev1.Secret {
	cr := s.Payload
	artemisName := resources.DefaultArtemisName(cr.Artemis.Name, p.Name)
	lbl := labels.StandardLabels(artemisName, "artemis", p.Name, managedBy)
	configPath := fmt.Sprintf(constants.SSOJAASConfigSuffix, artemisName)
	moduleBlock := `    org.keycloak.adapters.jaas.DirectAccessGrantsLoginModule required
        reload=true
        keycloak-config-file="/amq/extra/secrets/` + configPath + `/_keycloak-login-module.json"
        role-principal-class="org.apache.activemq.artemis.spi.core.security.jaas.RolePrincipal";`
	loginConfig := resources.BuildArtemisJaasLoginConfig(moduleBlock)
	keycloakConfig := fmt.Sprintf(`{
  "realm": "%s",
  "auth-server-url": "%s",
  "ssl-required": "external",
  "resource": "%s",
  "verify-token-audience": true,
  "credentials": {
    "secret": "%s"
  },
  "use-resource-role-mappings": true,
  "confidential-port": 0
}`, cr.Artemis.OIDC.Realm, cr.Artemis.OIDC.AuthServerUrl, cr.Artemis.OIDC.ClientId, cr.Artemis.OIDC.ClientSecret)
	return resources.BuildArtemisJaasConfigSecret(configPath, p.Namespace, lbl, loginConfig, keycloakConfig)
}

func (s Ed254ProviderStrategy) ArtemisAddressBPSecret(p ProviderBuildParams, managedBy string) *corev1.Secret {
	artemisName := resources.DefaultArtemisName(s.Payload.Artemis.Name, p.Name)
	lbl := labels.StandardLabels(artemisName, "artemis", p.Name, managedBy)
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf(ed254AddressBPSecretFmt, artemisName),
			Namespace: p.Namespace,
			Labels:    lbl,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"addressConfigurations.properties": "",
		},
	}
}

func (s Ed254ProviderStrategy) ArtemisSecurityBPSecret(p ProviderBuildParams, managedBy string) *corev1.Secret {
	artemisName := resources.DefaultArtemisName(s.Payload.Artemis.Name, p.Name)
	lbl := labels.StandardLabels(artemisName, "artemis", p.Name, managedBy)
	securityRoles := `securityRoles.#.admin.consume=true
securityRoles.#.admin.browse=true
securityRoles.#.admin.send=true
securityRoles.#.admin.manage=true
securityRoles.#.admin.createAddress=true
securityRoles.#.admin.deleteAddress=true
securityRoles.#.admin.createDurableQueue=true
securityRoles.#.admin.deleteDurableQueue=true
securityRoles.#.admin.createNonDurableQueue=true
securityRoles.#.admin.deleteNonDurableQueue=true
`
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf(ed254SecurityBPSecretFmt, artemisName),
			Namespace: p.Namespace,
			Labels:    lbl,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"securityRoles.properties": securityRoles,
		},
	}
}

func (s Ed254ProviderStrategy) KafkaTopicAllName() string { return ed254KafkaTopicArrivalSeq }

func (s Ed254ProviderStrategy) KafkaTopicDLQName() string { return ed254DlqTopic }

func (s Ed254ProviderStrategy) KafkaTopicPartitions(topicName string) int64 {
	if topicName == ed254KafkaTopicArrivalSeq {
		return 10
	}
	return 3
}

func (s Ed254ProviderStrategy) SkipIfKafkaExists() bool { return false }

var _ ProviderStrategy = Ed254ProviderStrategy{}
