package reconciler

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"software.sslmate.com/src/go-pkcs12"

	"github.com/swim-developer/swim-operator-common/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func GetCertificateSecretData(ctx context.Context, c client.Client, namespace, certSecretName string) (tlsCrt, tlsKey, caCrt []byte, err error) {
	certSecret := &corev1.Secret{}
	if err := c.Get(ctx, types.NamespacedName{Name: certSecretName, Namespace: namespace}, certSecret); err != nil {
		if !errors.IsNotFound(err) {
			log.FromContext(ctx).Error(err, "Failed to get certificate secret", "secret", certSecretName)
		}
		return nil, nil, nil, err
	}

	tlsCrt = certSecret.Data["tls.crt"]
	tlsKey = certSecret.Data["tls.key"]
	caCrt = certSecret.Data["ca.crt"]

	if len(tlsCrt) == 0 || len(tlsKey) == 0 || len(caCrt) == 0 {
		return nil, nil, nil, fmt.Errorf("certificate secret %s is not ready yet (missing tls.crt, tls.key, or ca.crt)", certSecretName)
	}

	return tlsCrt, tlsKey, caCrt, nil
}

func ParseCertAndKeyFromPEM(tlsCrt, tlsKey []byte, certSecretName string) (*x509.Certificate, crypto.PrivateKey, error) {
	tlsCrtPEM, _ := pem.Decode(tlsCrt)
	if tlsCrtPEM == nil {
		return nil, nil, fmt.Errorf("failed to decode tls.crt from secret %s", certSecretName)
	}

	tlsKeyPEM, _ := pem.Decode(tlsKey)
	if tlsKeyPEM == nil {
		return nil, nil, fmt.Errorf("failed to decode tls.key from secret %s", certSecretName)
	}

	serverCert, err := x509.ParseCertificate(tlsCrtPEM.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	var serverKey crypto.PrivateKey
	if tlsKeyPEM.Type == "RSA PRIVATE KEY" {
		serverKey, err = x509.ParsePKCS1PrivateKey(tlsKeyPEM.Bytes)
	} else if tlsKeyPEM.Type == "PRIVATE KEY" {
		serverKey, err = x509.ParsePKCS8PrivateKey(tlsKeyPEM.Bytes)
	} else {
		return nil, nil, fmt.Errorf("unsupported private key type: %s", tlsKeyPEM.Type)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return serverCert, serverKey, nil
}

func ParseCAFromPEM(caCrt []byte, certSecretName string) (*x509.Certificate, error) {
	caCertPEM, _ := pem.Decode(caCrt)
	if caCertPEM == nil {
		return nil, fmt.Errorf("failed to decode ca.crt from secret %s", certSecretName)
	}

	caCertObj, err := x509.ParseCertificate(caCertPEM.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	return caCertObj, nil
}

func CreateArtemisKeystoreAndTruststore(serverKey crypto.PrivateKey, serverCert, caCert *x509.Certificate, keystorePassword string) (keystoreData, truststoreData []byte, err error) {
	keystoreData, err = pkcs12.Encode(rand.Reader, serverKey, serverCert, []*x509.Certificate{}, keystorePassword)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode keystore (broker.ks): %w", err)
	}

	truststoreData, err = pkcs12.EncodeTrustStore(rand.Reader, []*x509.Certificate{caCert}, keystorePassword)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode truststore (client.ts): %w", err)
	}

	return keystoreData, truststoreData, nil
}

type ArtemisSSLSecretFromPEMInput struct {
	CertSecretName   string
	TargetSecretName string
	KeystorePassword string
	Labels           map[string]string
}

func ReconcileArtemisSSLSecretFromPEM(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, in ArtemisSSLSecretFromPEMInput) error {

	logger := log.FromContext(ctx)

	existingSecret := &corev1.Secret{}
	err := c.Get(ctx, types.NamespacedName{Name: in.TargetSecretName, Namespace: owner.GetNamespace()}, existingSecret)
	if err == nil {
		return nil
	}
	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check existing SSL secret %s: %w", in.TargetSecretName, err)
	}

	tlsCrt, tlsKey, caCrt, err := GetCertificateSecretData(ctx, c, owner.GetNamespace(), in.CertSecretName)
	if err != nil {
		return err
	}

	serverCert, serverKey, err := ParseCertAndKeyFromPEM(tlsCrt, tlsKey, in.CertSecretName)
	if err != nil {
		return err
	}

	caCertObj, err := ParseCAFromPEM(caCrt, in.CertSecretName)
	if err != nil {
		return err
	}

	keystoreData, truststoreData, err := CreateArtemisKeystoreAndTruststore(serverKey, serverCert, caCertObj, in.KeystorePassword)
	if err != nil {
		return err
	}

	sslSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      in.TargetSecretName,
			Namespace: owner.GetNamespace(),
			Labels:    in.Labels,
		},
		Type: corev1.SecretTypeOpaque,
		Data: resources.BuildArtemisSSLSecretData(keystoreData, truststoreData, in.KeystorePassword),
	}

	if err := ctrl.SetControllerReference(owner, sslSecret, scheme); err != nil {
		return fmt.Errorf("failed to set controller reference on SSL secret %s: %w", in.TargetSecretName, err)
	}

	logger.Info("Creating SSL secret from PEM", "secret", in.TargetSecretName)
	if err := c.Create(ctx, sslSecret); err != nil {
		return fmt.Errorf("failed to create SSL secret %s: %w", in.TargetSecretName, err)
	}

	return nil
}

type ArtemisSSLSecretFromJKSInput struct {
	SourceSecretName string
	TargetSecretName string
	KeystorePassword string
	Labels           map[string]string
}

func ReconcileArtemisSSLSecretFromJKS(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, in ArtemisSSLSecretFromJKSInput) error {

	logger := log.FromContext(ctx)

	sourceSecret := &corev1.Secret{}
	err := c.Get(ctx, types.NamespacedName{Name: in.SourceSecretName, Namespace: owner.GetNamespace()}, sourceSecret)
	if errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	keystoreJKS, hasKeystore := sourceSecret.Data["keystore.jks"]
	truststoreJKS, hasTruststore := sourceSecret.Data["truststore.jks"]
	if !hasKeystore || !hasTruststore {
		return nil
	}

	targetSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      in.TargetSecretName,
			Namespace: owner.GetNamespace(),
			Labels:    in.Labels,
		},
		Type: corev1.SecretTypeOpaque,
		Data: resources.BuildArtemisSSLSecretData(keystoreJKS, truststoreJKS, in.KeystorePassword),
	}

	if err := ctrl.SetControllerReference(owner, targetSecret, scheme); err != nil {
		return err
	}

	existing := &corev1.Secret{}
	err = c.Get(ctx, client.ObjectKeyFromObject(targetSecret), existing)
	if errors.IsNotFound(err) {
		logger.Info("Creating SSL secret from JKS", "secret", in.TargetSecretName)
		return c.Create(ctx, targetSecret)
	} else if err != nil {
		return err
	}

	targetSecret.SetResourceVersion(existing.GetResourceVersion())
	targetSecret.SetUID(existing.GetUID())
	return c.Update(ctx, targetSecret)
}
