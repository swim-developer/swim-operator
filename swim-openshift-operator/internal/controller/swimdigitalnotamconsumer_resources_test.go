package controller

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	appsv1alpha1 "github.com/swim-developer/swim-openshift-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testConsumerName        = "testConsumerName"
	testAppName             = "testAppName"
	testProviderURL         = "https://sm.example.com"
	testBrokerExampleDomain = "broker.example.com"
)

func minimalConsumer(name, namespace string) *appsv1alpha1.SwimDigitalNotamConsumer {
	return &appsv1alpha1.SwimDigitalNotamConsumer{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	}
}

func noopReconciler() *SwimDigitalNotamConsumerReconciler {
	return &SwimDigitalNotamConsumerReconciler{}
}

// ---------------------------------------------------------------------------
// ClientConfigMap
// ---------------------------------------------------------------------------

func TestClientConfigMap_ResourceName(t *testing.T) {
	g := NewWithT(t)
	cm := noopReconciler().ClientConfigMap(minimalConsumer("testConsumerName", "default"))
	g.Expect(cm.Name).To(Equal("testConsumerName-config"))
	g.Expect(cm.Namespace).To(Equal("default"))
}

func TestClientConfigMap_ManagedKafkaBootstrap(t *testing.T) {
	g := NewWithT(t)
	cr := minimalConsumer("testConsumerName", "swim-ns")
	cm := noopReconciler().ClientConfigMap(cr)
	g.Expect(cm.Data["KAFKA_BOOTSTRAP_SERVERS"]).To(Equal("kafka-kafka-bootstrap.swim-ns.svc.cluster.local:9092"))
}

func TestClientConfigMap_ExternalKafkaBootstrap(t *testing.T) {
	g := NewWithT(t)
	cr := minimalConsumer("testConsumerName", "default")
	cr.Spec.Kafka.DeploymentMode = "external"
	cr.Spec.Kafka.BootstrapServers = "kafka.mycompany.com:9092"
	cm := noopReconciler().ClientConfigMap(cr)
	g.Expect(cm.Data["KAFKA_BOOTSTRAP_SERVERS"]).To(Equal("kafka.mycompany.com:9092"))
}

func TestClientConfigMap_DefaultValues(t *testing.T) {
	g := NewWithT(t)
	cr := minimalConsumer("testConsumerName", "default")
	cm := noopReconciler().ClientConfigMap(cr)
	g.Expect(cm.Data["MONGODB_HOST"]).To(Equal("testConsumerName-mongodb.default.svc.cluster.local"))
	g.Expect(cm.Data["MONGODB_PORT"]).To(Equal("27017"))
	g.Expect(cm.Data["MONGODB_DATABASE"]).To(Equal("swim-dnotam"))
	g.Expect(cm.Data["SWIM_VALIDATION_ENABLED"]).To(Equal("true"))
	g.Expect(cm.Data["SWIM_VALIDATION_FAIL_ON_NULLBODY"]).To(Equal("false"))
	g.Expect(cm.Data["DNOTAM_DELETE_AND_RECREATE"]).To(Equal("true"))
}

func TestClientConfigMap_CustomMongoDatabase(t *testing.T) {
	g := NewWithT(t)
	cr := minimalConsumer("testConsumerName", "default")
	cr.Spec.Client.Mongo.Database = "custom-db"
	cm := noopReconciler().ClientConfigMap(cr)
	g.Expect(cm.Data["MONGODB_DATABASE"]).To(Equal("custom-db"))
}

// ---------------------------------------------------------------------------
// MongoSecret
// ---------------------------------------------------------------------------

func TestMongoSecret_ResourceName(t *testing.T) {
	g := NewWithT(t)
	s := noopReconciler().MongoSecret(minimalConsumer("testConsumerName", "default"))
	g.Expect(s.Name).To(Equal("testConsumerName-mongodb-credentials"))
	g.Expect(s.Namespace).To(Equal("default"))
}

func TestMongoSecret_DefaultCredentials(t *testing.T) {
	g := NewWithT(t)
	s := noopReconciler().MongoSecret(minimalConsumer("testConsumerName", "default"))
	g.Expect(s.StringData["database-name"]).To(Equal("swim-dnotam"))
	g.Expect(s.StringData[DatabaseUserKey]).To(Equal("swim"))
	g.Expect(s.StringData[DatabasePasswordKey]).To(Equal("swim123"))
	g.Expect(s.StringData["database-admin-password"]).To(Equal("admin"))
}

