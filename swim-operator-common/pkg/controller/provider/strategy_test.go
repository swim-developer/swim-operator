package provider

import (
	"testing"

	. "github.com/onsi/gomega"
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
)

const (
	testOIDCAuthServerURL    = "https://keycloak.example.com/realms/swim"
	testRealmSwim            = "swim"
	testAmqBrokerClientID    = "amq-broker"
	testArtemisName          = "my-artemis"
	testSwimProviderCred     = "swim-provider"
	testLabelKubeManagedBy   = "app.kubernetes.io/managed-by"
	testDnotamEventsAllTopic = "dnotam-events-all-topic"
	testAdminPassword        = "admin"
	testClientSecret         = "secret123"
	testNamespaceSwimDemo    = "swim-demo"
	testClusterIssuerName    = "swim-ca-issuer"
	testClusterIssuerKind    = "ClusterIssuer"
	testProviderName         = "my-provider"
	testDnotamOperatorMgr    = "dnotam-operator"
	testEd254OperatorMgr     = "ed254-operator"
	testProviderSecret       = "provider-secret"
	testAppsExampleDomain    = "apps.example.com"
	testMyAppName            = "my-app"
	testKeyMissingFmt        = "missing key: %s"
	testEd254Aerodrome1      = "EGLL"
	testEd254Aerodrome2      = "LFPG"
	testEd254Aerodrome3      = "EDDF"
	testSwimDnotamDatabase   = "swim-dnotam"
	testSwimEd254Credential  = "swim-ed254"
)

func newDnotamStrategy() DnotamProviderStrategy {
	return DnotamProviderStrategy{
		Payload: DnotamProviderPayload{
			Postgres: commonapi.ProviderPostgresSpec{
				Database: testSwimDnotamDatabase,
				User:     testSwimProviderCred,
				Password: testSwimProviderCred,
			},
			Artemis: commonapi.ProviderArtemisSpec{
				Name:          testArtemisName,
				AdminUser:     testAdminPassword,
				AdminPassword: testAdminPassword,
				OIDC: commonapi.ArtemisOIDCSpec{
					AuthServerUrl: testOIDCAuthServerURL,
					Realm:         testRealmSwim,
					ClientId:      testAmqBrokerClientID,
					ClientSecret:  testClientSecret,
				},
			},
			Provider: commonapi.ProviderAppBaseSpec{
				Image:    "quay.io/masales/swim-dnotam-provider:1.0",
				Replicas: 2,
				LogLevel: "DEBUG",
				OIDC: commonapi.ProviderOIDCSpec{
					AuthServerUrl: testOIDCAuthServerURL,
					ClientId:      "dnotam-provider",
					ClientSecret:  testProviderSecret,
				},
			},
		},
	}
}

func newEd254Strategy() Ed254ProviderStrategy {
	return Ed254ProviderStrategy{
		Payload: Ed254ProviderPayload{
			Postgres: commonapi.Ed254PostgresSpec{
				Database: testSwimEd254Credential,
				User:     testSwimEd254Credential,
				Password: testSwimEd254Credential,
			},
			Artemis: commonapi.Ed254ArtemisSpec{
				Name:          testArtemisName,
				AdminUser:     testAdminPassword,
				AdminPassword: testAdminPassword,
				OIDC: commonapi.ArtemisOIDCSpec{
					AuthServerUrl: testOIDCAuthServerURL,
					Realm:         testRealmSwim,
					ClientId:      testAmqBrokerClientID,
					ClientSecret:  testClientSecret,
				},
			},
			Provider: commonapi.Ed254ProviderAppBaseSpec{
				Image:                    "quay.io/masales/swim-ed254-provider:1.0",
				Replicas:                 3,
				Aerodromes:               []string{testEd254Aerodrome1, testEd254Aerodrome2, testEd254Aerodrome3},
				HeartbeatIntervalSeconds: 15,
				OIDC: commonapi.ProviderOIDCSpec{
					AuthServerUrl: testOIDCAuthServerURL,
					ClientId:      "ed254-provider",
					ClientSecret:  testProviderSecret,
				},
			},
		},
	}
}

