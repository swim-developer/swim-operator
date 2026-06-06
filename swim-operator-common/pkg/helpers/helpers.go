package helpers

import (
	"context"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IsPodReady(ctx context.Context, c client.Client, namespace string, labels map[string]string) bool {
	podList := &corev1.PodList{}
	if err := c.List(ctx, podList, client.InNamespace(namespace), client.MatchingLabels(labels)); err != nil {
		return false
	}
	for _, pod := range podList.Items {
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
				return true
			}
		}
	}
	return false
}

func IsKafkaClusterReady(ctx context.Context, c client.Client, namespace, name string) bool {
	kafka := &unstructured.Unstructured{}
	kafka.SetAPIVersion(constants.KafkaAPIVersion)
	kafka.SetKind("Kafka")
	if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, kafka); err != nil {
		return false
	}
	status, ok := kafka.Object["status"].(map[string]interface{})
	if !ok {
		return false
	}
	conditions, ok := status["conditions"].([]interface{})
	if !ok {
		return false
	}
	for _, cond := range conditions {
		condition, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}
		if condition["type"] == "Ready" && condition["status"] == "True" {
			return true
		}
	}
	return false
}