func TestMongoSecret_CustomCredentials(t *testing.T) {
	g := NewWithT(t)
	cr := minimalConsumer("testConsumerName", "default")
	cr.Spec.Client.Mongo.Database = "mydb"
	cr.Spec.Client.Mongo.User = "myuser"
	cr.Spec.Client.Mongo.Password = "mypassword"
	s := noopReconciler().MongoSecret(cr)
	g.Expect(s.StringData["database-name"]).To(Equal("mydb"))
	g.Expect(s.StringData[DatabaseUserKey]).To(Equal("myuser"))
	g.Expect(s.StringData[DatabasePasswordKey]).To(Equal("mypassword"))
}

// ---------------------------------------------------------------------------
// MongoPVC
// ---------------------------------------------------------------------------

func TestMongoPVC_ResourceName(t *testing.T) {
	g := NewWithT(t)
	pvc := noopReconciler().MongoPVC(minimalConsumer("testConsumerName", "default"))
	g.Expect(pvc.Name).To(Equal("testConsumerName-mongodb-data"))
}

func TestMongoPVC_DefaultStorage(t *testing.T) {
	g := NewWithT(t)
	pvc := noopReconciler().MongoPVC(minimalConsumer("testConsumerName", "default"))
	g.Expect(pvc.Spec.AccessModes).To(ContainElement(corev1.ReadWriteOnce))
	g.Expect(pvc.Spec.Resources.Requests[corev1.ResourceStorage]).To(Equal(resource.MustParse("1Gi")))
}

func TestMongoPVC_CustomStorage(t *testing.T) {
	g := NewWithT(t)
	cr := minimalConsumer("testConsumerName", "default")
	cr.Spec.Client.Mongo.StorageSize = "10Gi"
	pvc := noopReconciler().MongoPVC(cr)
	g.Expect(pvc.Spec.Resources.Requests[corev1.ResourceStorage]).To(Equal(resource.MustParse("10Gi")))
}

// ---------------------------------------------------------------------------
// ClientKeystorePasswordSecret
// ---------------------------------------------------------------------------

func TestClientKeystorePasswordSecret(t *testing.T) {
	g := NewWithT(t)
	s := noopReconciler().ClientKeystorePasswordSecret(minimalConsumer("testConsumerName", "default"))
	g.Expect(s.Name).To(Equal("testConsumerName-keystore-password"))
	g.Expect(s.StringData["password"]).To(Equal("changeit"))
}

// ---------------------------------------------------------------------------
// ClientProvidersSecret
// ---------------------------------------------------------------------------

func TestClientProvidersSecret_ResourceName(t *testing.T) {
	g := NewWithT(t)
	s := noopReconciler().ClientProvidersSecret(minimalConsumer("testConsumerName", "default"))
	g.Expect(s.Name).To(Equal("testConsumerName-providers"))
}

func TestClientProvidersSecret_DefaultTLS(t *testing.T) {
	g := NewWithT(t)
	cr := minimalConsumer("testConsumerName", "default")
	cr.Spec.Client.Config.Providers = []commonapi.ProviderSpec{
		{
			ProviderId:          "test-provider",
			SubscriptionManager: commonapi.SubscriptionManagerSpec{URL: testProviderURL},
			AmqpBroker:          commonapi.AmqpBrokerSpec{Host: testBrokerExampleDomain, Port: 5671, SSLEnabled: true},
		},
	}
	s := noopReconciler().ClientProvidersSecret(cr)
	var providers []map[string]interface{}
	g.Expect(json.Unmarshal([]byte(s.StringData["SWIM_PROVIDERS"]), &providers)).To(Succeed())
	g.Expect(providers).To(HaveLen(1))
	g.Expect(providers[0]["providerId"]).To(Equal("test-provider"))
	sm := providers[0]["subscriptionManager"].(map[string]interface{})
	g.Expect(sm["url"]).To(Equal(testProviderURL))
	tls := sm["tls"].(map[string]interface{})
	g.Expect(tls["trustStorePath"]).To(Equal("/secrets/truststore.p12"))
	g.Expect(tls["keyStorePath"]).To(Equal("/secrets/keystore.p12"))
}

