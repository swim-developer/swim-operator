package resources

import (
	"fmt"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ArtemisProviderParams struct {
	ArtemisName      string
	Namespace        string
	Labels           map[string]string
	AdminUser        string
	AdminPassword    string
	KeystorePassword string
	IssuerName       string
	IssuerKind       string
	IngressHost      string
	ExposeMode       string
	Size             int32
	VerifyHost       bool
	AMQPSPort        int32
	AMQPPort         int32
	BrokerProperties []string
	StorageSize      string
	StorageClassName string
	ConsoleExpose    bool
	ConsoleSSL       bool
	Image            string
	InitImage        string
	ExtraMounts      []string
	JMXEnabled       bool
	JMXPort          int32
	OIDCEnabled      bool
	OIDCRealm        string
	OIDCAuthServerURL string
	OIDCClientId     string
	OIDCClientSecret string
}

func DefaultArtemisName(specName, crName string) string {
	if specName != "" {
		return specName
	}
	return fmt.Sprintf("%s-artemis", crName)
}

func BuildProviderArtemisCredentialsSecret(p ArtemisProviderParams) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-credentials", p.ArtemisName), Namespace: p.Namespace, Labels: p.Labels,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"AMQ_USER":             StrDefault(p.AdminUser, "admin"),
			"AMQ_PASSWORD":         StrDefault(p.AdminPassword, "admin"),
			"AMQ_CLUSTER_USER":     "clusterUser",
			"AMQ_CLUSTER_PASSWORD": "clusterPassword",
		},
	}
}

func BuildProviderArtemisKeystoreSecret(p ArtemisProviderParams) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-keystore-password", p.ArtemisName), Namespace: p.Namespace, Labels: p.Labels,
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{"password": StrDefault(p.KeystorePassword, "changeit")},
	}
}

func BuildProviderArtemisCertificate(p ArtemisProviderParams) *certmanagerv1.Certificate {
	return &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-amqp", p.ArtemisName), Namespace: p.Namespace, Labels: p.Labels,
		},
		Spec: certmanagerv1.CertificateSpec{
			SecretName: fmt.Sprintf("%s-amqp-tls", p.ArtemisName),
			IssuerRef: cmmeta.ObjectReference{
				Name: p.IssuerName, Kind: p.IssuerKind,
			},
			CommonName: p.ArtemisName,
			DNSNames:   artemisDNSNames(p.ArtemisName, p.Namespace, p.IngressHost),
			Usages:     []certmanagerv1.KeyUsage{certmanagerv1.UsageServerAuth, certmanagerv1.UsageClientAuth},
			Keystores: &certmanagerv1.CertificateKeystores{
				JKS: &certmanagerv1.JKSKeystore{
					Create: true,
					PasswordSecretRef: cmmeta.SecretKeySelector{
						LocalObjectReference: cmmeta.LocalObjectReference{Name: fmt.Sprintf("%s-keystore-password", p.ArtemisName)},
						Key:                  "password",
					},
				},
			},
			PrivateKey: &certmanagerv1.CertificatePrivateKey{
				Algorithm: certmanagerv1.RSAKeyAlgorithm, Size: 2048, RotationPolicy: certmanagerv1.RotationPolicyAlways,
			},
		},
	}
}

func BuildProviderArtemisOIDCSecret(p ArtemisProviderParams) *corev1.Secret {
	configPath := fmt.Sprintf("%s-sso-jaas-config", p.ArtemisName)
	moduleBlock := `    org.keycloak.adapters.jaas.BearerTokenLoginModule required
        keycloak-config-file="/amq/extra/secrets/` + configPath + `/_keycloak-login-module.json"
        role-principal-class="org.apache.activemq.artemis.spi.core.security.jaas.RolePrincipal";`
	loginConfig := BuildArtemisJaasLoginConfig(moduleBlock)

	keycloakConfig := fmt.Sprintf(`{
  "realm": "%s",
  "auth-server-url": "%s",
  "ssl-required": "external",
  "resource": "%s",
  "bearer-only": true,
  "verify-token-audience": true,
  "use-resource-role-mappings": true
}`, p.OIDCRealm, p.OIDCAuthServerURL, p.OIDCClientId)

	if p.OIDCClientSecret != "" {
		keycloakConfig = fmt.Sprintf(`{
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
}`, p.OIDCRealm, p.OIDCAuthServerURL, p.OIDCClientId, p.OIDCClientSecret)
	}

	return BuildArtemisJaasConfigSecret(configPath, p.Namespace, p.Labels, loginConfig, keycloakConfig)
}

func BuildArtemisJaasLoginConfig(moduleBlock string) string {
	return `activemq {

    org.apache.activemq.artemis.spi.core.security.jaas.PropertiesLoginModule sufficient
        org.apache.activemq.jaas.properties.user="artemis-users.properties"
        org.apache.activemq.jaas.properties.role="artemis-roles.properties"
        baseDir="/home/jboss/amq-broker/etc";

` + moduleBlock + `

    org.apache.activemq.artemis.spi.core.security.jaas.PrincipalConversionLoginModule required
        principalClassList="org.keycloak.KeycloakPrincipal";

};`
}

func BuildArtemisJaasConfigSecret(name, namespace string, labels map[string]string, loginConfig, keycloakConfig string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"login.config":                loginConfig,
			"_keycloak-login-module.json": keycloakConfig,
		},
	}
}

