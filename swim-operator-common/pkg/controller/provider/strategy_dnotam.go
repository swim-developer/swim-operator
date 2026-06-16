package provider

import (
	"fmt"
	"strings"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/labels"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

const (
	dnotamKafkaTopicAll       = "dnotam-events-all-topic"
	dnotamKafkaTopicPattern   = "dnotam-events-(?!dlq).*-topic"
	dnotamAddressBPSecretFmt  = "%s-dnotam-address-bp"
	dnotamSecurityBPSecretFmt = "%s-dnotam-security-bp"
)

type DnotamProviderStrategy struct {
	Payload DnotamProviderPayload
}

func (s DnotamProviderStrategy) CRKind() string { return "SwimDigitalNotamProvider" }

func (s DnotamProviderStrategy) ArtemisBrokerCleanupPrefix() string { return "dnotam" }

func (s DnotamProviderStrategy) ArtemisSpecName() string { return s.Payload.Artemis.Name }

func (s DnotamProviderStrategy) ServiceMonitorEnabled() bool {
	return s.Payload.Provider.Observability.ServiceMonitorEnabled
}

func (s DnotamProviderStrategy) ReadyMessage() string { return "Provider app is ready" }

func (s DnotamProviderStrategy) NotReadyMessage() string { return "Provider app not ready" }

func (s DnotamProviderStrategy) AdditionalRoleRules() []rbacv1.PolicyRule {
	return []rbacv1.PolicyRule{
		{
			APIGroups: []string{"coordination.k8s.io"},
			Resources: []string{"leases"},
			Verbs:     []string{"get", "list", "watch", "create", "update", "patch"},
		},
	}
}

func (s DnotamProviderStrategy) Exposure() ProviderExposureSpec { return s.Payload.Exposure }

func (s DnotamProviderStrategy) ConfigMapData(p ProviderBuildParams, clusterDomain string) map[string]string {
	cr := s.Payload
	artemisName := resources.DefaultArtemisName(cr.Artemis.Name, p.Name)
	postgresDatabase := resources.StrDefault(cr.Postgres.Database, "swim-dnotam")
	kafkaBootstrapServers := fmt.Sprintf(kafkaBootstrapLocalFmt, p.Namespace)
	if p.Kafka.DeploymentMode == "external" && p.Kafka.BootstrapServers != "" {
		kafkaBootstrapServers = p.Kafka.BootstrapServers
	}
	kafkaTopic := dnotamKafkaTopicAll
	kafkaPattern := kafkaPatternDisabled
	if cr.Provider.ConsumeFromClientTopics {
		kafkaTopic = dnotamKafkaTopicPattern
		kafkaPattern = kafkaPatternEnabled
	}
	amqpPort := resources.Int32Default(cr.Artemis.Acceptors.AMQPSPort, 5671)
	pgName := fmt.Sprintf(constants.PostgresSuffix, p.Name)
	return map[string]string{
		"POSTGRES_HOST":           resources.PostgresHostFQDN(pgName, p.Namespace),
		"POSTGRES_PORT":           defaultPostgresEnvPort,
		"POSTGRES_DB":             postgresDatabase,
		"AMQP_HOST":               fmt.Sprintf(amqpBrokerHDLSHostFmt, artemisName, p.Namespace),
		"AMQP_PORT":               fmt.Sprintf("%d", amqpPort),
		"ARTEMIS_BROKER_NAME":     artemisName,
		"ARTEMIS_JMX_URL":         fmt.Sprintf(artemisJMXBrokerURLFmt, artemisName, p.Namespace),
		"KAFKA_BOOTSTRAP_SERVERS": kafkaBootstrapServers,
		"KAFKA_TOPIC":             kafkaTopic,
		"KAFKA_PATTERN":           kafkaPattern,
		"KAFKA_GROUP_ID":          p.Name,
		"QUARKUS_HTTP_PORT":       defaultQuarkusManagementPort,
		"LOG_LEVEL":               cr.Provider.LogLevel,
		"OTEL_ENABLED":            fmt.Sprintf("%t", cr.Provider.Observability.OpenTelemetryEnabled),
		"OTEL_SDK_DISABLED":       fmt.Sprintf("%t", !cr.Provider.Observability.OpenTelemetryEnabled),
		"OTEL_ENDPOINT":           cr.Provider.Observability.OtelEndpoint,
		"OTEL_HEADERS":            resources.StrDefault(cr.Provider.Observability.OtelHeaders, ""),
		"PROMETHEUS_ENABLED":      fmt.Sprintf("%t", cr.Provider.Observability.PrometheusEnabled),
		"OPENAPI_SERVERS":         openapiServersValue(cr.Exposure, p.Name, p.Namespace, clusterDomain),
		"K8S_SECRET_ADDRESS":      fmt.Sprintf(dnotamAddressBPSecretFmt, artemisName),
		"K8S_SECRET_SECURITY":     fmt.Sprintf(dnotamSecurityBPSecretFmt, artemisName),
	}
}

func (s DnotamProviderStrategy) AppSecretData() map[string]string {
	cr := s.Payload
	return map[string]string{
		"POSTGRES_USER":     resources.StrDefault(cr.Postgres.User, "swim-provider"),
		"POSTGRES_PASSWORD": resources.StrDefault(cr.Postgres.Password, "swim-provider"),
		"AMQP_USERNAME":     resources.StrDefault(cr.Artemis.AdminUser, defaultBrokerAdminCredential),
		"AMQP_PASSWORD":     resources.StrDefault(cr.Artemis.AdminPassword, defaultBrokerAdminCredential),
	}
}

func (s DnotamProviderStrategy) OIDCSecretData() map[string]string {
	o := s.Payload.Provider.OIDC
	return map[string]string{
		"OIDC_AUTH_SERVER_URL": o.AuthServerUrl,
		"OIDC_CLIENT_ID":       o.ClientId,
		"OIDC_CLIENT_SECRET":   o.ClientSecret,
	}
}

func (s DnotamProviderStrategy) AppImage() string {
	return resources.StrDefault(s.Payload.Provider.Image, "quay.io/masales/swim-dnotam-provider:latest")
}

func (s DnotamProviderStrategy) AppReplicas() int32 {
	return resources.Int32Default(s.Payload.Provider.Replicas, 1)
}

func (s DnotamProviderStrategy) AppResources() corev1.ResourceRequirements {
	return s.Payload.Provider.Resources
}

func (s DnotamProviderStrategy) PostgresParams(p ProviderBuildParams, managedBy string) resources.PostgresParams {
	cr := s.Payload
	name := fmt.Sprintf(constants.PostgresSuffix, p.Name)
	return resources.PostgresParams{
		Name:               name,
		Namespace:          p.Namespace,
		Labels:             labels.StandardLabels(name, "postgres", p.Name, managedBy),
		Image:              cr.Postgres.Image,
		StorageSize:        cr.Postgres.StorageSize,
		Database:           cr.Postgres.Database,
		User:               cr.Postgres.User,
		Password:           cr.Postgres.Password,
		Resources:          cr.Postgres.Resources,
		ServiceAccountName: p.Name,
		SecretName:         fmt.Sprintf(constants.PostgresSecretSuffix, p.Name),
		PVCName:            fmt.Sprintf("%s-postgres-pvc", p.Name),
	}
}

func (s DnotamProviderStrategy) ArtemisBaseParams(p ProviderBuildParams, ingressHost string) resources.ArtemisProviderParams {
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
			fmt.Sprintf(dnotamAddressBPSecretFmt, artemisName),
			fmt.Sprintf(dnotamSecurityBPSecretFmt, artemisName),
		},
		JMXEnabled: cr.Artemis.JMX.Enabled,
		JMXPort:    cr.Artemis.JMX.Port,
	}
}

