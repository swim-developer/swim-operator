package cv

import (
	"context"

	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CVFlavor int

const (
	CVFlavorDnotam CVFlavor = iota
	CVFlavorEd254
	CVFlavorFfice
)

// CvSpecFromEd254 converts an ED-254 consumer validator spec to the equivalent DNOTAM spec.
// The two types are structurally identical – this is a safe type cast.
func CvSpecFromEd254(s commonapi.SwimEd254ConsumerValidatorSpec) commonapi.SwimDnotamConsumerValidatorSpec {
	return commonapi.SwimDnotamConsumerValidatorSpec(s)
}

// CvSpecFromFfice converts a FF-ICE consumer validator spec to the equivalent DNOTAM spec.
// SwimFficeConsumerValidatorSpec is a type alias for SwimEd254ConsumerValidatorSpec, which is
// structurally identical to SwimDnotamConsumerValidatorSpec – this is a safe type cast.
func CvSpecFromFfice(s commonapi.SwimFficeConsumerValidatorSpec) commonapi.SwimDnotamConsumerValidatorSpec {
	return commonapi.SwimDnotamConsumerValidatorSpec(s)
}

type CVIngressParams struct {
	Enabled             bool
	HostOverride        string
	ArtemisHostOverride string
	APIHostOverride     string
	TLSSecretName       string
	Annotations         map[string]string
}

type CVBuildParams struct {
	Flavor            CVFlavor
	CRName            string
	Namespace         string
	Spec              commonapi.SwimDnotamConsumerValidatorSpec
	DefaultImageRepo  string
	DefaultDatabase   string
	Ingress           CVIngressParams
	ArtemisExposeMode string
}

type CVPhaseConfig struct {
	Client               client.Client
	Scheme               *runtime.Scheme
	Owner                client.Object
	Request              ctrl.Request
	FinalizerName        string
	CRKind               string
	BuildParams          CVBuildParams
	ManagedByLabel       string
	ManagedByValue       string
	ResolveClusterDomain func(ctx context.Context, specDomain, namespace string) string
	RemoveFinalizer      func(ctx context.Context) error
	ApplyStatus          func(ctx context.Context, condition metav1.Condition) error
	FetchLatest          func(ctx context.Context) (client.Object, error)
	ReconcileAppExposure func(ctx context.Context) error
}
