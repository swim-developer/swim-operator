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
	defaultSwimFficeDB       = "swim-ffice"
	fficeKafkaTopicAll       = "ffice-events-all-topic"
	fficeKafkaTopicEvents    = "ffice-events-topic"
	fficeDlqTopic            = "ffice-events-dlq-topic"
	fficeAmqpPortStr         = "5672"
	fficeAddressBPSecretFmt  = "%s-ffice-address-bp"
	fficeSecurityBPSecretFmt = "%s-ffice-security-bp"
)

type FficeProviderStrategy struct {
	Payload FficeProviderPayload
}

func (s FficeProviderStrategy) CRKind() string { return "SwimFficeProvider" }

func (s FficeProviderStrategy) ArtemisBrokerCleanupPrefix() string { return "ffice" }

func (s FficeProviderStrategy) ArtemisSpecName() string { return s.Payload.Artemis.Name }

func (s FficeProviderStrategy) ServiceMonitorEnabled() bool {
	return s.Payload.Provider.Observability.ServiceMonitorEnabled
}

func (s FficeProviderStrategy) ReadyMessage() string { return "FF-ICE Provider app is ready" }

func (s FficeProviderStrategy) NotReadyMessage() string { return "FF-ICE Provider app not ready" }

func (s FficeProviderStrategy) AdditionalRoleRules() []rbacv1.PolicyRule { return nil }

func (s FficeProviderStrategy) Exposure() ProviderExposureSpec { return s.Payload.Exposure }

func (s FficeProviderStrategy) ConfigMapData(p ProviderBuildParams, clusterDomain string) map[string]string {
	cr := s.Payload
	artemisName := resources.DefaultArtemisName(cr.Artemis.Name, p.Name)
	postgresDatabase := resources.StrDefault(cr.Postgres.Database, defaultSwimFficeDB)
	kafkaBootstrapServers := fmt.Sprintf(kafkaBootstrapLocalFmt, p.Namespace)
	if p.Kafka.DeploymentMode == "external" && p.Kafka.BootstrapServers != "" {
		kafkaBootstrapServers = p.Kafka.BootstrapServers
	}
	swimTopics := "FficeService"
	if len(cr.Provider.SwimTopics) > 0 {
		swimTopics = strings.Join(cr.Provider.SwimTopics, ",")
	}
	heartbeatInterval := resources.Int32Default(cr.Provider.HeartbeatIntervalSeconds, 15)
	pgName := fmt.Sprintf(constants.PostgresSuffix, p.Name)
	return map[string]string{
		"POSTGRES_HOST":              resources.PostgresHostFQDN(pgName, p.Namespace),
		"POSTGRES_PORT":              defaultPostgresEnvPort,
		"POSTGRES_DB":                postgresDatabase,
		"AMQP_HOST":                  fmt.Sprintf(amqpBrokerHDLSHostFmt, artemisName, p.Namespace),
		"AMQP_PORT":                  fficeAmqpPortStr,
		"ARTEMIS_BROKER_NAME":        artemisName,
		"ARTEMIS_JMX_URL":            fmt.Sprintf(artemisJMXBrokerURLFmt, artemisName, p.Namespace),
		"KAFKA_BOOTSTRAP_SERVERS":    kafkaBootstrapServers,
		"KAFKA_TOPIC":                fficeKafkaTopicAll,
		"KAFKA_PATTERN":              kafkaPatternDisabled,
		"KAFKA_GROUP_ID":             p.Name,
		"SWIM_TOPICS":                swimTopics,
		"HEARTBEAT_INTERVAL_SECONDS": fmt.Sprintf("%d", heartbeatInterval),
		"QUARKUS_HTTP_PORT":          defaultQuarkusManagementPort,
		"LOG_LEVEL":                  resources.StrDefault(cr.Provider.LogLevel, defaultLogLevelINFO),
		"OTEL_ENABLED":               fmt.Sprintf("%t", cr.Provider.Observability.OpenTelemetryEnabled),
		"OTEL_SDK_DISABLED":          fmt.Sprintf("%t", !cr.Provider.Observability.OpenTelemetryEnabled),
		"OTEL_ENDPOINT":              cr.Provider.Observability.OtelEndpoint,
		"OTEL_HEADERS":               resources.StrDefault(cr.Provider.Observability.OtelHeaders, ""),
		"PROMETHEUS_ENABLED":         fmt.Sprintf("%t", cr.Provider.Observability.PrometheusEnabled),
		"OPENAPI_SERVERS":            openapiServersValue(cr.Exposure, p.Name, p.Namespace, clusterDomain),
		"K8S_SECRET_ADDRESS":         fmt.Sprintf(fficeAddressBPSecretFmt, artemisName),
		"K8S_SECRET_SECURITY":        fmt.Sprintf(fficeSecurityBPSecretFmt, artemisName),
	}
}