func defaultBuildParams() ProviderBuildParams {
	return ProviderBuildParams{
		Name:      testProviderName,
		Namespace: testNamespaceSwimDemo,
		Kafka: commonapi.KafkaSpec{
			DeploymentMode: "managed",
		},
		CertManager: commonapi.CertManagerSpec{
			IssuerName: testClusterIssuerName,
			IssuerKind: testClusterIssuerKind,
		},
	}
}

// --- DNOTAM strategy tests ---

func TestDnotamStrategy_CRKind(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	g.Expect(s.CRKind()).To(Equal("SwimDigitalNotamProvider"))
}

func TestDnotamStrategy_ArtemisBrokerCleanupPrefix(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	g.Expect(s.ArtemisBrokerCleanupPrefix()).To(Equal("dnotam"))
}

func TestDnotamStrategy_ArtemisSpecName(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	g.Expect(s.ArtemisSpecName()).To(Equal(testArtemisName))
}

func TestDnotamStrategy_AppImage(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	g.Expect(s.AppImage()).To(Equal("quay.io/masales/swim-dnotam-provider:1.0"))
}

func TestDnotamStrategy_AppImage_Default(t *testing.T) {
	g := NewWithT(t)
	s := DnotamProviderStrategy{}
	g.Expect(s.AppImage()).To(Equal("quay.io/masales/swim-dnotam-provider:latest"))
}

func TestDnotamStrategy_AppReplicas(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	g.Expect(s.AppReplicas()).To(Equal(int32(2)))
}

func TestDnotamStrategy_AppReplicas_Default(t *testing.T) {
	g := NewWithT(t)
	s := DnotamProviderStrategy{}
	g.Expect(s.AppReplicas()).To(Equal(int32(1)))
}

func TestDnotamStrategy_AdditionalRoleRules_HasLeaseRule(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	rules := s.AdditionalRoleRules()
	g.Expect(rules).To(HaveLen(1))
	g.Expect(rules[0].APIGroups).To(ContainElement("coordination.k8s.io"))
	g.Expect(rules[0].Resources).To(ContainElement("leases"))
}

func TestDnotamStrategy_ConfigMapData_ContainsAllRequiredKeys(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	p := defaultBuildParams()
	data := s.ConfigMapData(p, testAppsExampleDomain)

	requiredKeys := []string{
		"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_DB",
		"AMQP_HOST", "AMQP_PORT",
		"ARTEMIS_BROKER_NAME", "ARTEMIS_JMX_URL",
		"KAFKA_BOOTSTRAP_SERVERS", "KAFKA_TOPIC", "KAFKA_PATTERN", "KAFKA_GROUP_ID",
		"QUARKUS_HTTP_PORT", "LOG_LEVEL",
		"OTEL_ENABLED", "OTEL_SDK_DISABLED", "OTEL_ENDPOINT", "OTEL_HEADERS",
		"PROMETHEUS_ENABLED", "OPENAPI_SERVERS",
		"K8S_SECRET_ADDRESS", "K8S_SECRET_SECURITY",
	}
	for _, key := range requiredKeys {
		g.Expect(data).To(HaveKey(key), testKeyMissingFmt, key)
	}
}

func TestDnotamStrategy_ConfigMapData_PostgresHost(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	p := defaultBuildParams()
	data := s.ConfigMapData(p, "")
	g.Expect(data["POSTGRES_HOST"]).To(Equal("my-provider-postgres.swim-demo.svc.cluster.local"))
	g.Expect(data["POSTGRES_DB"]).To(Equal(testSwimDnotamDatabase))
}

func TestDnotamStrategy_ConfigMapData_ArtemisNameDerived(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	p := defaultBuildParams()
	data := s.ConfigMapData(p, "")
	g.Expect(data["ARTEMIS_BROKER_NAME"]).To(Equal(testArtemisName))
	g.Expect(data["AMQP_HOST"]).To(ContainSubstring("my-artemis-hdls-svc"))
}

