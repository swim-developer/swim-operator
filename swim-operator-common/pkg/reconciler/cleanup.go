package reconciler

import (
	"context"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func CleanupSharedArtemis(ctx context.Context, c client.Client, namespace, artemisName string) {
	DeleteArtemisBroker(ctx, c, namespace, artemisName)
	DeleteArtemisSecrets(ctx, c, namespace, artemisName)
	DeleteArtemisCertificate(ctx, c, namespace, artemisName)
	DeleteArtemisJMXService(ctx, c, namespace, artemisName)
	DeleteArtemisPVCs(ctx, c, namespace, artemisName)
}

func DeleteArtemisBroker(ctx context.Context, c client.Client, namespace, artemisName string) {
	logger := log.FromContext(ctx)
	artemis := &unstructured.Unstructured{}
	artemis.SetAPIVersion("broker.amq.io/v1beta1")
	artemis.SetKind("ActiveMQArtemis")
	artemis.SetName(artemisName)
	artemis.SetNamespace(namespace)
	if err := c.Delete(ctx, artemis); err != nil && !errors.IsNotFound(err) {
		logger.Error(err, "Failed to delete shared Artemis broker", "name", artemisName)
	} else {
		logger.Info("Deleted shared Artemis broker", "name", artemisName)
	}
}

func DeleteArtemisSecrets(ctx context.Context, c client.Client, namespace, artemisName string) {
	logger := log.FromContext(ctx)
	secretNames := []string{
		artemisName + "-credentials",
		artemisName + "-keystore-password",
		artemisName + "-sso-jaas-config",
		artemisName + "-amqp-tls",
		artemisName + "-ssl-secret",
		artemisName + "-console-ssl-secret",
	}
	for _, name := range secretNames {
		secret := &unstructured.Unstructured{}
		secret.SetAPIVersion("v1")
		secret.SetKind("Secret")
		secret.SetName(name)
		secret.SetNamespace(namespace)
		if err := c.Delete(ctx, secret); err != nil && !errors.IsNotFound(err) {
			logger.V(1).Info("Failed to delete Artemis secret", "name", name, "error", err.Error())
		}
	}
}

func DeleteArtemisCertificate(ctx context.Context, c client.Client, namespace, artemisName string) {
	logger := log.FromContext(ctx)
	certName := artemisName + "-amqp"
	cert := &unstructured.Unstructured{}
	cert.SetAPIVersion("cert-manager.io/v1")
	cert.SetKind("Certificate")
	cert.SetName(certName)
	cert.SetNamespace(namespace)
	if err := c.Delete(ctx, cert); err != nil && !errors.IsNotFound(err) {
		logger.V(1).Info("Failed to delete Artemis certificate", "name", certName)
	}
}

func DeleteArtemisJMXService(ctx context.Context, c client.Client, namespace, artemisName string) {
	logger := log.FromContext(ctx)
	jmxSvcName := artemisName + "-jmx-svc"
	jmxSvc := &unstructured.Unstructured{}
	jmxSvc.SetAPIVersion("v1")
	jmxSvc.SetKind("Service")
	jmxSvc.SetName(jmxSvcName)
	jmxSvc.SetNamespace(namespace)
	if err := c.Delete(ctx, jmxSvc); err != nil && !errors.IsNotFound(err) {
		logger.V(1).Info("Failed to delete Artemis JMX service", "name", jmxSvcName)
	}
}

func DeleteArtemisPVCs(ctx context.Context, c client.Client, namespace, artemisName string) {
	logger := log.FromContext(ctx)
	pvcList := &corev1.PersistentVolumeClaimList{}
	if err := c.List(ctx, pvcList, client.InNamespace(namespace), client.MatchingLabels{"ActiveMQArtemis": artemisName}); err == nil {
		for i := range pvcList.Items {
			if err := c.Delete(ctx, &pvcList.Items[i]); err != nil && !errors.IsNotFound(err) {
				logger.V(1).Info("Failed to delete Artemis PVC", "pvc", pvcList.Items[i].Name)
			}
		}
	}
}

func CleanupSharedInfraIfLast(ctx context.Context, c client.Client, namespace, excludeKind, excludeName, artemisName string, kafkaEnabled bool) {
	if !IsLastSwimCRInNamespace(ctx, c, SwimNamespaceSweepQuery{
		Namespace:   namespace,
		Group:       "apps.swim-developer.github.io",
		Version:     "v1alpha1",
		ExcludeKind: excludeKind,
		ExcludeName: excludeName,
	}) {
		return
	}
	CleanupSharedArtemis(ctx, c, namespace, artemisName)
	if kafkaEnabled {
		CleanupSharedKafka(ctx, c, namespace)
	}
}

func CleanupSharedKafka(ctx context.Context, c client.Client, namespace string) {
	logger := log.FromContext(ctx)
	kafkaResources := []struct {
		apiVersion string
		kind       string
		name       string
	}{
		{"console.streamshub.github.com/v1alpha1", "Console", "kafka-ui"},
		{constants.KafkaAPIVersion, "Kafka", "kafka"},
		{constants.KafkaAPIVersion, "KafkaNodePool", "pool-a"},
	}
	for _, res := range kafkaResources {
		obj := &unstructured.Unstructured{}
		obj.SetAPIVersion(res.apiVersion)
		obj.SetKind(res.kind)
		obj.SetName(res.name)
		obj.SetNamespace(namespace)
		if err := c.Delete(ctx, obj); err != nil && !errors.IsNotFound(err) {
			logger.Error(err, "Failed to delete shared Kafka resource", "kind", res.kind, "name", res.name)
		} else if !errors.IsNotFound(err) {
			logger.Info("Deleted shared Kafka resource", "kind", res.kind, "name", res.name)
		}
	}
}

func CleanupServiceBrokerProperties(ctx context.Context, c client.Client, namespace, artemisName, servicePrefix string) {
	logger := log.FromContext(ctx)
	secretNames := []string{
		artemisName + "-" + servicePrefix + "-address-bp",
		artemisName + "-" + servicePrefix + "-security-bp",
	}
	for _, name := range secretNames {
		secret := &unstructured.Unstructured{}
		secret.SetAPIVersion("v1")
		secret.SetKind("Secret")
		secret.SetName(name)
		secret.SetNamespace(namespace)
		if err := c.Delete(ctx, secret); err != nil && !errors.IsNotFound(err) {
			logger.V(1).Info("Failed to delete broker properties secret", "name", name)
		}
	}
}

func ReconcileSharedUnstructured(ctx context.Context, c client.Client, desired *unstructured.Unstructured, managedByLabel, managedByValue string) error {
	if desired.GetLabels() == nil {
		desired.SetLabels(map[string]string{})
	}
	labels := desired.GetLabels()
	labels[managedByLabel] = managedByValue
	desired.SetLabels(labels)

	current := &unstructured.Unstructured{}
	current.SetGroupVersionKind(desired.GroupVersionKind())
	err := c.Get(ctx, client.ObjectKey{Namespace: desired.GetNamespace(), Name: desired.GetName()}, current)
	if errors.IsNotFound(err) {
		return c.Create(ctx, desired)
	} else if err != nil {
		return err
	}

	desired.SetResourceVersion(current.GetResourceVersion())
	desired.SetUID(current.GetUID())
	return c.Update(ctx, desired)
}
