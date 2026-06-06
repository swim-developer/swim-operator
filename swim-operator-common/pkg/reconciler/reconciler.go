package reconciler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ReconcileSecret(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, desired *corev1.Secret) error {
	if err := ctrl.SetControllerReference(owner, desired, scheme); err != nil {
		return err
	}
	current := &corev1.Secret{}
	err := c.Get(ctx, client.ObjectKeyFromObject(desired), current)
	if errors.IsNotFound(err) {
		return c.Create(ctx, desired)
	} else if err != nil {
		return err
	}
	if !equality.Semantic.DeepEqual(current.StringData, desired.StringData) || !equality.Semantic.DeepEqual(current.Labels, desired.Labels) || len(current.OwnerReferences) == 0 {
		desired.SetResourceVersion(current.GetResourceVersion())
		desired.SetUID(current.GetUID())
		return c.Update(ctx, desired)
	}
	return nil
}

func ReconcileConfigMap(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, desired *corev1.ConfigMap) error {
	if err := ctrl.SetControllerReference(owner, desired, scheme); err != nil {
		return err
	}
	current := &corev1.ConfigMap{}
	err := c.Get(ctx, client.ObjectKeyFromObject(desired), current)
	if errors.IsNotFound(err) {
		return c.Create(ctx, desired)
	} else if err != nil {
		return err
	}
	if !equality.Semantic.DeepEqual(current.Data, desired.Data) || !equality.Semantic.DeepEqual(current.Labels, desired.Labels) || len(current.OwnerReferences) == 0 {
		desired.SetResourceVersion(current.GetResourceVersion())
		desired.SetUID(current.GetUID())
		return c.Update(ctx, desired)
	}
	return nil
}

func ReconcileService(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, desired *corev1.Service) error {
	if err := ctrl.SetControllerReference(owner, desired, scheme); err != nil {
		return err
	}
	current := &corev1.Service{}
	err := c.Get(ctx, client.ObjectKeyFromObject(desired), current)
	if errors.IsNotFound(err) {
		return c.Create(ctx, desired)
	} else if err != nil {
		return err
	}
	if !equality.Semantic.DeepEqual(current.Labels, desired.Labels) || len(current.OwnerReferences) == 0 {
		desired.Spec.ClusterIP = current.Spec.ClusterIP
		desired.SetResourceVersion(current.GetResourceVersion())
		desired.SetUID(current.GetUID())
		return c.Update(ctx, desired)
	}
	return nil
}

func ReconcileDeployment(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, desired *appsv1.Deployment) error {
	if err := ctrl.SetControllerReference(owner, desired, scheme); err != nil {
		return err
	}
	current := &appsv1.Deployment{}
	err := c.Get(ctx, client.ObjectKeyFromObject(desired), current)
	if errors.IsNotFound(err) {
		return c.Create(ctx, desired)
	} else if err != nil {
		return err
	}
	needsUpdate := *current.Spec.Replicas != *desired.Spec.Replicas ||
		current.Spec.Template.Spec.Containers[0].Image != desired.Spec.Template.Spec.Containers[0].Image ||
		!equality.Semantic.DeepEqual(current.Labels, desired.Labels) ||
		len(current.OwnerReferences) == 0

	if needsUpdate {
		patch := client.MergeFrom(current.DeepCopy())
		current.Spec.Replicas = desired.Spec.Replicas
		current.Spec.Template = desired.Spec.Template
		current.Labels = desired.Labels
		if len(current.OwnerReferences) == 0 {
			current.OwnerReferences = desired.OwnerReferences
		}
		return c.Patch(ctx, current, patch)
	}
	return nil
}

func ReconcileStatefulSet(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, desired *appsv1.StatefulSet) error {
	if err := ctrl.SetControllerReference(owner, desired, scheme); err != nil {
		return err
	}
	current := &appsv1.StatefulSet{}
	err := c.Get(ctx, client.ObjectKeyFromObject(desired), current)
	if errors.IsNotFound(err) {
		return c.Create(ctx, desired)
	} else if err != nil {
		return err
	}

	needsUpdate := *current.Spec.Replicas != *desired.Spec.Replicas ||
		current.Spec.Template.Spec.Containers[0].Image != desired.Spec.Template.Spec.Containers[0].Image ||
		len(current.OwnerReferences) == 0

	if needsUpdate {
		patch := client.MergeFrom(current.DeepCopy())
		current.Spec.Replicas = desired.Spec.Replicas
		current.Spec.Template = desired.Spec.Template
		if len(current.OwnerReferences) == 0 {
			current.OwnerReferences = desired.OwnerReferences
		}
		return c.Patch(ctx, current, patch)
	}
	return nil
}

func ReconcileIngress(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, desired *networkingv1.Ingress) error {
	if err := ctrl.SetControllerReference(owner, desired, scheme); err != nil {
		return err
	}
	current := &networkingv1.Ingress{}
	err := c.Get(ctx, client.ObjectKeyFromObject(desired), current)
	if errors.IsNotFound(err) {
		return c.Create(ctx, desired)
	} else if err != nil {
		return err
	}
	if !equality.Semantic.DeepEqual(current.Spec, desired.Spec) || !equality.Semantic.DeepEqual(current.Labels, desired.Labels) || len(current.OwnerReferences) == 0 {
		desired.SetResourceVersion(current.GetResourceVersion())
		desired.SetUID(current.GetUID())
		return c.Update(ctx, desired)
	}
	return nil
}