func TestDnotamStrategy_ConfigMapData_ArtemisNameFallback(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	s.Payload.Artemis.Name = ""
	p := defaultBuildParams()
	data := s.ConfigMapData(p, "")
	g.Expect(data["ARTEMIS_BROKER_NAME"]).To(Equal("my-provider-artemis"))
}

func TestDnotamStrategy_ConfigMapData_ManagedKafka(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	p := defaultBuildParams()
	data := s.ConfigMapData(p, "")
	g.Expect(data["KAFKA_BOOTSTRAP_SERVERS"]).To(Equal("kafka-kafka-bootstrap.swim-demo.svc.cluster.local:9092"))
}

func TestDnotamStrategy_ConfigMapData_ExternalKafka(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	p := defaultBuildParams()
	p.Kafka.DeploymentMode = "external"
	p.Kafka.BootstrapServers = "kafka.prod.example.com:9093"
	data := s.ConfigMapData(p, "")
	g.Expect(data["KAFKA_BOOTSTRAP_SERVERS"]).To(Equal("kafka.prod.example.com:9093"))
}

func TestDnotamStrategy_ConfigMapData_ConsumeFromClientTopics(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	s.Payload.Provider.ConsumeFromClientTopics = true
	p := defaultBuildParams()
	data := s.ConfigMapData(p, "")
	g.Expect(data["KAFKA_TOPIC"]).To(Equal("dnotam-events-(?!dlq).*-topic"))
	g.Expect(data["KAFKA_PATTERN"]).To(Equal("true"))
}

func TestDnotamStrategy_ConfigMapData_DefaultTopicNotPattern(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	p := defaultBuildParams()
	data := s.ConfigMapData(p, "")
	g.Expect(data["KAFKA_TOPIC"]).To(Equal(testDnotamEventsAllTopic))
	g.Expect(data["KAFKA_PATTERN"]).To(Equal("false"))
}

func TestDnotamStrategy_ConfigMapData_K8sSecretNames(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	p := defaultBuildParams()
	data := s.ConfigMapData(p, "")
	g.Expect(data["K8S_SECRET_ADDRESS"]).To(Equal("my-artemis-dnotam-address-bp"))
	g.Expect(data["K8S_SECRET_SECURITY"]).To(Equal("my-artemis-dnotam-security-bp"))
}

func TestDnotamStrategy_ConfigMapData_AMQPSPort(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	p := defaultBuildParams()
	data := s.ConfigMapData(p, "")
	g.Expect(data["AMQP_PORT"]).To(Equal("5671"))
}

func TestDnotamStrategy_AppSecretData(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	data := s.AppSecretData()
	g.Expect(data["POSTGRES_USER"]).To(Equal(testSwimProviderCred))
	g.Expect(data["POSTGRES_PASSWORD"]).To(Equal(testSwimProviderCred))
	g.Expect(data["AMQP_USERNAME"]).To(Equal("admin"))
	g.Expect(data["AMQP_PASSWORD"]).To(Equal("admin"))
}

func TestDnotamStrategy_OIDCSecretData(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	data := s.OIDCSecretData()
	g.Expect(data["OIDC_AUTH_SERVER_URL"]).To(Equal(testOIDCAuthServerURL))
	g.Expect(data["OIDC_CLIENT_ID"]).To(Equal("dnotam-provider"))
	g.Expect(data["OIDC_CLIENT_SECRET"]).To(Equal(testProviderSecret))
}

func TestDnotamStrategy_KafkaTopics(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	g.Expect(s.KafkaTopicAllName()).To(Equal(testDnotamEventsAllTopic))
	g.Expect(s.KafkaTopicDLQName()).To(Equal("dnotam-events-dlq-topic"))
}

func TestDnotamStrategy_KafkaTopicPartitions(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	g.Expect(s.KafkaTopicPartitions(testDnotamEventsAllTopic)).To(Equal(int64(10)))
	g.Expect(s.KafkaTopicPartitions("dnotam-events-dlq-topic")).To(Equal(int64(3)))
}

