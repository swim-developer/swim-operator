package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1alpha1 "github.com/swim-developer/swim-openshift-operator/api/v1alpha1"
	"github.com/swim-developer/swim-operator-common/pkg/constants"
)

const ed254ProviderFinalizerName = constants.Ed254ProviderFinalizerName

type SwimEd254ProviderReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=apps.swim-developer.github.io,resources=swimed254providers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.swim-developer.github.io,resources=swimed254providers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.swim-developer.github.io,resources=swimed254providers/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments;statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods;services;configmaps;persistentvolumeclaims;secrets;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes/custom-host,verbs=create;update;patch
//+kubebuilder:rbac:groups=operator.openshift.io,resources=ingresscontrollers,verbs=get;list
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=broker.amq.io,resources=activemqartemises,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kafka.strimzi.io,resources=kafkas;kafkanodepools;kafkatopics,verbs=get;list;watch;create;update;patch;delete

func (r *SwimEd254ProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cr := &appsv1alpha1.SwimEd254Provider{}
	if err := r.Get(ctx, req.NamespacedName, cr); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	result, err := r.handleEd254ProviderFinalization(ctx, req, cr)
	if err != nil {
		return result, err
	}
	if providerRequeueResult(result) {
		return result, nil
	}
	if cr.DeletionTimestamp != nil {
		return result, nil
	}

	result, err = r.ensureEd254ProviderFinalizer(ctx, cr)
	if err != nil {
		return result, err
	}
	if providerRequeueResult(result) {
		return result, nil
	}

	result, err = r.validateEd254ProviderExternalKafka(ctx, cr)
	if err != nil {
		return result, err
	}
	if providerRequeueResult(result) {
		return result, nil
	}

	if err := ensureCRLabels(ctx, r.Client, cr, "provider"); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileEd254ProviderRBAC(ctx, cr); err != nil {
		return ctrl.Result{}, err
	}

	result, err = r.reconcileEd254ProviderPostgres(ctx, cr)
	if err != nil {
		return result, err
	}
	if providerRequeueResult(result) {
		return result, nil
	}

	clusterDomain := getOrDetectClusterDomain(ctx, r.Client, cr.Spec.Global.ClusterDomain, cr.Namespace)
	artemisIngressHost := r.buildEd254ProviderArtemisHost(cr, clusterDomain)

	result, err = r.reconcileEd254ProviderArtemis(ctx, cr, artemisIngressHost)
	if err != nil {
		return result, err
	}
	if providerRequeueResult(result) {
		return result, nil
	}

	result, err = r.reconcileEd254ProviderKafka(ctx, cr)
	if err != nil {
		return result, err
	}
	if providerRequeueResult(result) {
		return result, nil
	}

	configHash, certRes, err := r.reconcileEd254ProviderAppConfig(ctx, cr, clusterDomain)
	if err != nil {
		return ctrl.Result{}, err
	}
	if providerRequeueResult(certRes) {
		return certRes, nil
	}

	if err := r.reconcileEd254ProviderAppDeployment(ctx, cr, clusterDomain, configHash); err != nil {
		return ctrl.Result{}, err
	}

	result, err = r.checkEd254ProviderAppReadiness(ctx, cr)
	if err != nil {
		return result, err
	}
	if providerRequeueResult(result) {
		return result, nil
	}

	r.updateStatus(ctx, cr, "Available", metav1.ConditionTrue, "Reconciled", "All resources created successfully")
	return ctrl.Result{}, nil
}

func (r *SwimEd254ProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.SwimEd254Provider{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&autoscalingv2.HorizontalPodAutoscaler{}).
		Owns(&routev1.Route{}).
		Owns(&certmanagerv1.Certificate{}).
		Complete(r)
}
