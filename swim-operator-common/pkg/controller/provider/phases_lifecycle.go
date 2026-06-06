package provider

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	commonreconciler "github.com/swim-developer/swim-operator-common/pkg/reconciler"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func applyProviderCondition(ctx context.Context, cfg ProviderPhaseConfig, conditionType string, status metav1.ConditionStatus, reason, message string) {
	if cfg.ApplyStatus == nil {
		return
	}
	_ = cfg.ApplyStatus(ctx, metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})
}

func FinalizeProviderResources(ctx context.Context, cfg ProviderPhaseConfig) error {
	logger := log.FromContext(ctx)
	p := cfg.BuildParams
	k := p.Kafka
	if k.Enabled && k.DeploymentMode == "managed" {
		topics := []*unstructured.Unstructured{
			BuildProviderKafkaTopicAll(p, cfg.ManagedByValue),
			BuildProviderKafkaTopicDLQ(p, cfg.ManagedByValue),
		}
		for _, resource := range topics {
			if err := cfg.Client.Delete(ctx, resource); err != nil && !errors.IsNotFound(err) {
				logger.Error(err, "Failed to delete Kafka topic", "name", resource.GetName())
			}
		}
	}
	artemisName := ProviderArtemisName(p)
	commonreconciler.CleanupServiceBrokerProperties(ctx, cfg.Client, p.Namespace, artemisName, p.ArtemisBrokerCleanupPrefix())
	commonreconciler.CleanupSharedInfraIfLast(ctx, cfg.Client, p.Namespace, cfg.CRKind, p.Name, artemisName, k.Enabled)
	return nil
}

func HandleProviderFinalization(ctx context.Context, cfg ProviderPhaseConfig) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	metaObj, ok := cfg.Owner.(metav1.Object)
	if !ok {
		return ctrl.Result{}, fmt.Errorf("owner does not implement metav1.Object")
	}
	ts := metaObj.GetDeletionTimestamp()
	if ts == nil || ts.IsZero() {
		return ctrl.Result{}, nil
	}
	if controllerutil.ContainsFinalizer(cfg.Owner, cfg.FinalizerName) {
		logger.Info("Executing provider finalizer cleanup", "name", metaObj.GetName())
		if err := FinalizeProviderResources(ctx, cfg); err != nil {
			return ctrl.Result{}, err
		}
		if cfg.RemoveFinalizer != nil {
			if err := cfg.RemoveFinalizer(ctx); err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

func EnsureProviderFinalizer(ctx context.Context, cfg ProviderPhaseConfig) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(cfg.Owner, cfg.FinalizerName) {
		return ctrl.Result{}, nil
	}
	controllerutil.AddFinalizer(cfg.Owner, cfg.FinalizerName)
	if err := cfg.Client.Update(ctx, cfg.Owner); err != nil {
		if errors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}
	return ctrl.Result{Requeue: true}, nil
}

func ValidateProviderExternalKafkaPhase(ctx context.Context, cfg ProviderPhaseConfig) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	p := cfg.BuildParams
	k := p.Kafka
	if !k.Enabled || k.DeploymentMode != "external" {
		return ctrl.Result{}, nil
	}
	bootstrapServers := k.BootstrapServers
	if bootstrapServers == "" {
		return ctrl.Result{}, fmt.Errorf("bootstrapServers is required for external Kafka")
	}
	servers := strings.Split(bootstrapServers, ",")
	for _, server := range servers {
		server = strings.TrimSpace(server)
		logger.Info("Validating external Kafka connectivity", "server", server)
		conn, err := net.DialTimeout("tcp", server, 5*time.Second)
		if err != nil {
			logger.Error(err, "Failed to connect to external Kafka", "server", server)
			applyProviderCondition(ctx, cfg, "KafkaReady", metav1.ConditionFalse, "ValidationFailed", fmt.Sprintf("Cannot connect to external Kafka server %s: %v", server, err))
			return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}
		conn.Close()
	}
	return ctrl.Result{}, nil
}
