package resources

import (
	"fmt"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ArtemisCVParams struct {
	CRName           string
	Namespace        string
	Labels           map[string]string
	ArtemisName      string
	AdminUser        string
	AdminPassword    string
	ClusterUser      string
	ClusterPassword  string
	KeystorePassword string
	IssuerName       string
	IssuerKind       string
	IngressHost      string
	ExposeMode       string
	Size             int32
	BrokerProperties []string
	ConsoleEnabled   bool
	ConsoleExpose    bool
	ConsoleSSL       bool
	AmqpEnabled      bool
	AmqpPort         int32
	AmqpExpose       bool
}

func BuildArtemisCVCredentialsSecret(p ArtemisCVParams) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-artemis-credentials", p.CRName), Namespace: p.Namespace, Labels: p.Labels,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"AMQ_USER":             StrDefault(p.AdminUser, "admin"),
			"AMQ_PASSWORD":         StrDefault(p.AdminPassword, "admin"),
			"AMQ_CLUSTER_USER":     StrDefault(p.ClusterUser, "clusterUser"),
			"AMQ_CLUSTER_PASSWORD": StrDefault(p.ClusterPassword, "clusterPassword"),
		},
	}
}

func BuildArtemisCVKeystoreSecret(p ArtemisCVParams) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-artemis-keystore-password", p.CRName), Namespace: p.Namespace, Labels: p.Labels,
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{"password": StrDefault(p.KeystorePassword, "changeit")},
	}
}

func BuildArtemisCVCertificate(p ArtemisCVParams) *certmanagerv1.Certificate {
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
						LocalObjectReference: cmmeta.LocalObjectReference{Name: fmt.Sprintf("%s-artemis-keystore-password", p.CRName)},
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

func BuildArtemisCVBroker(p ArtemisCVParams) *unstructured.Unstructured {
	size := int64(Int32Default(p.Size, 1))
	exposeMode := StrDefault(p.ExposeMode, "route")

	acceptors := []interface{}{
		map[string]interface{}{
			"name": "amqps", "port": int64(5671), "protocols": "AMQP",
			"sslEnabled": true, "sslSecret": fmt.Sprintf("%s-ssl-secret", p.ArtemisName),
			"expose": true, "exposeMode": exposeMode, "ingressHost": p.IngressHost,
			"needClientAuth": true, "verifyHost": false,
		},
	}

	amqpEnabled := p.AmqpEnabled
	if amqpEnabled {
		amqpPort := int64(Int32Default(p.AmqpPort, 5672))
		acceptors = append(acceptors, map[string]interface{}{
			"name": "amqp", "port": amqpPort, "protocols": "AMQP",
			"sslEnabled": false, "expose": p.AmqpExpose,
		})
	}

	brokerProperties := []interface{}{
		"globalMaxSize=512m",
		"addressSettings.#.expiryDelay=86400000",
		"addressSettings.#.expiryAddress=ExpiryQueue",
		"addressSettings.#.autoDeleteQueues=true",
		"addressSettings.#.autoDeleteAddresses=true",
	}
	for _, prop := range p.BrokerProperties {
		brokerProperties = append(brokerProperties, prop)
	}

	spec := map[string]interface{}{
		"deploymentPlan": map[string]interface{}{
			"size": size, "persistenceEnabled": true, "journalType": "nio", "requireLogin": true,
		},
		"adminUser": StrDefault(p.AdminUser, "admin"), "adminPassword": StrDefault(p.AdminPassword, "admin"),
		"acceptors": acceptors, "brokerProperties": brokerProperties,
	}

	if p.ConsoleEnabled {
		console := map[string]interface{}{"expose": p.ConsoleExpose}
		if p.ConsoleSSL {
			console["sslEnabled"] = true
			console["sslSecret"] = fmt.Sprintf("%s-ssl-secret", p.ArtemisName)
		}
		spec["console"] = console
	}

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "broker.amq.io/v1beta1", "kind": "ActiveMQArtemis",
			"metadata": map[string]interface{}{"name": p.ArtemisName, "namespace": p.Namespace, "labels": LabelsToUnstructured(p.Labels)},
			"spec":     spec,
		},
	}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "broker.amq.io", Version: "v1beta1", Kind: "ActiveMQArtemis"})
	return obj
}

func artemisDNSNames(artemisName, namespace, ingressHost string) []string {
	names := []string{}
	if ingressHost != "" {
		names = append(names, ingressHost)
	}
	names = append(names,
		fmt.Sprintf("%s-amqps-0-svc.%s.svc.cluster.local", artemisName, namespace),
		fmt.Sprintf("%s-hdls-svc.%s.svc.cluster.local", artemisName, namespace),
		fmt.Sprintf("%s-amqps-0-svc.%s.svc", artemisName, namespace),
		fmt.Sprintf("%s-hdls-svc.%s.svc", artemisName, namespace),
		fmt.Sprintf("%s-amqps-0-svc.%s", artemisName, namespace),
		fmt.Sprintf("%s-hdls-svc.%s", artemisName, namespace),
		fmt.Sprintf("%s-amqps-0-svc", artemisName),
		fmt.Sprintf("%s-hdls-svc", artemisName),
		artemisName,
		"localhost",
	)
	return names
}

func BuildArtemisSSLSecretData(keystoreJKS, truststoreJKS []byte, keystorePassword string) map[string][]byte {
	return map[string][]byte{
		"broker.ks":          keystoreJKS,
		"client.ts":          truststoreJKS,
		"keyStorePassword":   []byte(keystorePassword),
		"trustStorePassword": []byte(keystorePassword),
	}
}
