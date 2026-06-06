package controller

import (
	"context"

	appsv1alpha1 "github.com/swim-developer/swim-openshift-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/swim-developer/swim-operator-common/pkg/constants"
)

const fficeConsumerValidatorFinalizerName = constants.FficeConsumerValidatorFinalizerName

type SwimFficeConsumerValidatorReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups=apps.swim-developer.github.io,resources=swimfficeconsumervalidators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.swim-developer.github.io,resources=swimfficeconsumervalidators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.swim-developer.github.io,resources=swimfficeconsumervalidators/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments;statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods;services;configmaps;secrets;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes/custom-host,verbs=create;update;patch
//+kubebuilder:rbac:groups=broker.amq.io,resources=activemqartemises,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.openshift.io,resources=ingresscontrollers,verbs=get;list
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete

func (r *SwimFficeConsumerValidatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cr := &appsv1alpha1.SwimFficeConsumerValidator{}
	err := r.Get(ctx, req.NamespacedName, cr)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return r.reconcileSwimFficeCV(ctx, req, cr)
}

func (r *SwimFficeConsumerValidatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.SwimFficeConsumerValidator{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Secret{}).
		Owns(&certmanagerv1.Certificate{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&routev1.Route{}).
		Owns(&autoscalingv2.HorizontalPodAutoscaler{}).
		Complete(r)
}
