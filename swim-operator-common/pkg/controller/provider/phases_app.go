package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/helpers"
	"github.com/swim-developer/swim-operator-common/pkg/labels"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func ReconcileProviderAppConfigPhase(ctx context.Context, cfg ProviderPhaseConfig, clusterDomain string) (string, ctrl.Result, error) {
	mb := cfg.ManagedByValue
	if err := providerReconcileConfigMap(ctx, cfg, BuildProviderAppConfigMap(cfg.BuildParams, mb, clusterDomain)); err != nil {
		return "", ctrl.Result{}, err
	}
	if err := providerReconcileSecret(ctx, cfg, BuildProviderAppSecret(cfg.BuildParams, mb)); err != nil {
		return "", ctrl.Result{}, err
	}
	if err := providerReconcileSecret(ctx, cfg, BuildProviderAppOIDCSecret(cfg.BuildParams, mb)); err != nil {
		return "", ctrl.Result{}, err
	}
	if res, err := providerReconcileCertificate(ctx, cfg, BuildProviderAppServerCertificate(cfg.BuildParams, mb, clusterDomain)); err != nil {
		return "", ctrl.Result{}, err
	} else if res.RequeueAfter > 0 || res.Requeue {
		return "", res, nil
	}
	if err := ReconcileProviderCABundlePhase(ctx, cfg); err != nil {
		return "", ctrl.Result{}, err
	}
	cm := BuildProviderAppConfigMap(cfg.BuildParams, mb, clusterDomain)
	sec := BuildProviderAppSecret(cfg.BuildParams, mb)
	oidc := BuildProviderAppOIDCSecret(cfg.BuildParams, mb)
	return resources.ComputeConfigHash(cm, sec, oidc), ctrl.Result{}, nil
}

func ReconcileProviderCABundlePhase(ctx context.Context, cfg ProviderPhaseConfig) error {
	p := cfg.BuildParams
	name := p.Name
	ns := p.Namespace
	mb := cfg.ManagedByValue
	certSecret := &corev1.Secret{}
	if err := cfg.Client.Get(ctx, client.ObjectKey{
		Name:      fmt.Sprintf(constants.ServerTLSSuffix, name),
		Namespace: ns,
	}, certSecret); err != nil {
		return fmt.Errorf("failed to get certificate secret: %w", err)
	}
	caCrt, ok := certSecret.Data["ca.crt"]
	if !ok {
		return fmt.Errorf("ca.crt not found in certificate secret")
	}
	caBundle := resources.ConfigMap(fmt.Sprintf("%s-ca-bundle", name), ns, labels.StandardLabels(name, "provider", name, mb), map[string]string{"ca.crt": string(caCrt)})
	return providerReconcileConfigMap(ctx, cfg, caBundle)
}

func ReconcileProviderAppDeploymentPhase(ctx context.Context, cfg ProviderPhaseConfig, clusterDomain, configHash string) error {
	logger := log.FromContext(ctx)
	mb := cfg.ManagedByValue
	if err := providerReconcileDeployment(ctx, cfg, BuildProviderAppDeployment(cfg.BuildParams, mb, configHash)); err != nil {
		return err
	}
	if err := providerReconcileService(ctx, cfg, BuildProviderAppService(cfg.BuildParams, mb)); err != nil {
		return err
	}
	if cfg.BuildParams.Strategy.ServiceMonitorEnabled() {
		if err := providerReconcileServiceMonitor(ctx, cfg, BuildProviderAppServiceMonitor(cfg.BuildParams, mb)); err != nil {
			logger.Info("ServiceMonitor creation failed (observability stack may not be available)", "error", err)
		}
	}
	if err := providerReconcileHPA(ctx, cfg, BuildProviderHPA(cfg.BuildParams, mb)); err != nil {
		return err
	}
	if cfg.ReconcileAppExposure != nil {
		return cfg.ReconcileAppExposure(ctx, clusterDomain)
	}
	return nil
}

func CheckProviderAppReadinessPhase(ctx context.Context, cfg ProviderPhaseConfig) (ctrl.Result, error) {
	p := cfg.BuildParams
	if !isProviderAppReady(ctx, cfg) {
		applyProviderCondition(ctx, cfg, "ProviderReady", metav1.ConditionFalse, "NotReady", p.Strategy.NotReadyMessage())
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}
	applyProviderCondition(ctx, cfg, "ProviderReady", metav1.ConditionTrue, "Ready", p.Strategy.ReadyMessage())
	return ctrl.Result{}, nil
}

func isProviderAppReady(ctx context.Context, cfg ProviderPhaseConfig) bool {
	return helpers.IsPodReady(ctx, cfg.Client, cfg.BuildParams.Namespace, map[string]string{"app": cfg.BuildParams.Name})
}