func (s FficeProviderStrategy) AppSecretData() map[string]string {
	cr := s.Payload
	return map[string]string{
		"POSTGRES_USER":     resources.StrDefault(cr.Postgres.User, defaultSwimFficeDB),
		"POSTGRES_PASSWORD": resources.StrDefault(cr.Postgres.Password, defaultSwimFficeDB),
		"AMQP_USERNAME":     resources.StrDefault(cr.Artemis.AdminUser, defaultBrokerAdminCredential),
		"AMQP_PASSWORD":     resources.StrDefault(cr.Artemis.AdminPassword, defaultBrokerAdminCredential),
	}
}

func (s FficeProviderStrategy) OIDCSecretData() map[string]string {
	o := s.Payload.Provider.OIDC
	return map[string]string{
		"OIDC_AUTH_SERVER_URL": o.AuthServerUrl,
		"OIDC_CLIENT_ID":       o.ClientId,
		"OIDC_CLIENT_SECRET":   o.ClientSecret,
	}
}

func (s FficeProviderStrategy) AppImage() string {
	return resources.StrDefault(s.Payload.Provider.Image, "quay.io/masales/swim-ffice-provider:latest")
}

func (s FficeProviderStrategy) AppReplicas() int32 {
	return resources.Int32Default(s.Payload.Provider.Replicas, 1)
}

func (s FficeProviderStrategy) AppResources() corev1.ResourceRequirements {
	return resources.ResourcesOrDefault(s.Payload.Provider.Resources, "512Mi", "500m", "1Gi", "2")
}

func (s FficeProviderStrategy) PostgresParams(p ProviderBuildParams, managedBy string) resources.PostgresParams {
	cr := s.Payload
	name := fmt.Sprintf(constants.PostgresSuffix, p.Name)
	return resources.PostgresParams{
		Name:               name,
		Namespace:          p.Namespace,
		Labels:             labels.StandardLabels(name, "postgres", p.Name, managedBy),
		Image:              cr.Postgres.Image,
		StorageSize:        cr.Postgres.StorageSize,
		Database:           resources.StrDefault(cr.Postgres.Database, defaultSwimFficeDB),
		User:               resources.StrDefault(cr.Postgres.User, defaultSwimFficeDB),
		Password:           resources.StrDefault(cr.Postgres.Password, defaultSwimFficeDB),
		Resources:          cr.Postgres.Resources,
		ServiceAccountName: p.Name,
		SecretName:         fmt.Sprintf(constants.PostgresSecretSuffix, p.Name),
		PVCName:            fmt.Sprintf("%s-postgres-pvc", p.Name),
	}
}

func (s FficeProviderStrategy) ArtemisBaseParams(p ProviderBuildParams, ingressHost string) resources.ArtemisProviderParams {
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
			fmt.Sprintf(fficeAddressBPSecretFmt, artemisName),
			fmt.Sprintf(fficeSecurityBPSecretFmt, artemisName),
		},
		JMXEnabled: cr.Artemis.JMX.Enabled,
		JMXPort:    cr.Artemis.JMX.Port,
	}
}

func (s FficeProviderStrategy) ArtemisOIDCSecret(p ProviderBuildParams, managedBy string) *corev1.Secret {
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

func (s FficeProviderStrategy) ArtemisAddressBPSecret(p ProviderBuildParams, managedBy string) *corev1.Secret {
	artemisName := resources.DefaultArtemisName(s.Payload.Artemis.Name, p.Name)
	lbl := labels.StandardLabels(artemisName, "artemis", p.Name, managedBy)
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf(fficeAddressBPSecretFmt, artemisName),
			Namespace: p.Namespace,
			Labels:    lbl,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"addressConfigurations.properties": "",
		},
	}
}

func (s FficeProviderStrategy) ArtemisSecurityBPSecret(p ProviderBuildParams, managedBy string) *corev1.Secret {
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
			Name:      fmt.Sprintf(fficeSecurityBPSecretFmt, artemisName),
			Namespace: p.Namespace,
			Labels:    lbl,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"securityRoles.properties": securityRoles,
		},
	}
}

func (s FficeProviderStrategy) KafkaTopicAllName() string { return fficeKafkaTopicEvents }

func (s FficeProviderStrategy) KafkaTopicDLQName() string { return fficeDlqTopic }

func (s FficeProviderStrategy) KafkaTopicPartitions(topicName string) int64 {
	if topicName == fficeKafkaTopicEvents {
		return 10
	}
	return 3
}

func (s FficeProviderStrategy) SkipIfKafkaExists() bool { return false }

var _ ProviderStrategy = FficeProviderStrategy{}
