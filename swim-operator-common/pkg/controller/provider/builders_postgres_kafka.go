package provider

import (
	"context"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/labels"
	commonreconciler "github.com/swim-developer/swim-operator-common/pkg/reconciler"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func BuildProviderPostgresSecret(p ProviderBuildParams, managedBy string) *corev1.Secret {
	return resources.BuildPostgresSecret(p.Strategy.PostgresParams(p, managedBy))
}

func BuildProviderPostgresPVC(p ProviderBuildParams, managedBy string) *corev1.PersistentVolumeClaim {
	return resources.BuildPostgresPVC(p.Strategy.PostgresParams(p, managedBy))
}

func BuildProviderPostgresStatefulSet(p ProviderBuildParams, managedBy string) *appsv1.StatefulSet {
	pp := p.Strategy.PostgresParams(p, managedBy)
	if p.PostgresUpstream {
		return resources.BuildUpstreamPostgresStatefulSet(pp)
	}
	return resources.BuildPostgresStatefulSet(pp)
}

func BuildProviderPostgresService(p ProviderBuildParams, managedBy string) *corev1.Service {
	pp := p.Strategy.PostgresParams(p, managedBy)
	return resources.BuildPostgresService(pp.Name, pp.Namespace, pp.Labels)
}

func BuildProviderKafkaTopic(p ProviderBuildParams, managedBy string, topicName string) *unstructured.Unstructured {
	partitions := p.Strategy.KafkaTopicPartitions(topicName)
	return resources.BuildKafkaTopic(p.Namespace, topicName, labels.StandardLabels("kafka", "kafka", p.Name, managedBy), partitions)
}

func BuildProviderKafkaTopicAll(p ProviderBuildParams, managedBy string) *unstructured.Unstructured {
	return BuildProviderKafkaTopic(p, managedBy, p.Strategy.KafkaTopicAllName())
}

func BuildProviderKafkaTopicDLQ(p ProviderBuildParams, managedBy string) *unstructured.Unstructured {
	return BuildProviderKafkaTopic(p, managedBy, p.Strategy.KafkaTopicDLQName())
}

func BuildProviderKafkaNodePool(p ProviderBuildParams, managedBy string) *unstructured.Unstructured {
	k := p.Kafka
	return resources.BuildKafkaNodePool(resources.KafkaClusterParams{
		Namespace:   p.Namespace,
		Labels:      labels.StandardLabels("kafka", "kafka", p.Name, managedBy),
		Replicas:    k.Replicas,
		StorageSize: k.StorageSize,
	})
}

func BuildProviderKafkaCluster(p ProviderBuildParams, managedBy string) *unstructured.Unstructured {
	k := p.Kafka
	return resources.BuildKafkaCluster(resources.KafkaClusterParams{
		Namespace:   p.Namespace,
		Labels:      labels.StandardLabels("kafka", "kafka", p.Name, managedBy),
		Version:     k.Version,
		Replicas:    k.Replicas,
		StorageSize: k.StorageSize,
	})
}

func ReconcileProviderManagedKafka(ctx context.Context, c client.Client, p ProviderBuildParams, managedByLabel, managedByValue string, reconcileTopic func(*unstructured.Unstructured) error, resolve func(context.Context, string, string) string) error {
	kafka := &unstructured.Unstructured{}
	kafka.SetAPIVersion(constants.KafkaAPIVersion)
	kafka.SetKind("Kafka")
	err := c.Get(ctx, client.ObjectKey{Namespace: p.Namespace, Name: "kafka"}, kafka)
	if err == nil && p.Strategy.SkipIfKafkaExists() {
		return nil
	}
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if errors.IsNotFound(err) {
		if err := commonreconciler.ReconcileSharedUnstructured(ctx, c, BuildProviderKafkaCluster(p, managedByValue), managedByLabel, managedByValue); err != nil {
			return err
		}
		if err := commonreconciler.ReconcileSharedUnstructured(ctx, c, BuildProviderKafkaNodePool(p, managedByValue), managedByLabel, managedByValue); err != nil {
			return err
		}
		if p.Kafka.KafkaConsoleEnabled {
			if err := commonreconciler.ReconcileSharedUnstructured(ctx, c, BuildProviderKafkaConsole(ctx, p, managedByValue, resolve), managedByLabel, managedByValue); err != nil {
				return err
			}
		}
	}
	if err := reconcileTopic(BuildProviderKafkaTopicAll(p, managedByValue)); err != nil {
		return err
	}
	return reconcileTopic(BuildProviderKafkaTopicDLQ(p, managedByValue))
}