func TestDnotamStrategy_SkipIfKafkaExists(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	g.Expect(s.SkipIfKafkaExists()).To(BeTrue())
}

func TestDnotamStrategy_PostgresParams(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	p := defaultBuildParams()
	pg := s.PostgresParams(p, testDnotamOperatorMgr)
	g.Expect(pg.Name).To(Equal("my-provider-postgres"))
	g.Expect(pg.Namespace).To(Equal(testNamespaceSwimDemo))
	g.Expect(pg.Database).To(Equal(testSwimDnotamDatabase))
	g.Expect(pg.ServiceAccountName).To(Equal("my-provider"))
	g.Expect(pg.SecretName).To(Equal("my-provider-postgres-secret"))
	g.Expect(pg.Labels["app.kubernetes.io/component"]).To(Equal("postgres"))
	g.Expect(pg.Labels[testLabelKubeManagedBy]).To(Equal(testDnotamOperatorMgr))
}

// --- ED-254 strategy tests ---

func TestEd254Strategy_CRKind(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	g.Expect(s.CRKind()).To(Equal("SwimEd254Provider"))
}

func TestEd254Strategy_ArtemisBrokerCleanupPrefix(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	g.Expect(s.ArtemisBrokerCleanupPrefix()).To(Equal("ed254"))
}

func TestEd254Strategy_AppImage(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	g.Expect(s.AppImage()).To(Equal("quay.io/masales/swim-ed254-provider:1.0"))
}

func TestEd254Strategy_AppImage_Default(t *testing.T) {
	g := NewWithT(t)
	s := Ed254ProviderStrategy{}
	g.Expect(s.AppImage()).To(Equal("quay.io/masales/swim-ed254-provider:latest"))
}

func TestEd254Strategy_AppReplicas(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	g.Expect(s.AppReplicas()).To(Equal(int32(3)))
}

func TestEd254Strategy_AdditionalRoleRules_IsNil(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	g.Expect(s.AdditionalRoleRules()).To(BeNil())
}

func TestEd254Strategy_ConfigMapData_ContainsAllRequiredKeys(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	p := defaultBuildParams()
	data := s.ConfigMapData(p, testAppsExampleDomain)

	requiredKeys := []string{
		"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_DB",
		"AMQP_HOST", "AMQP_PORT",
		"ARTEMIS_BROKER_NAME", "ARTEMIS_JMX_URL",
		"KAFKA_BOOTSTRAP_SERVERS", "KAFKA_TOPIC", "KAFKA_PATTERN", "KAFKA_GROUP_ID",
		"QUARKUS_HTTP_PORT", "LOG_LEVEL",
		"OTEL_ENABLED", "OTEL_SDK_DISABLED", "OTEL_ENDPOINT", "OTEL_HEADERS",
		"PROMETHEUS_ENABLED", "OPENAPI_SERVERS",
		"AERODROMES", "HEARTBEAT_INTERVAL_SECONDS",
		"K8S_SECRET_ADDRESS", "K8S_SECRET_SECURITY",
	}
	for _, key := range requiredKeys {
		g.Expect(data).To(HaveKey(key), testKeyMissingFmt, key)
	}
}

func TestEd254Strategy_ConfigMapData_Aerodromes(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	p := defaultBuildParams()
	data := s.ConfigMapData(p, "")
	g.Expect(data["AERODROMES"]).To(Equal("EGLL,LFPG,EDDF"))
}

func TestEd254Strategy_ConfigMapData_EmptyAerodromes(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	s.Payload.Provider.Aerodromes = nil
	p := defaultBuildParams()
	data := s.ConfigMapData(p, "")
	g.Expect(data["AERODROMES"]).To(BeEmpty())
}

func TestEd254Strategy_ConfigMapData_HeartbeatInterval(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	p := defaultBuildParams()
	data := s.ConfigMapData(p, "")
	g.Expect(data["HEARTBEAT_INTERVAL_SECONDS"]).To(Equal("15"))
}

