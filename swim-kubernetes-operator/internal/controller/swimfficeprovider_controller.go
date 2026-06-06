package controller

import (
	"context"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1alpha1 "github.com/swim-developer/swim-kubernetes-operator/api/v1alpha1"
	"github.com/swim-developer/swim-operator-common/pkg/constants"
	swimlabels "github.com/swim-developer/swim-operator-common/pkg/labels"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const fficeProviderFinalizerName = constants.FficeProviderFinalizerName

type SwimFficeProviderReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=apps.swim-developer.github.io,resources=swimfficeproviders,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.swim-developer.github.io,resources=swimfficeproviders/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.swim-developer.github.io,resources=swimfficeproviders/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments;statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods;services;configmaps;persistentvolumeclaims;secrets;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=broker.amq.io,resources=activemqartemises,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kafka.strimzi.io,resources=kafkas;kafkanodepools;kafkatopics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete

func (r *SwimFficeProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cr := &appsv1alpha1.SwimFficeProvider{}
	if err := r.Get(ctx, req.NamespacedName, cr); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	result, err := r.handleFficeProviderFinalization(ctx, req, cr)
	if res, halt, e := shouldHaltAfterProviderStep(result, err); halt {
		return res, e
	}
	if cr.DeletionTimestamp != nil {
		return result, nil
	}

	result, err = r.ensureFficeProviderFinalizer(ctx, cr)
	if res, halt, e := shouldHaltAfterProviderStep(result, err); halt {
		return res, e
	}

	result, err = r.validateFficeProviderExternalKafka(ctx, cr)
	if res, halt, e := shouldHaltAfterProviderStep(result, err); halt {
		return res, e
	}

	if err := swimlabels.EnsureCRLabels(ctx, r.Client, cr, "provider", sharedManagedByValue); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileFficeProviderRBAC(ctx, cr); err != nil {
		return ctrl.Result{}, err
	}

	result, err = r.reconcileFficeProviderPostgres(ctx, cr)
	if res, halt, e := shouldHaltAfterProviderStep(result, err); halt {
		return res, e
	}

	clusterDomain := getOrDetectClusterDomain(ctx, r.Client, cr.Spec.Global.ClusterDomain, cr.Namespace)
	artemisIngressHost := r.buildFficeProviderArtemisHost(cr, clusterDomain)

	result, err = r.reconcileFficeProviderArtemis(ctx, cr, artemisIngressHost)
	if res, halt, e := shouldHaltAfterProviderStep(result, err); halt {
		return res, e
	}

	result, err = r.reconcileFficeProviderKafka(ctx, cr)
	if res, halt, e := shouldHaltAfterProviderStep(result, err); halt {
		return res, e
	}

	configHash, cfgRes, err := r.reconcileFficeProviderAppConfig(ctx, cr, clusterDomain)
	if res, halt, e := shouldHaltAfterProviderStepEmptyOnError(cfgRes, err); halt {
		return res, e
	}

	if err := r.reconcileFficeProviderAppDeployment(ctx, cr, clusterDomain, configHash); err != nil {
		return ctrl.Result{}, err
	}

	result, err = r.checkFficeProviderAppReadiness(ctx, cr)
	if res, halt, e := shouldHaltAfterProviderStep(result, err); halt {
		return res, e
	}

	r.updateFficeProviderStatus(ctx, cr, "Available", metav1.ConditionTrue, "Reconciled", "All resources created successfully")
	return ctrl.Result{}, nil
}

func (r *SwimFficeProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	b := ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.SwimFficeProvider{}).
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
		Owns(&networkingv1.Ingress{}).
		Owns(&certmanagerv1.Certificate{})
	if serviceMonitorAvailable(mgr) {
		b = b.Owns(&monitoringv1.ServiceMonitor{})
	}
	return b.Complete(r)
}
