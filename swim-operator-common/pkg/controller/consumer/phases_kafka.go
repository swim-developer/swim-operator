package consumer

import (
	"context"
	"time"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/helpers"
	commonreconciler "github.com/swim-developer/swim-operator-common/pkg/reconciler"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func ReconcileConsumerKafkaPhase(ctx context.Context, cfg ConsumerPhaseConfig) (ctrl.Result, error) {
	p := cfg.BuildParams
	if !p.Kafka.Enabled {
		return ctrl.Result{}, nil
	}
	deploymentMode := p.Kafka.DeploymentMode
	if deploymentMode == "" {
		deploymentMode = "managed"
	}
	if deploymentMode == "managed" {
		return reconcileManagedConsumerKafka(ctx, cfg)
	}
	if err := ValidateConsumerExternalKafka(ctx, cfg); err != nil {
		return ctrl.Result{}, err
	}
	applyConsumerCondition(ctx, cfg, "KafkaReady", metav1.ConditionTrue, "Validated", "External Kafka is reachable and validated")
	return ctrl.Result{}, nil
}

func reconcileManagedConsumerKafka(ctx context.Context, cfg ConsumerPhaseConfig) (ctrl.Result, error) {
	if err := ensureConsumerKafkaCluster(ctx, cfg); err != nil {
		return ctrl.Result{}, err
	}
	if result, err := waitForConsumerKafkaReady(ctx, cfg); err != nil || result.RequeueAfter > 0 {
		return result, err
	}
	if err := reconcileConsumerKafkaTopics(ctx, cfg); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func ensureConsumerKafkaCluster(ctx context.Context, cfg ConsumerPhaseConfig) error {
	p := cfg.BuildParams
	existingKafka := &unstructured.Unstructured{}
	existingKafka.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   constants.KafkaGroup,
		Version: "v1beta2",
		Kind:    "Kafka",
	})
	err := cfg.Client.Get(ctx, types.NamespacedName{Name: "kafka", Namespace: p.Namespace}, existingKafka)
	if errors.IsNotFound(err) {
		resolve := cfg.ResolveClusterDomain
		if resolve == nil {
			resolve = func(_ context.Context, spec, _ string) string { return spec }
		}
		if p.Kafka.KafkaConsoleEnabled {
			if err := commonreconciler.ReconcileSharedUnstructured(ctx, cfg.Client, BuildConsumerKafkaConsole(ctx, p, cfg.ManagedByValue, resolve), cfg.ManagedByLabel, cfg.ManagedByValue); err != nil {
				return err
			}
		}
		if err := commonreconciler.ReconcileSharedUnstructured(ctx, cfg.Client, BuildConsumerKafkaNodePool(p, cfg.ManagedByValue), cfg.ManagedByLabel, cfg.ManagedByValue); err != nil {
			return err
		}
		if err := commonreconciler.ReconcileSharedUnstructured(ctx, cfg.Client, BuildConsumerKafkaCluster(p, cfg.ManagedByValue), cfg.ManagedByLabel, cfg.ManagedByValue); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func waitForConsumerKafkaReady(ctx context.Context, cfg ConsumerPhaseConfig) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	p := cfg.BuildParams
	ready := helpers.IsKafkaClusterReady(ctx, cfg.Client, p.Namespace, "kafka")
	if !ready {
		logger.Info("Kafka cluster not ready yet, requeuing in 10 seconds")
		applyConsumerCondition(ctx, cfg, "KafkaReady", metav1.ConditionFalse, "Provisioning", "Waiting for Kafka cluster to be ready")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	applyConsumerCondition(ctx, cfg, "KafkaReady", metav1.ConditionTrue, "Ready", "Kafka cluster is ready and accepting connections")
	return ctrl.Result{}, nil
}

func reconcileConsumerKafkaTopics(ctx context.Context, cfg ConsumerPhaseConfig) error {
	for _, topicName := range cfg.KafkaTopics {
		if err := consumerReconcileUnstructured(ctx, cfg, BuildConsumerKafkaTopic(cfg.BuildParams, cfg.ManagedByValue, topicName)); err != nil {
			return err
		}
	}
	return nil
}
