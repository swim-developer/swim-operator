package cv

import (
	"context"
	"fmt"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/labels"
	commonreconciler "github.com/swim-developer/swim-operator-common/pkg/reconciler"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func artemisCVParams(spec commonapi.SwimDnotamConsumerValidatorSpec, crName, namespace, ingressHost, exposeMode, managedBy string) resources.ArtemisCVParams {
	artemisName := fmt.Sprintf(constants.ArtemisSuffix, crName)
	lbl := labels.StandardLabels(artemisName, "artemis", crName, managedBy)
	amqpEnabled := spec.Artemis.Acceptors.Amqp.Enabled == nil || *spec.Artemis.Acceptors.Amqp.Enabled
	amqpExpose := spec.Artemis.Acceptors.Amqp.Expose != nil && *spec.Artemis.Acceptors.Amqp.Expose
	consoleEnabled := spec.Artemis.Console.Enabled != nil && *spec.Artemis.Console.Enabled
	consoleExpose := spec.Artemis.Console.Expose != nil && *spec.Artemis.Console.Expose
	consoleSSL := spec.Artemis.Console.SSLEnabled != nil && *spec.Artemis.Console.SSLEnabled
	return resources.ArtemisCVParams{
		CRName:           crName,
		Namespace:        namespace,
		Labels:           lbl,
		ArtemisName:      artemisName,
		AdminUser:        spec.Artemis.Credentials.AdminUser,
		AdminPassword:    spec.Artemis.Credentials.AdminPassword,
		ClusterUser:      spec.Artemis.Credentials.ClusterUser,
		ClusterPassword:  spec.Artemis.Credentials.ClusterPassword,
		KeystorePassword: spec.Artemis.CertManager.KeystorePassword,
		IssuerName:       spec.CertManager.IssuerName,
		IssuerKind:       spec.CertManager.IssuerKind,
		IngressHost:      ingressHost,
		ExposeMode:       exposeMode,
		Size:             spec.Artemis.Broker.Size,
		BrokerProperties: spec.Artemis.BrokerProperties,
		ConsoleEnabled:   consoleEnabled,
		ConsoleExpose:    consoleExpose,
		ConsoleSSL:       consoleSSL,
		AmqpEnabled:      amqpEnabled,
		AmqpPort:         spec.Artemis.Acceptors.Amqp.Port,
		AmqpExpose:       amqpExpose,
	}
}

func BuildCVArtemisCredentialsSecret(p CVBuildParams, managedBy string) *corev1.Secret {
	return resources.BuildArtemisCVCredentialsSecret(artemisCVParams(p.Spec, p.CRName, p.Namespace, "", p.ArtemisExposeMode, managedBy))
}

func BuildCVArtemisKeystoreSecret(p CVBuildParams, managedBy string) *corev1.Secret {
	return resources.BuildArtemisCVKeystoreSecret(artemisCVParams(p.Spec, p.CRName, p.Namespace, "", p.ArtemisExposeMode, managedBy))
}

func BuildCVArtemisCertificate(p CVBuildParams, managedBy string, ingressHost string) *certmanagerv1.Certificate {
	return resources.BuildArtemisCVCertificate(artemisCVParams(p.Spec, p.CRName, p.Namespace, ingressHost, p.ArtemisExposeMode, managedBy))
}

func BuildCVArtemisBroker(p CVBuildParams, managedBy string, ingressHost string) *unstructured.Unstructured {
	return resources.BuildArtemisCVBroker(artemisCVParams(p.Spec, p.CRName, p.Namespace, ingressHost, p.ArtemisExposeMode, managedBy))
}

type CVArtemisSSLReconcileParams struct {
	Spec      commonapi.SwimDnotamConsumerValidatorSpec
	CRName    string
	Namespace string
	ManagedBy string
}

func ReconcileCVArtemisSSLSecret(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, p CVArtemisSSLReconcileParams) error {
	artemisName := fmt.Sprintf(constants.ArtemisSuffix, p.CRName)
	sourceSecretName := fmt.Sprintf("%s-amqp-tls", artemisName)
	targetSecretName := fmt.Sprintf(constants.SSLSecretSuffix, artemisName)
	keystorePassword := resources.StrDefault(p.Spec.Artemis.CertManager.KeystorePassword, "changeit")
	lbls := labels.StandardLabels(artemisName, "artemis", p.CRName, p.ManagedBy)
	return commonreconciler.ReconcileArtemisSSLSecretFromJKS(ctx, c, scheme, owner, commonreconciler.ArtemisSSLSecretFromJKSInput{
		SourceSecretName: sourceSecretName,
		TargetSecretName: targetSecretName,
		KeystorePassword: keystorePassword,
		Labels:           lbls,
	})
}
