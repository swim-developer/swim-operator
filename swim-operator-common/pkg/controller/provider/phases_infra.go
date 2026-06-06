package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
)

func reconcileProviderArtemisSecretsPhase(ctx context.Context, cfg ProviderPhaseConfig, artemisIngressHost string) (ctrl.Result, error) {
	mb := cfg.ManagedByValue
	if err := providerReconcileSecret(ctx, cfg, BuildProviderArtemisKeystoreSecret(cfg.BuildParams, mb)); err != nil {
		return ctrl.Result{}, err
	}
	if err := providerReconcileSecret(ctx, cfg, BuildProviderArtemisCredentialsSecret(cfg.BuildParams, mb)); err != nil {
		return ctrl.Result{}, err
	}
	if err := providerReconcileSecret(ctx, cfg, BuildProviderArtemisOIDCSecret(cfg.BuildParams, mb)); err != nil {
		return ctrl.Result{}, err
	}
	if err := providerReconcileUnmanagedSecret(ctx, cfg, BuildProviderArtemisAddressBPSecret(cfg.BuildParams, mb)); err != nil {
		return ctrl.Result{}, err
	}
	if err := providerReconcileUnmanagedSecret(ctx, cfg, BuildProviderArtemisSecurityBPSecret(cfg.BuildParams, mb)); err != nil {
		return ctrl.Result{}, err
	}
	if res, err := providerReconcileCertificate(ctx, cfg, BuildProviderArtemisCertificate(cfg.BuildParams, mb, artemisIngressHost)); err != nil {
		return ctrl.Result{}, err
	} else if res.RequeueAfter > 0 || res.Requeue {
		return res, nil
	}
	if err := reconcileProviderArtemisSSLSecretsPhase(ctx, cfg); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func reconcileProviderArtemisSSLSecretsPhase(ctx context.Context, cfg ProviderPhaseConfig) error {
	artemisName := ProviderArtemisName(cfg.BuildParams)
	certSecretName := fmt.Sprintf("%s-amqp-tls", artemisName)
	sslSecretName := fmt.Sprintf(constants.SSLSecretSuffix, artemisName)
	consoleSecretName := fmt.Sprintf("%s-console-ssl-secret", artemisName)
	if err := ReconcileProviderArtemisSSLSecret(ctx, cfg.Client, cfg.Scheme, cfg.Owner, ProviderArtemisSSLSecretInput{
		P:                cfg.BuildParams,
		ManagedBy:        cfg.ManagedByValue,
		CertSecretName:   certSecretName,
		TargetSecretName: sslSecretName,
	}); err != nil {
		return fmt.Errorf("failed to reconcile AMQPS SSL secret: %w", err)
	}
	if err := ReconcileProviderArtemisSSLSecret(ctx, cfg.Client, cfg.Scheme, cfg.Owner, ProviderArtemisSSLSecretInput{
		P:                cfg.BuildParams,
		ManagedBy:        cfg.ManagedByValue,
		CertSecretName:   certSecretName,
		TargetSecretName: consoleSecretName,
	}); err != nil {
		return fmt.Errorf("failed to reconcile console SSL secret: %w", err)
	}
	return nil
}

func reconcileProviderArtemisBrokerPhase(ctx context.Context, cfg ProviderPhaseConfig, artemisIngressHost string) error {
	mb := cfg.ManagedByValue
	if err := ReconcileSharedProviderUnstructured(ctx, cfg, BuildProviderArtemisBroker(cfg.BuildParams, mb, artemisIngressHost)); err != nil {
		return err
	}
	return providerReconcileService(ctx, cfg, BuildProviderArtemisJMXService(cfg.BuildParams, mb))
}

func waitForProviderArtemisReadyPhase(ctx context.Context, cfg ProviderPhaseConfig) (ctrl.Result, error) {
	if !isProviderArtemisReady(ctx, cfg) {
		applyProviderCondition(ctx, cfg, "ArtemisReady", metav1.ConditionFalse, "NotReady", "Artemis broker not ready")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}
	applyProviderCondition(ctx, cfg, "ArtemisReady", metav1.ConditionTrue, "Ready", "Artemis broker is ready")
	return ctrl.Result{}, nil
}

func ReconcileProviderArtemisPhase(ctx context.Context, cfg ProviderPhaseConfig, artemisIngressHost string) (ctrl.Result, error) {
	if res, err := reconcileProviderArtemisSecretsPhase(ctx, cfg, artemisIngressHost); err != nil || res.RequeueAfter > 0 || res.Requeue {
		return res, err
	}
	if err := reconcileProviderArtemisBrokerPhase(ctx, cfg, artemisIngressHost); err != nil {
		return ctrl.Result{}, err
	}
	return waitForProviderArtemisReadyPhase(ctx, cfg)
}

func isProviderArtemisReady(ctx context.Context, cfg ProviderPhaseConfig) bool {
	artemisName := ProviderArtemisName(cfg.BuildParams)
	return helpers.IsPodReady(ctx, cfg.Client, cfg.BuildParams.Namespace, map[string]string{"ActiveMQArtemis": artemisName})
}

func ReconcileProviderKafkaPhase(ctx context.Context, cfg ProviderPhaseConfig) (ctrl.Result, error) {
	p := cfg.BuildParams
	k := p.Kafka
	if !k.Enabled || k.DeploymentMode != "managed" {
		return ctrl.Result{}, nil
	}
	resolve := cfg.ResolveClusterDomain
	if resolve == nil {
		resolve = func(_ context.Context, spec, _ string) string { return spec }
	}
	reconcileTopic := func(u *unstructured.Unstructured) error {
		return providerReconcileUnstructured(ctx, cfg, u)
	}
	if err := ReconcileProviderManagedKafka(ctx, cfg.Client, p, cfg.ManagedByLabel, cfg.ManagedByValue, reconcileTopic, resolve); err != nil {
		return ctrl.Result{}, err
	}
	if !helpers.IsKafkaClusterReady(ctx, cfg.Client, p.Namespace, "kafka") {
		applyProviderCondition(ctx, cfg, "KafkaReady", metav1.ConditionFalse, "NotReady", "Kafka cluster not ready")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}
	applyProviderCondition(ctx, cfg, "KafkaReady", metav1.ConditionTrue, "Ready", "Kafka cluster is ready")
	return ctrl.Result{}, nil
}
