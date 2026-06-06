package consumer

import (
	"context"
	"fmt"
	"strings"

	"github.com/swim-developer/swim-operator-common/pkg/domain"
	"github.com/swim-developer/swim-operator-common/pkg/labels"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// DnotamConsumerKafkaTopics returns the canonical list of Kafka topics for the DNOTAM consumer.
// Identical on both OpenShift and vanilla Kubernetes.
func DnotamConsumerKafkaTopics() []string {
	return []string{
		"dnotam-events-airspace-topic",
		"dnotam-events-closure-topic",
		"dnotam-events-hazards-navaids-topic",
		"dnotam-events-others-topic",
		"dnotam-events-restriction-topic",
		"dnotam-events-surface-condition-topic",
		"dnotam-events-all-topic",
		"dnotam-events-dlq-topic",
	}
}

// FficeConsumerKafkaTopics returns the canonical list of Kafka topics for the FF-ICE consumer.
func FficeConsumerKafkaTopics() []string {
	return []string{
		"ffice-events-topic",
		"ffice-flight-plan-topic",
		"ffice-flight-update-topic",
		"ffice-operations-topic",
		"ffice-trial-topic",
		"ffice-submission-topic",
		"ffice-data-topic",
		"ffice-dlq-topic",
		"ffice-inbox-topic",
	}
}

func ConsumerKafkaTopicPartitions(p ConsumerBuildParams, topicName string) int64 {
	if p.Flavor == ConsumerFlavorDnotam && topicName == "dnotam-events-all-topic" {
		return 10
	}
	if p.Flavor == ConsumerFlavorEd254 && (topicName == "ed254-arrival-sequence-topic" || topicName == "ed254-events-all-topic") {
		return 10
	}
	if p.Flavor == ConsumerFlavorFfice && topicName == "ffice-events-topic" {
		return 10
	}
	return 3
}

func BuildConsumerKafkaConsole(ctx context.Context, p ConsumerBuildParams, managedBy string, resolve func(context.Context, string, string) string) *unstructured.Unstructured {
	clusterDomain := resolve(ctx, p.GlobalClusterDomain, p.Namespace)
	hostname := fmt.Sprintf("kafka-ui-%s.%s", p.Namespace, domain.GetAppsDomain(strings.TrimPrefix(clusterDomain, "apps.")))
	return resources.BuildKafkaConsole(p.Namespace, hostname, labels.StandardLabels("kafka-ui", "kafka", p.Name, managedBy))
}

func BuildConsumerKafkaNodePool(p ConsumerBuildParams, managedBy string) *unstructured.Unstructured {
	return resources.BuildKafkaNodePool(resources.KafkaClusterParams{
		Namespace:   p.Namespace,
		Labels:      labels.StandardLabels("kafka", "kafka", p.Name, managedBy),
		Replicas:    p.Kafka.Replicas,
		StorageSize: p.Kafka.StorageSize,
	})
}

func BuildConsumerKafkaCluster(p ConsumerBuildParams, managedBy string) *unstructured.Unstructured {
	return resources.BuildKafkaCluster(resources.KafkaClusterParams{
		Namespace:   p.Namespace,
		Labels:      labels.StandardLabels("kafka", "kafka", p.Name, managedBy),
		Version:     p.Kafka.Version,
		Replicas:    p.Kafka.Replicas,
		StorageSize: p.Kafka.StorageSize,
	})
}

func BuildConsumerKafkaTopic(p ConsumerBuildParams, managedBy string, topicName string) *unstructured.Unstructured {
	partitions := ConsumerKafkaTopicPartitions(p, topicName)
	return resources.BuildKafkaTopic(p.Namespace, topicName, labels.StandardLabels("kafka", "kafka", p.Name, managedBy), partitions)
}
