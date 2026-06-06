package controller

import (
	"context"
	"reflect"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/swim-developer/swim-operator-common/pkg/constants"
	commonreconciler "github.com/swim-developer/swim-operator-common/pkg/reconciler"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const sharedManagedByLabel = constants.SharedManagedByLabel
const sharedManagedByValue = "swim-operator"

func reconcileSharedUnstructured(ctx context.Context, c client.Client, desired *unstructured.Unstructured) error {
	return commonreconciler.ReconcileSharedUnstructured(ctx, c, desired, sharedManagedByLabel, sharedManagedByValue)
}

func cleanupSharedArtemis(ctx context.Context, c client.Client, namespace, artemisName string) {
	commonreconciler.CleanupSharedArtemis(ctx, c, namespace, artemisName)
}

func deleteArtemisBroker(ctx context.Context, c client.Client, namespace, artemisName string) {
	commonreconciler.DeleteArtemisBroker(ctx, c, namespace, artemisName)
}

func deleteArtemisSecrets(ctx context.Context, c client.Client, namespace, artemisName string) {
	commonreconciler.DeleteArtemisSecrets(ctx, c, namespace, artemisName)
}

func deleteArtemisCertificate(ctx context.Context, c client.Client, namespace, artemisName string) {
	commonreconciler.DeleteArtemisCertificate(ctx, c, namespace, artemisName)
}

func deleteArtemisJMXService(ctx context.Context, c client.Client, namespace, artemisName string) {
	commonreconciler.DeleteArtemisJMXService(ctx, c, namespace, artemisName)
}

func deleteArtemisPVCs(ctx context.Context, c client.Client, namespace, artemisName string) {
	commonreconciler.DeleteArtemisPVCs(ctx, c, namespace, artemisName)
}

func cleanupSharedKafka(ctx context.Context, c client.Client, namespace string) {
	commonreconciler.CleanupSharedKafka(ctx, c, namespace)
}

func cleanupSharedInfraIfLast(ctx context.Context, c client.Client, namespace, excludeKind, excludeName, artemisName string, kafkaEnabled bool) {
	commonreconciler.CleanupSharedInfraIfLast(ctx, c, namespace, excludeKind, excludeName, artemisName, kafkaEnabled)
}

func cleanupServiceBrokerProperties(ctx context.Context, c client.Client, namespace, artemisName, servicePrefix string) {
	commonreconciler.CleanupServiceBrokerProperties(ctx, c, namespace, artemisName, servicePrefix)
}

func reconcileRouteResource(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, desired *routev1.Route) error {
	if err := ctrl.SetControllerReference(owner, desired, scheme); err != nil {
		return err
	}
	current := &routev1.Route{}
	err := c.Get(ctx, client.ObjectKeyFromObject(desired), current)
	if errors.IsNotFound(err) {
		return c.Create(ctx, desired)
	} else if err != nil {
		return err
	}
	if !reflect.DeepEqual(current.Spec, desired.Spec) || !reflect.DeepEqual(current.Labels, desired.Labels) || len(current.OwnerReferences) == 0 {
		desired.SetResourceVersion(current.GetResourceVersion())
		desired.SetUID(current.GetUID())
		return c.Update(ctx, desired)
	}
	return nil
}
