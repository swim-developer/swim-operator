package pv

import (
	"context"
	"fmt"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	commonreconciler "github.com/swim-developer/swim-operator-common/pkg/reconciler"
	ctrl "sigs.k8s.io/controller-runtime"
)

func ReconcilePVCertManagerMTLS(ctx context.Context, cfg PVPhaseConfig) (ctrl.Result, error) {
	p := cfg.BuildParams
	if !p.Spec.MTLS.Enabled || p.Spec.MTLS.CertsSecretName != "" || !p.CertManager.Enabled {
		return ctrl.Result{}, nil
	}
	mb := cfg.ManagedByValue
	if err := pvReconcileSecret(ctx, cfg, BuildPVKeystorePasswordSecret(p, mb)); err != nil {
		return ctrl.Result{}, err
	}
	cert := resources.BuildMTLSCertificate(
		p.CRName,
		p.Namespace,
		StandardLabels(p.CRName, mb),
		resources.StrDefault(p.CertManager.IssuerName, "swim-ca-issuer"),
		resources.StrDefault(p.CertManager.IssuerKind, "ClusterIssuer"),
		fmt.Sprintf(constants.KeystorePasswordSuffix, p.CRName),
	)
	return commonreconciler.ReconcileCertificate(ctx, cfg.Client, cfg.Scheme, cfg.Owner, cert)
}
