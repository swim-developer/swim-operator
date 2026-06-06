package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func ReconcileProviderRBACPhase(ctx context.Context, cfg ProviderPhaseConfig) error {
	mb := cfg.ManagedByValue
	p := cfg.BuildParams
	if err := providerReconcileServiceAccount(ctx, cfg, BuildProviderServiceAccount(p, mb)); err != nil {
		return err
	}
	if err := providerReconcileRole(ctx, cfg, BuildProviderRole(p, mb)); err != nil {
		return err
	}
	return providerReconcileRoleBinding(ctx, cfg, BuildProviderRoleBinding(p, mb))
}

func ReconcileProviderPostgresPhase(ctx context.Context, cfg ProviderPhaseConfig) (ctrl.Result, error) {
	mb := cfg.ManagedByValue
	if err := providerReconcileSecret(ctx, cfg, BuildProviderPostgresSecret(cfg.BuildParams, mb)); err != nil {
		return ctrl.Result{}, err
	}
	if err := providerReconcilePVC(ctx, cfg, BuildProviderPostgresPVC(cfg.BuildParams, mb)); err != nil {
		return ctrl.Result{}, err
	}
	if err := providerReconcileStatefulSet(ctx, cfg, BuildProviderPostgresStatefulSet(cfg.BuildParams, mb)); err != nil {
		return ctrl.Result{}, err
	}
	if err := providerReconcileService(ctx, cfg, BuildProviderPostgresService(cfg.BuildParams, mb)); err != nil {
		return ctrl.Result{}, err
	}
	if !isProviderPostgresReady(ctx, cfg) {
		applyProviderCondition(ctx, cfg, "PostgresReady", metav1.ConditionFalse, "NotReady", "Postgres pod not ready")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}
	applyProviderCondition(ctx, cfg, "PostgresReady", metav1.ConditionTrue, "Ready", "Postgres pod is ready")
	return ctrl.Result{}, nil
}

func isProviderPostgresReady(ctx context.Context, cfg ProviderPhaseConfig) bool {
	name := cfg.BuildParams.Name
	ns := cfg.BuildParams.Namespace
	appLabel := fmt.Sprintf(constants.PostgresSuffix, name)
	return helpers.IsPodReady(ctx, cfg.Client, ns, map[string]string{"app": appLabel})
}