func TestEd254Strategy_ConfigMapData_HeartbeatDefault(t *testing.T) {
	g := NewWithT(t)
	s := Ed254ProviderStrategy{}
	p := defaultBuildParams()
	data := s.ConfigMapData(p, "")
	g.Expect(data["HEARTBEAT_INTERVAL_SECONDS"]).To(Equal("30"))
}

func TestEd254Strategy_ConfigMapData_AMQPPort(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	p := defaultBuildParams()
	data := s.ConfigMapData(p, "")
	g.Expect(data["AMQP_PORT"]).To(Equal("5672"))
}

func TestEd254Strategy_ConfigMapData_K8sSecretNames(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	p := defaultBuildParams()
	data := s.ConfigMapData(p, "")
	g.Expect(data["K8S_SECRET_ADDRESS"]).To(Equal("my-artemis-ed254-address-bp"))
	g.Expect(data["K8S_SECRET_SECURITY"]).To(Equal("my-artemis-ed254-security-bp"))
}

func TestEd254Strategy_ConfigMapData_PostgresDB(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	p := defaultBuildParams()
	data := s.ConfigMapData(p, "")
	g.Expect(data["POSTGRES_DB"]).To(Equal(testSwimEd254Credential))
}

func TestEd254Strategy_AppSecretData_DefaultCredentials(t *testing.T) {
	g := NewWithT(t)
	s := Ed254ProviderStrategy{}
	data := s.AppSecretData()
	g.Expect(data["POSTGRES_USER"]).To(Equal(testSwimEd254Credential))
	g.Expect(data["POSTGRES_PASSWORD"]).To(Equal(testSwimEd254Credential))
	g.Expect(data["AMQP_USERNAME"]).To(Equal("admin"))
	g.Expect(data["AMQP_PASSWORD"]).To(Equal("admin"))
}

func TestEd254Strategy_OIDCSecretData(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	data := s.OIDCSecretData()
	g.Expect(data["OIDC_AUTH_SERVER_URL"]).To(Equal(testOIDCAuthServerURL))
	g.Expect(data["OIDC_CLIENT_ID"]).To(Equal("ed254-provider"))
	g.Expect(data["OIDC_CLIENT_SECRET"]).To(Equal(testProviderSecret))
}

func TestEd254Strategy_KafkaTopics(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	g.Expect(s.KafkaTopicAllName()).To(Equal("ed254-arrival-sequence-topic"))
	g.Expect(s.KafkaTopicDLQName()).To(Equal("ed254-dlq-topic"))
}

func TestEd254Strategy_KafkaTopicPartitions(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	g.Expect(s.KafkaTopicPartitions("ed254-arrival-sequence-topic")).To(Equal(int64(10)))
	g.Expect(s.KafkaTopicPartitions("ed254-dlq-topic")).To(Equal(int64(3)))
}

func TestEd254Strategy_SkipIfKafkaExists(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	g.Expect(s.SkipIfKafkaExists()).To(BeFalse())
}

func TestEd254Strategy_PostgresParams(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	p := defaultBuildParams()
	pg := s.PostgresParams(p, testEd254OperatorMgr)
	g.Expect(pg.Name).To(Equal("my-provider-postgres"))
	g.Expect(pg.Database).To(Equal(testSwimEd254Credential))
	g.Expect(pg.User).To(Equal(testSwimEd254Credential))
	g.Expect(pg.Password).To(Equal(testSwimEd254Credential))
	g.Expect(pg.Labels[testLabelKubeManagedBy]).To(Equal(testEd254OperatorMgr))
}

// --- Cross-strategy behavioral difference tests ---

func TestDnotamVsEd254_DifferentCRKind(t *testing.T) {
	g := NewWithT(t)
	dnotam := newDnotamStrategy()
	ed254 := newEd254Strategy()
	g.Expect(dnotam.CRKind()).NotTo(Equal(ed254.CRKind()))
}

func TestDnotamVsEd254_DifferentCleanupPrefix(t *testing.T) {
	g := NewWithT(t)
	dnotam := newDnotamStrategy()
	ed254 := newEd254Strategy()
	g.Expect(dnotam.ArtemisBrokerCleanupPrefix()).NotTo(Equal(ed254.ArtemisBrokerCleanupPrefix()))
}

