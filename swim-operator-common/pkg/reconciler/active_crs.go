package reconciler

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var SwimCRKinds = []string{
	"SwimDigitalNotamProvider",
	"SwimDigitalNotamConsumer",
	"SwimDnotamConsumerValidator",
	"SwimDnotamProviderValidator",
	"SwimEd254Provider",
	"SwimEd254Consumer",
	"SwimEd254ConsumerValidator",
}

type SwimCRKindActivityQuery struct {
	Namespace   string
	Group       string
	Version     string
	Kind        string
	ExcludeKind string
	ExcludeName string
}

type SwimNamespaceSweepQuery struct {
	Namespace   string
	Group       string
	Version     string
	ExcludeKind string
	ExcludeName string
}

func HasActiveCRs(ctx context.Context, c client.Client, q SwimCRKindActivityQuery) bool {
	logger := log.FromContext(ctx)
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{Group: q.Group, Version: q.Version, Kind: q.Kind + "List"})

	if err := c.List(ctx, list, client.InNamespace(q.Namespace)); err != nil {
		return false
	}

	for _, item := range list.Items {
		if item.GetKind() == q.ExcludeKind && item.GetName() == q.ExcludeName {
			continue
		}
		if item.GetDeletionTimestamp() == nil {
			logger.V(1).Info("Found active SWIM CR", "kind", q.Kind, "name", item.GetName())
			return true
		}
	}
	return false
}

func IsLastSwimCRInNamespace(ctx context.Context, c client.Client, base SwimNamespaceSweepQuery) bool {
	logger := log.FromContext(ctx)

	for _, kind := range SwimCRKinds {
		if HasActiveCRs(ctx, c, SwimCRKindActivityQuery{
			Namespace:   base.Namespace,
			Group:       base.Group,
			Version:     base.Version,
			Kind:        kind,
			ExcludeKind: base.ExcludeKind,
			ExcludeName: base.ExcludeName,
		}) {
			return false
		}
	}

	logger.Info("This is the last SWIM CR in namespace, shared infrastructure will be cleaned up")
	return true
}