func BuildProviderArtemisAddressBPSecret(artemisName, namespace string, labels map[string]string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-dnotam-address-bp", artemisName), Namespace: namespace, Labels: labels,
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{"addressConfigurations.properties": ""},
	}
}

func BuildProviderArtemisSecurityBPSecret(artemisName, namespace string, labels map[string]string) *corev1.Secret {
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
			Name: fmt.Sprintf("%s-dnotam-security-bp", artemisName), Namespace: namespace, Labels: labels,
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{"securityRoles.properties": securityRoles},
	}
}

func BuildProviderArtemisBroker(p ArtemisProviderParams) *unstructured.Unstructured {
	size := int64(Int32Default(p.Size, 1))
	amqpsPort := int64(Int32Default(p.AMQPSPort, 5671))
	amqpPort := int64(Int32Default(p.AMQPPort, 5672))
	storageSize := StrDefault(p.StorageSize, "5Gi")
	exposeMode := StrDefault(p.ExposeMode, "route")

	acceptors := []interface{}{
		map[string]interface{}{
			"name": "amqps", "port": amqpsPort, "protocols": "AMQP",
			"sslEnabled": true, "sslSecret": fmt.Sprintf("%s-ssl-secret", p.ArtemisName),
			"expose": true, "exposeMode": exposeMode, "ingressHost": p.IngressHost,
			"needClientAuth": true, "verifyHost": p.VerifyHost,
		},
		map[string]interface{}{
			"name": "amqp", "port": amqpPort, "protocols": "AMQP",
			"sslEnabled": false, "expose": false,
		},
	}

	brokerProperties := []interface{}{
		"globalMaxSize=1G",
		"addressSettings.#.expiryDelay=86400000",
		"addressSettings.#.expiryAddress=ExpiryQueue",
		"addressSettings.#.autoDeleteQueues=false",
		"addressSettings.#.autoDeleteAddresses=false",
		"addressSettings.#.maxDeliveryAttempts=10",
		"addressSettings.#.redeliveryDelay=5000",
		"addressSettings.#.addressFullMessagePolicy=PAGE",
		"addressSettings.#.slowConsumerThreshold=5",
		"addressSettings.#.slowConsumerPolicy=NOTIFY",
	}
	for _, prop := range p.BrokerProperties {
		brokerProperties = append(brokerProperties, prop)
	}

	storage := map[string]interface{}{"size": storageSize}
	if p.StorageClassName != "" {
		storage["storageClassName"] = p.StorageClassName
	}

	deploymentPlan := map[string]interface{}{
		"size": size, "persistenceEnabled": true, "journalType": "nio", "requireLogin": true,
		"storage": storage,
	}
	if p.Image != "" {
		deploymentPlan["image"] = p.Image
	}
	if p.InitImage != "" {
		deploymentPlan["initImage"] = p.InitImage
	}
	if len(p.ExtraMounts) > 0 {
		secrets := make([]interface{}, len(p.ExtraMounts))
		for i, s := range p.ExtraMounts {
			secrets[i] = s
		}
		deploymentPlan["extraMounts"] = map[string]interface{}{"secrets": secrets}
	}

	console := map[string]interface{}{"expose": p.ConsoleExpose}
	if p.ConsoleSSL {
		console["sslEnabled"] = true
		console["sslSecret"] = fmt.Sprintf("%s-ssl-secret", p.ArtemisName)
	}

	spec := map[string]interface{}{
		"deploymentPlan":   deploymentPlan,
		"adminUser":        StrDefault(p.AdminUser, "admin"),
		"adminPassword":    StrDefault(p.AdminPassword, "admin"),
		"acceptors":        acceptors,
		"brokerProperties": brokerProperties,
		"console":          console,
	}

	if p.JMXEnabled {
		jmxPort := int64(Int32Default(p.JMXPort, 1099))
		jmxHost := fmt.Sprintf("%s-jmx-svc.%s.svc.cluster.local", p.ArtemisName, p.Namespace)
		spec["env"] = []interface{}{
			map[string]interface{}{
				"name": "JDK_JAVA_OPTIONS",
				"value": fmt.Sprintf(
					"-Dcom.sun.management.jmxremote=true -Dcom.sun.management.jmxremote.port=%d -Dcom.sun.management.jmxremote.rmi.port=%d -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote.authenticate=false -Djava.rmi.server.hostname=%s",
					jmxPort, jmxPort, jmxHost,
				),
			},
		}
	}

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "broker.amq.io/v1beta1", "kind": "ActiveMQArtemis",
			"metadata": map[string]interface{}{
				"name": p.ArtemisName, "namespace": p.Namespace, "labels": LabelsToUnstructured(p.Labels),
			},
			"spec": spec,
		},
	}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "broker.amq.io", Version: "v1beta1", Kind: "ActiveMQArtemis"})
	return obj
}

func BuildProviderArtemisJMXService(artemisName, namespace string, labels map[string]string, jmxPort int32) *corev1.Service {
	port := Int32Default(jmxPort, 1099)
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-jmx-svc", artemisName), Namespace: namespace, Labels: labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"ActiveMQArtemis": artemisName,
				"application":     fmt.Sprintf("%s-app", artemisName),
			},
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{{
				Name: "jmx", Port: port, TargetPort: intstr.FromInt32(port), Protocol: corev1.ProtocolTCP,
			}},
		},
	}
}