func TestDnotamVsEd254_DifferentKafkaTopics(t *testing.T) {
	g := NewWithT(t)
	dnotam := newDnotamStrategy()
	ed254 := newEd254Strategy()
	g.Expect(dnotam.KafkaTopicAllName()).NotTo(Equal(ed254.KafkaTopicAllName()))
	g.Expect(dnotam.KafkaTopicDLQName()).NotTo(Equal(ed254.KafkaTopicDLQName()))
}

func TestDnotamVsEd254_OnlyDnotamHasAdditionalRoles(t *testing.T) {
	g := NewWithT(t)
	dnotam := newDnotamStrategy()
	ed254 := newEd254Strategy()
	g.Expect(dnotam.AdditionalRoleRules()).NotTo(BeNil())
	g.Expect(ed254.AdditionalRoleRules()).To(BeNil())
}

func TestDnotamVsEd254_OppositeKafkaSkipBehavior(t *testing.T) {
	g := NewWithT(t)
	dnotam := newDnotamStrategy()
	ed254 := newEd254Strategy()
	g.Expect(dnotam.SkipIfKafkaExists()).To(BeTrue())
	g.Expect(ed254.SkipIfKafkaExists()).To(BeFalse())
}

func TestDnotamVsEd254_K8sSecretPrefixesDiffer(t *testing.T) {
	g := NewWithT(t)
	dnotam := newDnotamStrategy()
	ed254 := newEd254Strategy()
	p := defaultBuildParams()
	dnotamData := dnotam.ConfigMapData(p, "")
	ed254Data := ed254.ConfigMapData(p, "")
	g.Expect(dnotamData["K8S_SECRET_ADDRESS"]).To(ContainSubstring("dnotam"))
	g.Expect(ed254Data["K8S_SECRET_ADDRESS"]).To(ContainSubstring("ed254"))
	g.Expect(dnotamData["K8S_SECRET_SECURITY"]).To(ContainSubstring("dnotam"))
	g.Expect(ed254Data["K8S_SECRET_SECURITY"]).To(ContainSubstring("ed254"))
}

func TestDnotamVsEd254_Ed254HasAerodromesAndHeartbeat(t *testing.T) {
	g := NewWithT(t)
	dnotam := newDnotamStrategy()
	ed254 := newEd254Strategy()
	p := defaultBuildParams()
	dnotamData := dnotam.ConfigMapData(p, "")
	ed254Data := ed254.ConfigMapData(p, "")
	g.Expect(dnotamData).NotTo(HaveKey("AERODROMES"))
	g.Expect(dnotamData).NotTo(HaveKey("HEARTBEAT_INTERVAL_SECONDS"))
	g.Expect(ed254Data).To(HaveKey("AERODROMES"))
	g.Expect(ed254Data).To(HaveKey("HEARTBEAT_INTERVAL_SECONDS"))
}

// --- Artemis Secrets tests ---

func TestDnotamStrategy_ArtemisAddressBPSecret(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	p := defaultBuildParams()
	secret := s.ArtemisAddressBPSecret(p, testDnotamOperatorMgr)
	g.Expect(secret.Name).To(ContainSubstring("dnotam-address-bp"))
	g.Expect(secret.Namespace).To(Equal(testNamespaceSwimDemo))
	g.Expect(secret.Labels[testLabelKubeManagedBy]).To(Equal(testDnotamOperatorMgr))
}

func TestDnotamStrategy_ArtemisSecurityBPSecret(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	p := defaultBuildParams()
	secret := s.ArtemisSecurityBPSecret(p, testDnotamOperatorMgr)
	g.Expect(secret.Name).To(ContainSubstring("dnotam-security-bp"))
	g.Expect(secret.StringData).To(HaveKey("securityRoles.properties"))
}

