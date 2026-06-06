package labels

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func StandardLabels(app, component, instance, managedBy string) map[string]string {
	return map[string]string{
		"app":                          app,
		"app.kubernetes.io/name":       app,
		"app.kubernetes.io/instance":   instance,
		"app.kubernetes.io/component":  component,
		"app.kubernetes.io/part-of":    managedBy,
		"app.kubernetes.io/managed-by": managedBy,
	}
}

func EnsureCRLabels(ctx context.Context, c client.Client, obj client.Object, component, managedBy string) error {
	objLabels := obj.GetLabels()
	if objLabels == nil {
		objLabels = make(map[string]string)
	}

	desired := StandardLabels(obj.GetName(), component, obj.GetName(), managedBy)

	needsUpdate := false
	for k, v := range desired {
		if objLabels[k] != v {
			objLabels[k] = v
			needsUpdate = true
		}
	}

	if !needsUpdate {
		return nil
	}

	obj.SetLabels(objLabels)
	return c.Update(ctx, obj)
}
