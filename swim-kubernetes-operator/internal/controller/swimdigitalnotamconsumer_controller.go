package controller

import (
	"context"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	appsv1alpha1 "github.com/swim-developer/swim-kubernetes-operator/api/v1alpha1"
	swimlabels "github.com/swim-developer/swim-operator-common/pkg/labels"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SwimDigitalNotamConsumerReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=apps.swim-developer.github.io,resources=swimdigitalnotamconsumers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.swim-developer.github.io,resources=swimdigitalnotamconsumers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.swim-developer.github.io,resources=swimdigitalnotamconsumers/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments;statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods;services;configmaps;secrets;persistentvolumeclaims;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=kafka.strimzi.io,resources=kafkas;kafkanodepools;kafkatopics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=console.streamshub.github.com,resources=consoles,verbs=get;list;watch;create;update;patch;delete

func (r *SwimDigitalNotamConsumerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cr := &appsv1alpha1.SwimDigitalNotamConsumer{}
	if err := r.Get(ctx, req.NamespacedName, cr); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if result, err := r.handleConsumerFinalization(ctx, req, cr); err != nil || result.Requeue || result.RequeueAfter > 0 || !cr.ObjectMeta.DeletionTimestamp.IsZero() {
		return result, err
	}

	if result, err := r.ensureConsumerFinalizer(ctx, cr, req); err != nil || result.Requeue || result.RequeueAfter > 0 {
		return result, err
	}

	if err := swimlabels.EnsureCRLabels(ctx, r.Client, cr, "consumer", sharedManagedByValue); err != nil {
		return ctrl.Result{}, err
	}

	if result, err := r.reconcileConsumerKafka(ctx, req, cr); err != nil || result.Requeue || result.RequeueAfter > 0 {
		return result, err
	}

	if err := r.reconcileConsumerRBAC(ctx, req, cr); err != nil {
		return ctrl.Result{}, err
	}

	bundle := r.reconcileConsumerSecrets(ctx, req, cr)

	if result, err := r.reconcileConsumerMongoDB(ctx, req, cr); err != nil || result.Requeue || result.RequeueAfter > 0 {
		return result, err
	}

	if result, err := r.reconcileConsumerClient(ctx, req, cr, bundle); err != nil || result.Requeue || result.RequeueAfter > 0 {
		return result, err
	}

	_ = r.updateStatus(ctx, req, cr, "Available", metav1.ConditionTrue, "Reconciled", "All resources created and ready")

	return ctrl.Result{}, nil
}

func (r *SwimDigitalNotamConsumerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.SwimDigitalNotamConsumer{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&appsv1.Deployment{}).
		Owns(&autoscalingv2.HorizontalPodAutoscaler{}).
		Owns(&certmanagerv1.Certificate{}).
		Complete(r)
}
