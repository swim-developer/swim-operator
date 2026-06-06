package provider

import (
	"context"

	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const ManagedByLabelKey = "apps.swim-developer.github.io/managed-by"

type ProviderExposureSpec struct {
	HTTPEdgeEnabled         bool
	HTTPSPassthroughEnabled bool
	HTTPSEdgeHost           string
	HTTPSPassthroughHost    string
}

type DnotamProviderPayload struct {
	Postgres commonapi.ProviderPostgresSpec
	Artemis  commonapi.ProviderArtemisSpec
	Provider commonapi.ProviderAppBaseSpec
	Exposure ProviderExposureSpec
}

type Ed254ProviderPayload struct {
	Postgres commonapi.Ed254PostgresSpec
	Artemis  commonapi.Ed254ArtemisSpec
	Provider commonapi.Ed254ProviderAppBaseSpec
	Exposure ProviderExposureSpec
}

type FficeProviderPayload struct {
	Postgres commonapi.Ed254PostgresSpec
	Artemis  commonapi.Ed254ArtemisSpec
	Provider commonapi.FficeProviderAppBaseSpec
	Exposure ProviderExposureSpec
}

type ProviderBuildParams struct {
	Name                       string
	Namespace                  string
	GlobalClusterDomain        string
	Kafka                      commonapi.KafkaSpec
	CertManager                commonapi.CertManagerSpec
	ArtemisExposeMode          string
	ArtemisIngressHostOverride string
	HPA                        commonapi.HPAConfig
	PostgresUpstream           bool
	ArtemisUpstream            bool
	Strategy                   ProviderStrategy
}

func (p ProviderBuildParams) CRKindForCleanup() string {
	return p.Strategy.CRKind()
}

func (p ProviderBuildParams) ArtemisBrokerCleanupPrefix() string {
	return p.Strategy.ArtemisBrokerCleanupPrefix()
}

type ProviderPhaseConfig struct {
	Client               client.Client
	Scheme               *runtime.Scheme
	Owner                client.Object
	Request              ctrl.Request
	FinalizerName        string
	CRKind               string
	BuildParams          ProviderBuildParams
	ManagedByLabel       string
	ManagedByValue       string
	ResolveClusterDomain func(ctx context.Context, specDomain, namespace string) string
	RemoveFinalizer      func(ctx context.Context) error
	ApplyStatus          func(ctx context.Context, condition metav1.Condition) error
	ReconcileAppExposure func(ctx context.Context, clusterDomain string) error
}