func TestClientProvidersSecret_DefaultResilienceValues(t *testing.T) {
	g := NewWithT(t)
	cr := minimalConsumer("testConsumerName", "default")
	cr.Spec.Client.Config.Providers = []commonapi.ProviderSpec{
		{
			ProviderId:          "provider",
			SubscriptionManager: commonapi.SubscriptionManagerSpec{URL: testProviderURL},
			AmqpBroker:          commonapi.AmqpBrokerSpec{Host: testBrokerExampleDomain, Port: 5671},
		},
	}
	s := noopReconciler().ClientProvidersSecret(cr)
	var providers []map[string]interface{}
	g.Expect(json.Unmarshal([]byte(s.StringData["SWIM_PROVIDERS"]), &providers)).To(Succeed())
	sm := providers[0]["subscriptionManager"].(map[string]interface{})
	resilience := sm["resilience"].(map[string]interface{})
	g.Expect(resilience["connectTimeoutMs"]).To(BeEquivalentTo(5000))
	g.Expect(resilience["readTimeoutMs"]).To(BeEquivalentTo(30000))
	g.Expect(resilience["retryMaxAttempts"]).To(BeEquivalentTo(3))
	g.Expect(resilience["retryDelayMs"]).To(BeEquivalentTo(1000))
}

// ---------------------------------------------------------------------------
// ClientCertificate
// ---------------------------------------------------------------------------

func TestClientCertificate_DefaultIssuer(t *testing.T) {
	g := NewWithT(t)
	cert := noopReconciler().ClientCertificate(minimalConsumer("testConsumerName", "default"))
	g.Expect(cert.Name).To(Equal("testConsumerName-mtls-cert"))
	g.Expect(cert.Spec.SecretName).To(Equal("testConsumerName-mtls"))
	g.Expect(cert.Spec.IssuerRef.Name).To(Equal("swim-ca-issuer"))
	g.Expect(cert.Spec.IssuerRef.Kind).To(Equal("ClusterIssuer"))
	g.Expect(cert.Spec.CommonName).To(Equal("testConsumerName"))
}

func TestClientCertificate_CustomIssuer(t *testing.T) {
	g := NewWithT(t)
	cr := minimalConsumer("testConsumerName", "default")
	cr.Spec.CertManager.IssuerName = "my-custom-issuer"
	cr.Spec.CertManager.IssuerKind = "Issuer"
	cert := noopReconciler().ClientCertificate(cr)
	g.Expect(cert.Spec.IssuerRef.Name).To(Equal("my-custom-issuer"))
	g.Expect(cert.Spec.IssuerRef.Kind).To(Equal("Issuer"))
}

func TestClientCertificate_KeystoresEnabled(t *testing.T) {
	g := NewWithT(t)
	cert := noopReconciler().ClientCertificate(minimalConsumer("testConsumerName", "default"))
	g.Expect(cert.Spec.Keystores).NotTo(BeNil())
	g.Expect(cert.Spec.Keystores.PKCS12).NotTo(BeNil())
	g.Expect(cert.Spec.Keystores.PKCS12.Create).To(BeTrue())
	g.Expect(cert.Spec.Keystores.JKS).NotTo(BeNil())
	g.Expect(cert.Spec.Keystores.JKS.Create).To(BeTrue())
}

// ---------------------------------------------------------------------------
// MongoDeployment
// ---------------------------------------------------------------------------

func TestMongoDeployment_ResourceName(t *testing.T) {
	g := NewWithT(t)
	d := noopReconciler().MongoDeployment(minimalConsumer("testConsumerName", "default"), "abc123")
	g.Expect(d.Name).To(Equal("testConsumerName-mongodb"))
}

func TestMongoDeployment_ContainerImage(t *testing.T) {
	g := NewWithT(t)
	d := noopReconciler().MongoDeployment(minimalConsumer("testConsumerName", "default"), "abc123")
	g.Expect(d.Spec.Template.Spec.Containers[0].Image).To(Equal("quay.io/mongodb/mongodb-community-server:8.0-ubi8"))
}

func TestMongoDeployment_ReplicasAndStrategy(t *testing.T) {
	g := NewWithT(t)
	d := noopReconciler().MongoDeployment(minimalConsumer("testConsumerName", "default"), "abc123")
	g.Expect(*d.Spec.Replicas).To(Equal(int32(1)))
	g.Expect(d.Spec.Strategy.Type).To(Equal(appsv1.RecreateDeploymentStrategyType))
}

func TestMongoDeployment_ConfigHashAnnotation(t *testing.T) {
	g := NewWithT(t)
	d := noopReconciler().MongoDeployment(minimalConsumer("testConsumerName", "default"), "deadbeef")
	g.Expect(d.Spec.Template.Annotations["config-hash"]).To(Equal("deadbeef"))
}

// ---------------------------------------------------------------------------
// ClientDeployment
// ---------------------------------------------------------------------------

func TestClientDeployment_ResourceName(t *testing.T) {
	g := NewWithT(t)
	d := noopReconciler().ClientDeployment(minimalConsumer("testConsumerName", "default"), "abc123")
	g.Expect(d.Name).To(Equal("testConsumerName"))
	g.Expect(d.Namespace).To(Equal("default"))
}