func ReconcileCertificate(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, desired *certmanagerv1.Certificate) (ctrl.Result, error) {
	if err := ctrl.SetControllerReference(owner, desired, scheme); err != nil {
		return ctrl.Result{}, err
	}
	current := &certmanagerv1.Certificate{}
	err := c.Get(ctx, client.ObjectKeyFromObject(desired), current)
	if errors.IsNotFound(err) {
		if err := c.Create(ctx, desired); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}
	if !equality.Semantic.DeepEqual(current.Labels, desired.Labels) || len(current.OwnerReferences) == 0 {
		desired.SetResourceVersion(current.GetResourceVersion())
		desired.SetUID(current.GetUID())
		if err := c.Update(ctx, desired); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func ReconcilePVC(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, desired *corev1.PersistentVolumeClaim) error {
	if err := ctrl.SetControllerReference(owner, desired, scheme); err != nil {
		return err
	}
	current := &corev1.PersistentVolumeClaim{}
	err := c.Get(ctx, client.ObjectKeyFromObject(desired), current)
	if errors.IsNotFound(err) {
		return c.Create(ctx, desired)
	}
	return client.IgnoreNotFound(err)
}

func ReconcileServiceAccount(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, desired *corev1.ServiceAccount) error {
	if err := ctrl.SetControllerReference(owner, desired, scheme); err != nil {
		return err
	}
	current := &corev1.ServiceAccount{}
	err := c.Get(ctx, client.ObjectKeyFromObject(desired), current)
	if errors.IsNotFound(err) {
		return c.Create(ctx, desired)
	}
	return client.IgnoreNotFound(err)
}

func ReconcileRole(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, desired *rbacv1.Role) error {
	if err := ctrl.SetControllerReference(owner, desired, scheme); err != nil {
		return err
	}
	current := &rbacv1.Role{}
	err := c.Get(ctx, client.ObjectKeyFromObject(desired), current)
	if errors.IsNotFound(err) {
		return c.Create(ctx, desired)
	} else if err != nil {
		return err
	}
	if !equality.Semantic.DeepEqual(current.Rules, desired.Rules) || len(current.OwnerReferences) == 0 {
		desired.SetResourceVersion(current.GetResourceVersion())
		desired.SetUID(current.GetUID())
		return c.Update(ctx, desired)
	}
	return nil
}

func ReconcileRoleBinding(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, desired *rbacv1.RoleBinding) error {
	if err := ctrl.SetControllerReference(owner, desired, scheme); err != nil {
		return err
	}
	current := &rbacv1.RoleBinding{}
	err := c.Get(ctx, client.ObjectKeyFromObject(desired), current)
	if errors.IsNotFound(err) {
		return c.Create(ctx, desired)
	} else if err != nil {
		return err
	}
	if !equality.Semantic.DeepEqual(current.Subjects, desired.Subjects) || !equality.Semantic.DeepEqual(current.RoleRef, desired.RoleRef) || len(current.OwnerReferences) == 0 {
		desired.SetResourceVersion(current.GetResourceVersion())
		desired.SetUID(current.GetUID())
		return c.Update(ctx, desired)
	}
	return nil
}

func ReconcileHPA(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, desired *autoscalingv2.HorizontalPodAutoscaler) error {
	if err := ctrl.SetControllerReference(owner, desired, scheme); err != nil {
		return err
	}
	current := &autoscalingv2.HorizontalPodAutoscaler{}
	err := c.Get(ctx, client.ObjectKeyFromObject(desired), current)
	if errors.IsNotFound(err) {
		return c.Create(ctx, desired)
	} else if err != nil {
		return err
	}
	if !equality.Semantic.DeepEqual(current.Spec, desired.Spec) || len(current.OwnerReferences) == 0 {
		desired.SetResourceVersion(current.GetResourceVersion())
		desired.SetUID(current.GetUID())
		return c.Update(ctx, desired)
	}
	return nil
}

func ReconcileUnstructured(ctx context.Context, c client.Client, owner client.Object, desired *unstructured.Unstructured) error {
	desired.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         owner.GetObjectKind().GroupVersionKind().GroupVersion().String(),
			Kind:               owner.GetObjectKind().GroupVersionKind().Kind,
			Name:               owner.GetName(),
			UID:                owner.GetUID(),
			Controller:         func() *bool { b := true; return &b }(),
			BlockOwnerDeletion: func() *bool { b := true; return &b }(),
		},
	})
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

func ReconcileServiceMonitor(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, desired *monitoringv1.ServiceMonitor) error {
	if err := ctrl.SetControllerReference(owner, desired, scheme); err != nil {
		return err
	}
	current := &monitoringv1.ServiceMonitor{}
	err := c.Get(ctx, client.ObjectKeyFromObject(desired), current)
	if errors.IsNotFound(err) {
		return c.Create(ctx, desired)
	} else if err != nil {
		return err
	}
	if !equality.Semantic.DeepEqual(current.Spec, desired.Spec) || !equality.Semantic.DeepEqual(current.Labels, desired.Labels) || len(current.OwnerReferences) == 0 {
		desired.SetResourceVersion(current.GetResourceVersion())
		desired.SetUID(current.GetUID())
		return c.Update(ctx, desired)
	}
	return nil
}

func ComputeConfigHash(obj client.Object) string {
	data, _ := json.Marshal(obj)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func ReconcileArtemisSSLSecret(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner client.Object, certSecretName, targetSecretName, keystorePassword string) error {
	return ReconcileArtemisSSLSecretFromPEM(ctx, c, scheme, owner, ArtemisSSLSecretFromPEMInput{
		CertSecretName:   certSecretName,
		TargetSecretName: targetSecretName,
		KeystorePassword: keystorePassword,
	})
}
