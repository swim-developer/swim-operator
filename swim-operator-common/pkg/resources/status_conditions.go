package resources

import (
	"context"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func UpsertCondition(conditions []metav1.Condition, cond metav1.Condition) ([]metav1.Condition, bool) {
	if conditions == nil {
		return []metav1.Condition{cond}, true
	}
	for i, existing := range conditions {
		if existing.Type != cond.Type {
			continue
		}
		if existing.Status == cond.Status && existing.Reason == cond.Reason && existing.Message == cond.Message {
			return conditions, false
		}
		conditions[i] = cond
		return conditions, true
	}
	return append(conditions, cond), true
}

// MakeApplyStatusFunc returns a closure that fetches the latest CR, sets ObservedGeneration,
// upserts the given condition and persists the status subresource.
// T must be a pointer to a kubebuilder-generated CR struct that embeds Status.Conditions.
func MakeApplyStatusFunc[T client.Object](
	c client.Client,
	namespacedName types.NamespacedName,
	factory func() T,
	getConditions func(T) *[]metav1.Condition,
) func(context.Context, metav1.Condition) error {
	return func(ctx context.Context, condition metav1.Condition) error {
		latest := factory()
		if err := c.Get(ctx, namespacedName, latest); err != nil {
			return err
		}
		condition.ObservedGeneration = latest.GetGeneration()
		if condition.LastTransitionTime.IsZero() {
			condition.LastTransitionTime = metav1.Now()
		}
		conds, changed := UpsertCondition(*getConditions(latest), condition)
		if !changed {
			return nil
		}
		*getConditions(latest) = conds
		if err := c.Status().Update(ctx, latest); err != nil {
			if k8serrors.IsConflict(err) {
				return nil
			}
			return err
		}
		return nil
	}
}

// MakeRemoveFinalizerFunc returns a closure that fetches the latest CR and removes
// the given finalizer if present.
func MakeRemoveFinalizerFunc[T client.Object](
	c client.Client,
	namespacedName types.NamespacedName,
	factory func() T,
	finalizerName string,
) func(context.Context) error {
	return func(ctx context.Context) error {
		latest := factory()
		if err := c.Get(ctx, namespacedName, latest); err != nil {
			return client.IgnoreNotFound(err)
		}
		if controllerutil.ContainsFinalizer(latest, finalizerName) {
			controllerutil.RemoveFinalizer(latest, finalizerName)
			if err := c.Update(ctx, latest); err != nil {
				return client.IgnoreNotFound(err)
			}
		}
		return nil
	}
}
