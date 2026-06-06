package consumer

import (
	"context"

	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const ManagedByLabelKey = "apps.swim-developer.github.io/managed-by"

type ConsumerFlavor int

const (
	ConsumerFlavorDnotam ConsumerFlavor = iota
	ConsumerFlavorEd254
	ConsumerFlavorFfice
)

type ConsumerBuildParams struct {
	Flavor              ConsumerFlavor
	Name                string
	Namespace           string
	GlobalClusterDomain string
	Kafka               commonapi.KafkaSpec
	CertManager         commonapi.CertManagerSpec
	Client              commonapi.ClientSpec
	Consumer            commonapi.Ed254ConsumerAppSpec
	FficeConsumer       commonapi.FficeConsumerAppSpec
	HPA                 commonapi.HPAConfig
}

type ConsumerSecretsBundle struct {
	Keystore   *corev1.Secret
	Providers  *corev1.Secret
	KafkaCreds *corev1.Secret
}

type ConsumerPhaseConfig struct {
	Client               client.Client
	Scheme               *runtime.Scheme
	Owner                client.Object
	Request              ctrl.Request
	FinalizerName        string
	CRKind               string
	BuildParams          ConsumerBuildParams
	KafkaTopics          []string
	ManagedByLabel       string
	ManagedByValue       string
	ResolveClusterDomain func(ctx context.Context, specDomain, namespace string) string
	RemoveFinalizer      func(ctx context.Context) error
	ApplyStatus          func(ctx context.Context, condition metav1.Condition) error
}