func TestEd254Strategy_ArtemisAddressBPSecret(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	p := defaultBuildParams()
	secret := s.ArtemisAddressBPSecret(p, testEd254OperatorMgr)
	g.Expect(secret.Name).To(Equal("my-artemis-ed254-address-bp"))
	g.Expect(secret.Namespace).To(Equal(testNamespaceSwimDemo))
	g.Expect(secret.StringData).To(HaveKey("addressConfigurations.properties"))
}

func TestEd254Strategy_ArtemisSecurityBPSecret_ContainsAdminRoles(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	p := defaultBuildParams()
	secret := s.ArtemisSecurityBPSecret(p, testEd254OperatorMgr)
	g.Expect(secret.Name).To(Equal("my-artemis-ed254-security-bp"))
	roles := secret.StringData["securityRoles.properties"]
	g.Expect(roles).To(ContainSubstring("securityRoles.#.admin.consume=true"))
	g.Expect(roles).To(ContainSubstring("securityRoles.#.admin.send=true"))
	g.Expect(roles).To(ContainSubstring("securityRoles.#.admin.createAddress=true"))
}

// --- ArtemisBaseParams tests ---

func TestDnotamStrategy_ArtemisBaseParams_ExtraMounts(t *testing.T) {
	g := NewWithT(t)
	s := newDnotamStrategy()
	p := defaultBuildParams()
	params := s.ArtemisBaseParams(p, "")
	g.Expect(params.ExtraMounts).To(HaveLen(3))
	g.Expect(params.ExtraMounts[0]).To(ContainSubstring("sso-jaas-config"))
	g.Expect(params.ExtraMounts[1]).To(ContainSubstring("dnotam-address-bp"))
	g.Expect(params.ExtraMounts[2]).To(ContainSubstring("dnotam-security-bp"))
}

func TestEd254Strategy_ArtemisBaseParams_ExtraMounts(t *testing.T) {
	g := NewWithT(t)
	s := newEd254Strategy()
	p := defaultBuildParams()
	params := s.ArtemisBaseParams(p, "")
	g.Expect(params.ExtraMounts).To(HaveLen(3))
	g.Expect(params.ExtraMounts[0]).To(ContainSubstring("sso-jaas-config"))
	g.Expect(params.ExtraMounts[1]).To(ContainSubstring("ed254-address-bp"))
	g.Expect(params.ExtraMounts[2]).To(ContainSubstring("ed254-security-bp"))
}

// --- Interface compliance tests ---

func TestDnotamStrategy_ImplementsProviderStrategy(t *testing.T) {
	var _ ProviderStrategy = DnotamProviderStrategy{}
}

func TestEd254Strategy_ImplementsProviderStrategy(t *testing.T) {
	var _ ProviderStrategy = Ed254ProviderStrategy{}
}

// --- openapiServersValue tests ---

func TestOpenapiServersValue_BothEnabled(t *testing.T) {
	g := NewWithT(t)
	ex := ProviderExposureSpec{
		HTTPEdgeEnabled:         true,
		HTTPSPassthroughEnabled: true,
		HTTPSEdgeHost:           "api.example.com",
		HTTPSPassthroughHost:    "mtls.example.com",
	}
	result := openapiServersValue(ex, testMyAppName, testNamespaceSwimDemo, testAppsExampleDomain)
	g.Expect(result).To(ContainSubstring("https://api.example.com"))
	g.Expect(result).To(ContainSubstring("https://mtls.example.com"))
}

func TestOpenapiServersValue_NoneEnabled(t *testing.T) {
	g := NewWithT(t)
	ex := ProviderExposureSpec{}
	result := openapiServersValue(ex, testMyAppName, testNamespaceSwimDemo, testAppsExampleDomain)
	g.Expect(result).To(BeEmpty())
}

func TestOpenapiServersValue_OnlyEdge(t *testing.T) {
	g := NewWithT(t)
	ex := ProviderExposureSpec{
		HTTPEdgeEnabled: true,
		HTTPSEdgeHost:   "api.example.com",
	}
	result := openapiServersValue(ex, testMyAppName, testNamespaceSwimDemo, testAppsExampleDomain)
	g.Expect(result).To(Equal("https://api.example.com"))
}
