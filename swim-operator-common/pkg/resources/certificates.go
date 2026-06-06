package resources

import (
	"fmt"
	"strings"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	svcClusterLocalFmt           = "%s.%s.svc.cluster.local"
	svcShortFmt                  = "%s.%s.svc"
	nameNamespaceFmt             = "%s.%s"
	defaultCertificateIssuerName = "swim-ca-issuer"
	defaultCertificateIssuerKind = "ClusterIssuer"
	defaultKeystoreSecretKey     = "password"
	localhostLoopback            = "localhost"
)

type CertificateParams struct {
	Name               string
	Namespace          string
	Labels             map[string]string
	SecretName         string
	IssuerName         string
	IssuerKind         string
	CommonName         string
	DNSNames           []string
	Usages             []certmanagerv1.KeyUsage
	WithPKCS12         bool
	WithJKS            bool
	KeystoreSecretName string
	KeystoreSecretKey  string
}

func BuildCertificate(p CertificateParams) *certmanagerv1.Certificate {
	cert := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{Name: p.Name, Namespace: p.Namespace, Labels: p.Labels},
		Spec: certmanagerv1.CertificateSpec{
			SecretName: p.SecretName,
			IssuerRef: cmmeta.ObjectReference{
				Name: StrDefault(p.IssuerName, defaultCertificateIssuerName),
				Kind: StrDefault(p.IssuerKind, defaultCertificateIssuerKind),
			},
			CommonName: p.CommonName,
			DNSNames:   p.DNSNames,
			Usages:     p.Usages,
			PrivateKey: &certmanagerv1.CertificatePrivateKey{
				Algorithm:      certmanagerv1.RSAKeyAlgorithm,
				Size:           2048,
				RotationPolicy: certmanagerv1.RotationPolicyAlways,
			},
		},
	}

	if p.WithPKCS12 || p.WithJKS {
		cert.Spec.Keystores = &certmanagerv1.CertificateKeystores{}
		pwRef := cmmeta.SecretKeySelector{
			LocalObjectReference: cmmeta.LocalObjectReference{Name: p.KeystoreSecretName},
			Key:                  StrDefault(p.KeystoreSecretKey, defaultKeystoreSecretKey),
		}
		if p.WithPKCS12 {
			cert.Spec.Keystores.PKCS12 = &certmanagerv1.PKCS12Keystore{Create: true, PasswordSecretRef: pwRef}
		}
		if p.WithJKS {
			cert.Spec.Keystores.JKS = &certmanagerv1.JKSKeystore{Create: true, PasswordSecretRef: pwRef}
		}
	}

	return cert
}

func BuildServerCertificate(name, namespace, crName string, labels map[string]string, issuerName, issuerKind, host string) *certmanagerv1.Certificate {
	apiHost := strings.Replace(host, name, name+"-api", 1)
	apiName := name + "-api"
	return BuildCertificate(CertificateParams{
		Name:       fmt.Sprintf("%s-server-cert", name),
		Namespace:  namespace,
		Labels:     labels,
		SecretName: fmt.Sprintf("%s-server-tls", name),
		IssuerName: issuerName,
		IssuerKind: issuerKind,
		CommonName: name,
		DNSNames: []string{
			host,
			apiHost,
			fmt.Sprintf(svcClusterLocalFmt, name, namespace),
			fmt.Sprintf(svcClusterLocalFmt, apiName, namespace),
			fmt.Sprintf(svcShortFmt, name, namespace),
			fmt.Sprintf(svcShortFmt, apiName, namespace),
			fmt.Sprintf(nameNamespaceFmt, name, namespace),
			fmt.Sprintf(nameNamespaceFmt, apiName, namespace),
			name,
			apiName,
			localhostLoopback,
		},
		Usages: []certmanagerv1.KeyUsage{certmanagerv1.UsageServerAuth, certmanagerv1.UsageClientAuth},
	})
}

func BuildClientCertificate(name, namespace string, labels map[string]string, issuerName, issuerKind string) *certmanagerv1.Certificate {
	clientName := fmt.Sprintf("%s-client", name)
	return BuildCertificate(CertificateParams{
		Name:       fmt.Sprintf("%s-client-cert", name),
		Namespace:  namespace,
		Labels:     labels,
		SecretName: fmt.Sprintf("%s-client-tls", name),
		IssuerName: issuerName,
		IssuerKind: issuerKind,
		CommonName: clientName,
		DNSNames: []string{
			clientName,
			fmt.Sprintf(svcClusterLocalFmt, clientName, namespace),
			fmt.Sprintf(svcShortFmt, clientName, namespace),
			fmt.Sprintf(nameNamespaceFmt, clientName, namespace),
			localhostLoopback,
		},
		Usages: []certmanagerv1.KeyUsage{certmanagerv1.UsageDigitalSignature, certmanagerv1.UsageKeyEncipherment, certmanagerv1.UsageClientAuth},
	})
}

func BuildMTLSCertificate(name, namespace string, labels map[string]string, issuerName, issuerKind, keystoreSecretName string) *certmanagerv1.Certificate {
	return BuildCertificate(CertificateParams{
		Name:               fmt.Sprintf("%s-mtls-cert", name),
		Namespace:          namespace,
		Labels:             labels,
		SecretName:         fmt.Sprintf("%s-mtls", name),
		IssuerName:         issuerName,
		IssuerKind:         issuerKind,
		CommonName:         name,
		Usages:             []certmanagerv1.KeyUsage{certmanagerv1.UsageClientAuth},
		WithPKCS12:         true,
		WithJKS:            true,
		KeystoreSecretName: keystoreSecretName,
		KeystoreSecretKey:  defaultKeystoreSecretKey,
	})
}