func (s DnotamProviderStrategy) ArtemisOIDCSecret(p ProviderBuildParams, managedBy string) *corev1.Secret {
	cr := s.Payload
	artemisName := resources.DefaultArtemisName(cr.Artemis.Name, p.Name)
	lbl := labels.StandardLabels(artemisName, "artemis", p.Name, managedBy)
	ap := resources.ArtemisProviderParams{
		ArtemisName:       artemisName,
		Namespace:         p.Namespace,
		Labels:            lbl,
		OIDCRealm:         cr.Artemis.OIDC.Realm,
		OIDCAuthServerURL: cr.Artemis.OIDC.AuthServerUrl,
		OIDCClientId:      cr.Artemis.OIDC.ClientId,
		OIDCClientSecret:  cr.Artemis.OIDC.ClientSecret,
	}
	return resources.BuildProviderArtemisOIDCSecret(ap)
}

func (s DnotamProviderStrategy) ArtemisAddressBPSecret(p ProviderBuildParams, managedBy string) *corev1.Secret {
	artemisName := resources.DefaultArtemisName(s.Payload.Artemis.Name, p.Name)
	lbl := labels.StandardLabels(artemisName, "artemis", p.Name, managedBy)
	return resources.BuildProviderArtemisAddressBPSecret(artemisName, p.Namespace, lbl)
}

func (s DnotamProviderStrategy) ArtemisSecurityBPSecret(p ProviderBuildParams, managedBy string) *corev1.Secret {
	artemisName := resources.DefaultArtemisName(s.Payload.Artemis.Name, p.Name)
	lbl := labels.StandardLabels(artemisName, "artemis", p.Name, managedBy)
	return resources.BuildProviderArtemisSecurityBPSecret(artemisName, p.Namespace, lbl)
}

func (s DnotamProviderStrategy) KafkaTopicAllName() string { return dnotamKafkaTopicAll }

func (s DnotamProviderStrategy) KafkaTopicDLQName() string { return "dnotam-events-dlq-topic" }

func (s DnotamProviderStrategy) KafkaTopicPartitions(topicName string) int64 {
	if topicName == dnotamKafkaTopicAll {
		return 10
	}
	return 3
}

func (s DnotamProviderStrategy) SkipIfKafkaExists() bool { return true }

var _ ProviderStrategy = DnotamProviderStrategy{}

// openapiServersValue builds the OPENAPI_SERVERS value from the exposure spec.
func openapiServersValue(ex ProviderExposureSpec, name, ns, clusterDomain string) string {
	var hosts []string
	if ex.HTTPEdgeEnabled {
		h := ex.HTTPSEdgeHost
		if h == "" {
			h = fmt.Sprintf(constants.HostnameSuffix, name, ns, clusterDomain)
		}
		hosts = append(hosts, fmt.Sprintf("https://%s", h))
	}
	if ex.HTTPSPassthroughEnabled {
		h := ex.HTTPSPassthroughHost
		if h == "" {
			h = fmt.Sprintf(constants.MTLSHostnameSuffix, name, ns, clusterDomain)
		}
		hosts = append(hosts, fmt.Sprintf("https://%s", h))
	}
	if len(hosts) == 0 {
		return ""
	}
	return strings.Join(hosts, ",")
}
