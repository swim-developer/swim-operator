package resources

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	KafkaGroup      = "kafka.strimzi.io"
	KafkaAPIVersion = "kafka.strimzi.io/v1beta2"
)

func MergeLabelsUnstructured(base, extra map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range extra {
		merged[k] = v
	}
	return merged
}

func LabelsToUnstructured(labels map[string]string) map[string]interface{} {
	m := make(map[string]interface{}, len(labels))
	for k, v := range labels {
		m[k] = v
	}
	return m
}

type KafkaClusterParams struct {
	Namespace   string
	Labels      map[string]string
	Version     string
	Replicas    int32
	StorageSize string
}

func BuildKafkaNodePool(p KafkaClusterParams) *unstructured.Unstructured {
	replicas := int64(p.Replicas)
	if replicas == 0 {
		replicas = 1
	}
	storageSize := StrDefault(p.StorageSize, "10Gi")

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": KafkaAPIVersion,
			"kind":       "KafkaNodePool",
			"metadata": map[string]interface{}{
				"name":      "pool-a",
				"namespace": p.Namespace,
				"labels": MergeLabelsUnstructured(LabelsToUnstructured(p.Labels), map[string]interface{}{
					"strimzi.io/cluster": "kafka",
				}),
			},
			"spec": map[string]interface{}{
				"replicas": replicas,
				"roles":    []interface{}{"controller", "broker"},
				"storage": map[string]interface{}{
					"type": "jbod",
					"volumes": []interface{}{
						map[string]interface{}{"id": 0, "type": "persistent-claim", "size": storageSize, "deleteClaim": true},
					},
				},
			},
		},
	}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: KafkaGroup, Version: "v1beta2", Kind: "KafkaNodePool"})
	return obj
}

func BuildKafkaCluster(p KafkaClusterParams) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": KafkaAPIVersion,
			"kind":       "Kafka",
			"metadata": map[string]interface{}{
				"name":      "kafka",
				"namespace": p.Namespace,
				"labels":    LabelsToUnstructured(p.Labels),
				"annotations": map[string]interface{}{
					"strimzi.io/node-pools": "enabled",
					"strimzi.io/kraft":      "enabled",
				},
			},
			"spec": map[string]interface{}{
				"kafka": map[string]interface{}{
					"version":         StrDefault(p.Version, "4.1.0"),
					"metadataVersion": "4.0-IV3",
					"listeners": []interface{}{
						map[string]interface{}{"name": "plain", "port": 9092, "type": "internal", "tls": false},
					},
					"config": map[string]interface{}{
						"offsets.topic.replication.factor":         1,
						"transaction.state.log.replication.factor": 1,
						"transaction.state.log.min.isr":            1,
						"default.replication.factor":               1,
						"min.insync.replicas":                      1,
						"log.retention.bytes":                      1073741824,
						"log.retention.hours":                      24,
					},
				},
				"entityOperator": map[string]interface{}{
					"topicOperator": map[string]interface{}{},
					"userOperator":  map[string]interface{}{},
				},
			},
		},
	}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: KafkaGroup, Version: "v1beta2", Kind: "Kafka"})
	return obj
}

func BuildKafkaTopic(namespace, topicName string, labels map[string]string, partitions int64) *unstructured.Unstructured {
	if partitions == 0 {
		partitions = 3
	}
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": KafkaAPIVersion,
			"kind":       "KafkaTopic",
			"metadata": map[string]interface{}{
				"name":      topicName,
				"namespace": namespace,
				"labels": MergeLabelsUnstructured(LabelsToUnstructured(labels), map[string]interface{}{
					"strimzi.io/cluster": "kafka",
				}),
			},
			"spec": map[string]interface{}{
				"partitions": partitions,
				"replicas":   1,
				"config": map[string]interface{}{
					"retention.ms":        604800000,
					"min.insync.replicas": 1,
					"segment.bytes":       1073741824,
				},
			},
		},
	}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: KafkaGroup, Version: "v1beta2", Kind: "KafkaTopic"})
	return obj
}

func BuildKafkaConsole(namespace, hostname string, labels map[string]string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "console.streamshub.github.com/v1alpha1",
			"kind":       "Console",
			"metadata": map[string]interface{}{
				"name":      "kafka-ui",
				"namespace": namespace,
				"labels":    LabelsToUnstructured(labels),
			},
			"spec": map[string]interface{}{
				"hostname": hostname,
				"kafkaClusters": []interface{}{
					map[string]interface{}{"name": "kafka", "namespace": namespace, "listener": "plain"},
				},
			},
		},
	}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "console.streamshub.github.com", Version: "v1alpha1", Kind: "Console"})
	return obj
}

func KafkaBootstrapInternal(namespace string) string {
	return "kafka-kafka-bootstrap." + namespace + ".svc.cluster.local:9092"
}
