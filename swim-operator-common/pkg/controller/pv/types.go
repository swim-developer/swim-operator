package pv

import (
	"context"

	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PVBuildParams struct {
	CRName               string
	Namespace            string
	Spec                 commonapi.ProviderValidatorBaseSpec
	CertManager          commonapi.CertManagerSpec
	DefaultImage         string
	IngressEnabled       bool
	IngressHost          string
	IngressTLSSecretName string
	IngressAnnotations   map[string]string
	RouteHost            string
}

type PVPhaseConfig struct {
	Client               client.Client
	Scheme               *runtime.Scheme
	Owner                client.Object
	Request              ctrl.Request
	BuildParams          PVBuildParams
	ManagedByLabel       string
	ManagedByValue       string
	ResolveClusterDomain func(ctx context.Context, specDomain, namespace string) string
	ApplyStatus          func(ctx context.Context, condition metav1.Condition) error
	ReconcilePreAppTLS   func(ctx context.Context) (ctrl.Result, error)
	ReconcileAppExposure func(ctx context.Context) error
}