func TestClientDeployment_DefaultImage(t *testing.T) {
	g := NewWithT(t)
	d := noopReconciler().ClientDeployment(minimalConsumer("testConsumerName", "default"), "abc123")
	g.Expect(d.Spec.Template.Spec.Containers[0].Image).To(Equal("quay.io/masales/swim-dnotam-consumer:latest"))
}

func TestClientDeployment_CustomImage(t *testing.T) {
	g := NewWithT(t)
	cr := minimalConsumer("testConsumerName", "default")
	cr.Spec.Client.Image = "quay.io/myorg/testConsumerName:v1.2.3"
	d := noopReconciler().ClientDeployment(cr, "abc123")
	g.Expect(d.Spec.Template.Spec.Containers[0].Image).To(Equal("quay.io/myorg/testConsumerName:v1.2.3"))
}

func TestClientDeployment_ServiceAccount(t *testing.T) {
	g := NewWithT(t)
	d := noopReconciler().ClientDeployment(minimalConsumer("testConsumerName", "default"), "abc123")
	g.Expect(d.Spec.Template.Spec.ServiceAccountName).To(Equal("testConsumerName"))
}

func TestClientDeployment_InitContainerValidatesSecrets(t *testing.T) {
	g := NewWithT(t)
	d := noopReconciler().ClientDeployment(minimalConsumer("testConsumerName", "default"), "abc123")
	g.Expect(d.Spec.Template.Spec.InitContainers).To(HaveLen(1))
	g.Expect(d.Spec.Template.Spec.InitContainers[0].Name).To(Equal("validate-secrets"))
}

func TestClientDeployment_MtlsVolumeFromSecret(t *testing.T) {
	g := NewWithT(t)
	d := noopReconciler().ClientDeployment(minimalConsumer("testConsumerName", "default"), "abc123")
	var mtls *corev1.Volume
	for i := range d.Spec.Template.Spec.Volumes {
		if d.Spec.Template.Spec.Volumes[i].Name == MTLSCertsVolume {
			mtls = &d.Spec.Template.Spec.Volumes[i]
			break
		}
	}
	g.Expect(mtls).NotTo(BeNil())
	g.Expect(mtls.Secret.SecretName).To(Equal("testConsumerName-mtls"))
}

// ---------------------------------------------------------------------------
// Services
// ---------------------------------------------------------------------------

func TestClientService_PortsAndName(t *testing.T) {
	g := NewWithT(t)
	svc := noopReconciler().ClientService(minimalConsumer("testConsumerName", "default"))
	g.Expect(svc.Name).To(Equal("testConsumerName"))
	portNames := make([]string, len(svc.Spec.Ports))
	for i, p := range svc.Spec.Ports {
		portNames[i] = p.Name
	}
	g.Expect(portNames).To(ContainElements("http", "management"))
}

func TestMongoService_PortAndName(t *testing.T) {
	g := NewWithT(t)
	svc := noopReconciler().MongoService(minimalConsumer("testConsumerName", "default"))
	g.Expect(svc.Name).To(Equal("testConsumerName-mongodb"))
	g.Expect(svc.Spec.Ports[0].Port).To(Equal(int32(27017)))
}

// ---------------------------------------------------------------------------
// ClientHPA
// ---------------------------------------------------------------------------

func TestClientHPA(t *testing.T) {
	g := NewWithT(t)
	hpa := noopReconciler().ClientHPA(minimalConsumer("testConsumerName", "default"))
	g.Expect(hpa.Name).To(Equal("testConsumerName-hpa"))
	g.Expect(*hpa.Spec.MinReplicas).To(Equal(int32(1)))
	g.Expect(hpa.Spec.MaxReplicas).To(Equal(int32(5)))
	g.Expect(*hpa.Spec.Metrics[0].Resource.Target.AverageUtilization).To(Equal(int32(70)))
}

// ---------------------------------------------------------------------------
// Labels
// ---------------------------------------------------------------------------

func TestStandardLabels_AllKeysPresent(t *testing.T) {
	g := NewWithT(t)
	labels := standardLabels("testAppName", "client", "my-cr")
	g.Expect(labels["app"]).To(Equal("testAppName"))
	g.Expect(labels["app.kubernetes.io/name"]).To(Equal("testAppName"))
	g.Expect(labels["app.kubernetes.io/instance"]).To(Equal("my-cr"))
	g.Expect(labels["app.kubernetes.io/component"]).To(Equal("client"))
	g.Expect(labels["app.kubernetes.io/part-of"]).To(Equal("swim-operator"))
	g.Expect(labels["app.kubernetes.io/managed-by"]).To(Equal("swim-operator"))
}
