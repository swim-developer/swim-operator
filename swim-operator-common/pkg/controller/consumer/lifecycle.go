package consumer

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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func HandleConsumerFinalization(ctx context.Context, cfg ConsumerPhaseConfig) (ctrl.Result, error) {
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
		logger.Info("Executing finalizer cleanup", "name", metaObj.GetName())
		if err := deleteConsumerExternalResources(ctx, cfg); err != nil {
			logger.Error(err, "Failed to delete external resources")
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

func EnsureConsumerFinalizer(ctx context.Context, cfg ConsumerPhaseConfig) (ctrl.Result, error) {
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

func deleteConsumerExternalResources(ctx context.Context, cfg ConsumerPhaseConfig) error {
	logger := log.FromContext(ctx)
	p := cfg.BuildParams
	kafkaMode := p.Kafka.DeploymentMode
	if kafkaMode == "" {
		kafkaMode = "managed"
	}
	if p.Kafka.Enabled && kafkaMode == "managed" {
		for _, topicName := range cfg.KafkaTopics {
			logger.Info("Deleting Kafka Topic", "topic", topicName)
			if err := consumerDeleteUnstructured(ctx, cfg, BuildConsumerKafkaTopic(p, cfg.ManagedByValue, topicName)); err != nil {
				logger.Error(err, "Failed to delete Kafka Topic", "topic", topicName)
			}
		}
	}
	commonreconciler.CleanupSharedInfraIfLast(ctx, cfg.Client, p.Namespace, cfg.CRKind, p.Name, "swim-artemis", p.Kafka.Enabled)
	return nil
}

func consumerDeleteUnstructured(ctx context.Context, cfg ConsumerPhaseConfig, obj *unstructured.Unstructured) error {
	err := cfg.Client.Delete(ctx, obj)
	return client.IgnoreNotFound(err)
}

func ValidateConsumerExternalKafka(ctx context.Context, cfg ConsumerPhaseConfig) error {
	logger := log.FromContext(ctx)
	p := cfg.BuildParams
	bootstrapServers := p.Kafka.BootstrapServers
	if bootstrapServers == "" {
		return fmt.Errorf("external Kafka enabled but bootstrapServers not provided")
	}
	servers := strings.Split(bootstrapServers, ",")
	for _, server := range servers {
		server = strings.TrimSpace(server)
		logger.Info("Validating external Kafka connectivity", "server", server)
		conn, err := net.DialTimeout("tcp", server, 5*time.Second)
		if err != nil {
			logger.Error(err, "Failed to connect to external Kafka", "server", server)
			if cfg.ApplyStatus != nil {
				_ = cfg.ApplyStatus(ctx, metav1.Condition{
					Type:               "KafkaReady",
					Status:             metav1.ConditionFalse,
					Reason:             "ValidationFailed",
					Message:            fmt.Sprintf("Cannot connect to external Kafka server %s: %v", server, err),
					LastTransitionTime: metav1.Now(),
				})
			}
			return fmt.Errorf("external Kafka validation failed for %s: %w", server, err)
		}
		conn.Close()
		logger.Info("External Kafka server is reachable", "server", server)
	}
	return nil
}
