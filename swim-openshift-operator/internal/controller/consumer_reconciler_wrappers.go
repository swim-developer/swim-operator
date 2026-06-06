package controller

import (
	"context"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1alpha1 "github.com/swim-developer/swim-openshift-operator/api/v1alpha1"
	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/controller/consumer"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *SwimDigitalNotamConsumerReconciler) consumerPhaseConfig(req ctrl.Request, cr *appsv1alpha1.SwimDigitalNotamConsumer) consumer.ConsumerPhaseConfig {
	return consumer.ConsumerPhaseConfig{
		Client:         r.Client,
		Scheme:         r.Scheme,
		Owner:          cr,
		Request:        req,
		FinalizerName:  constants.DnotamConsumerFinalizerName,
		CRKind:         "SwimDigitalNotamConsumer",
		BuildParams:    swimDigitalNotamConsumerBuildParams(cr),
		KafkaTopics:    consumer.DnotamConsumerKafkaTopics(),
		ManagedByLabel: sharedManagedByLabel,
		ManagedByValue: sharedManagedByValue,
		ResolveClusterDomain: func(ctx context.Context, specDomain, namespace string) string {
			return getOrDetectClusterDomain(ctx, r.Client, specDomain, namespace)
		},
		RemoveFinalizer: resources.MakeRemoveFinalizerFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimDigitalNotamConsumer { return &appsv1alpha1.SwimDigitalNotamConsumer{} },
			constants.DnotamConsumerFinalizerName,
		),
		ApplyStatus: resources.MakeApplyStatusFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimDigitalNotamConsumer { return &appsv1alpha1.SwimDigitalNotamConsumer{} },
			func(o *appsv1alpha1.SwimDigitalNotamConsumer) *[]metav1.Condition { return &o.Status.Conditions },
		),
	}
}

func (r *SwimDigitalNotamConsumerReconciler) handleConsumerFinalization(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimDigitalNotamConsumer) (ctrl.Result, error) {
	return consumer.HandleConsumerFinalization(ctx, r.consumerPhaseConfig(req, cr))
}

func (r *SwimDigitalNotamConsumerReconciler) ensureConsumerFinalizer(ctx context.Context, cr *appsv1alpha1.SwimDigitalNotamConsumer, req ctrl.Request) (ctrl.Result, error) {
	return consumer.EnsureConsumerFinalizer(ctx, r.consumerPhaseConfig(req, cr))
}

func (r *SwimDigitalNotamConsumerReconciler) reconcileConsumerKafka(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimDigitalNotamConsumer) (ctrl.Result, error) {
	return consumer.ReconcileConsumerKafkaPhase(ctx, r.consumerPhaseConfig(req, cr))
}

func (r *SwimDigitalNotamConsumerReconciler) reconcileConsumerRBAC(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimDigitalNotamConsumer) error {
	return consumer.ReconcileConsumerRBACPhase(ctx, r.consumerPhaseConfig(req, cr))
}

func (r *SwimDigitalNotamConsumerReconciler) reconcileConsumerSecrets(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimDigitalNotamConsumer) consumer.ConsumerSecretsBundle {
	return consumer.ReconcileConsumerSecretsBundle(ctx, r.consumerPhaseConfig(req, cr))
}

func (r *SwimDigitalNotamConsumerReconciler) reconcileConsumerMongoDB(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimDigitalNotamConsumer) (ctrl.Result, error) {
	return consumer.ReconcileConsumerMongoPhase(ctx, r.consumerPhaseConfig(req, cr))
}

func (r *SwimDigitalNotamConsumerReconciler) reconcileConsumerClient(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimDigitalNotamConsumer, bundle consumer.ConsumerSecretsBundle) (ctrl.Result, error) {
	return consumer.ReconcileConsumerClientPhase(ctx, r.consumerPhaseConfig(req, cr), bundle)
}

func (r *SwimDigitalNotamConsumerReconciler) updateStatus(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimDigitalNotamConsumer, conditionType string, status metav1.ConditionStatus, reason, message string) error {
	cfg := r.consumerPhaseConfig(req, cr)
	if cfg.ApplyStatus == nil {
		return nil
	}
	return cfg.ApplyStatus(ctx, metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})
}

func ocpEd254ConsumerKafkaTopics() []string {
	return []string{
		"ed254-arrival-sequence-topic",
		"ed254-provider-exception-topic",
		"ed254-dlq-topic",
	}
}

func (r *SwimEd254ConsumerReconciler) consumerPhaseConfig(req ctrl.Request, cr *appsv1alpha1.SwimEd254Consumer) consumer.ConsumerPhaseConfig {
	return consumer.ConsumerPhaseConfig{
		Client:         r.Client,
		Scheme:         r.Scheme,
		Owner:          cr,
		Request:        req,
		FinalizerName:  constants.Ed254ConsumerFinalizerName,
		CRKind:         "SwimEd254Consumer",
		BuildParams:    swimEd254ConsumerBuildParams(cr),
		KafkaTopics:    ocpEd254ConsumerKafkaTopics(),
		ManagedByLabel: sharedManagedByLabel,
		ManagedByValue: sharedManagedByValue,
		ResolveClusterDomain: func(ctx context.Context, specDomain, namespace string) string {
			return getOrDetectClusterDomain(ctx, r.Client, specDomain, namespace)
		},
		RemoveFinalizer: resources.MakeRemoveFinalizerFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimEd254Consumer { return &appsv1alpha1.SwimEd254Consumer{} },
			constants.Ed254ConsumerFinalizerName,
		),
		ApplyStatus: resources.MakeApplyStatusFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimEd254Consumer { return &appsv1alpha1.SwimEd254Consumer{} },
			func(o *appsv1alpha1.SwimEd254Consumer) *[]metav1.Condition { return &o.Status.Conditions },
		),
	}
}

func (r *SwimEd254ConsumerReconciler) handleEd254ConsumerFinalization(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimEd254Consumer) (ctrl.Result, error) {
	return consumer.HandleConsumerFinalization(ctx, r.consumerPhaseConfig(req, cr))
}

func (r *SwimEd254ConsumerReconciler) ensureEd254ConsumerFinalizer(ctx context.Context, cr *appsv1alpha1.SwimEd254Consumer, req ctrl.Request) (ctrl.Result, error) {
	return consumer.EnsureConsumerFinalizer(ctx, r.consumerPhaseConfig(req, cr))
}

func (r *SwimEd254ConsumerReconciler) reconcileEd254ConsumerKafka(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimEd254Consumer) (ctrl.Result, error) {
	return consumer.ReconcileConsumerKafkaPhase(ctx, r.consumerPhaseConfig(req, cr))
}

func (r *SwimEd254ConsumerReconciler) reconcileEd254ConsumerRBAC(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimEd254Consumer) error {
	return consumer.ReconcileConsumerRBACPhase(ctx, r.consumerPhaseConfig(req, cr))
}

func (r *SwimEd254ConsumerReconciler) reconcileEd254ConsumerSecrets(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimEd254Consumer) consumer.ConsumerSecretsBundle {
	return consumer.ReconcileConsumerSecretsBundle(ctx, r.consumerPhaseConfig(req, cr))
}

func (r *SwimEd254ConsumerReconciler) reconcileEd254ConsumerMongoDB(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimEd254Consumer) (ctrl.Result, error) {
	return consumer.ReconcileConsumerMongoPhase(ctx, r.consumerPhaseConfig(req, cr))
}

func (r *SwimEd254ConsumerReconciler) reconcileEd254ConsumerClient(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimEd254Consumer, bundle consumer.ConsumerSecretsBundle) (ctrl.Result, error) {
	return consumer.ReconcileConsumerClientPhase(ctx, r.consumerPhaseConfig(req, cr), bundle)
}

func (r *SwimEd254ConsumerReconciler) updateStatus(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimEd254Consumer, conditionType string, status metav1.ConditionStatus, reason, message string) error {
	cfg := r.consumerPhaseConfig(req, cr)
	if cfg.ApplyStatus == nil {
		return nil
	}
	return cfg.ApplyStatus(ctx, metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})
}

func (r *SwimDigitalNotamConsumerReconciler) ClientKeystorePasswordSecret(cr *appsv1alpha1.SwimDigitalNotamConsumer) *corev1.Secret {
	return consumer.BuildConsumerKeystorePasswordSecret(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue)
}

func (r *SwimDigitalNotamConsumerReconciler) ClientKafkaCredentialsSecret(cr *appsv1alpha1.SwimDigitalNotamConsumer) *corev1.Secret {
	return consumer.BuildConsumerKafkaCredentialsSecret(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue)
}

func (r *SwimDigitalNotamConsumerReconciler) MongoSecret(cr *appsv1alpha1.SwimDigitalNotamConsumer) *corev1.Secret {
	return consumer.BuildConsumerMongoSecret(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue)
}

func (r *SwimDigitalNotamConsumerReconciler) MongoPVC(cr *appsv1alpha1.SwimDigitalNotamConsumer) *corev1.PersistentVolumeClaim {
	return consumer.BuildConsumerMongoPVC(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue)
}

func (r *SwimDigitalNotamConsumerReconciler) MongoService(cr *appsv1alpha1.SwimDigitalNotamConsumer) *corev1.Service {
	return consumer.BuildConsumerMongoService(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue)
}

func (r *SwimDigitalNotamConsumerReconciler) MongoDeployment(cr *appsv1alpha1.SwimDigitalNotamConsumer, configHash string) *appsv1.Deployment {
	return consumer.BuildConsumerMongoDeployment(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue, configHash)
}

func (r *SwimDigitalNotamConsumerReconciler) ClientProvidersSecret(cr *appsv1alpha1.SwimDigitalNotamConsumer) *corev1.Secret {
	return consumer.BuildConsumerProvidersSecret(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue)
}

func (r *SwimDigitalNotamConsumerReconciler) ClientConfigMap(cr *appsv1alpha1.SwimDigitalNotamConsumer) *corev1.ConfigMap {
	return consumer.BuildConsumerConfigMap(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue)
}

func (r *SwimDigitalNotamConsumerReconciler) ClientCertificate(cr *appsv1alpha1.SwimDigitalNotamConsumer) *certmanagerv1.Certificate {
	return consumer.BuildConsumerCertificate(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue)
}

func (r *SwimDigitalNotamConsumerReconciler) ClientService(cr *appsv1alpha1.SwimDigitalNotamConsumer) *corev1.Service {
	return consumer.BuildConsumerClientService(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue)
}

func (r *SwimDigitalNotamConsumerReconciler) ClientServiceMonitor(cr *appsv1alpha1.SwimDigitalNotamConsumer) *monitoringv1.ServiceMonitor {
	return consumer.BuildConsumerServiceMonitor(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue)
}

func (r *SwimDigitalNotamConsumerReconciler) ClientServiceAccount(cr *appsv1alpha1.SwimDigitalNotamConsumer) *corev1.ServiceAccount {
	return consumer.BuildConsumerServiceAccount(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue)
}

func (r *SwimDigitalNotamConsumerReconciler) ClientRole(cr *appsv1alpha1.SwimDigitalNotamConsumer) *rbacv1.Role {
	return consumer.BuildConsumerRole(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue)
}

func (r *SwimDigitalNotamConsumerReconciler) ClientRoleBinding(cr *appsv1alpha1.SwimDigitalNotamConsumer) *rbacv1.RoleBinding {
	return consumer.BuildConsumerRoleBinding(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue)
}

func (r *SwimDigitalNotamConsumerReconciler) ClientDeployment(cr *appsv1alpha1.SwimDigitalNotamConsumer, configHash string) *appsv1.Deployment {
	return consumer.BuildConsumerClientDeployment(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue, configHash)
}

func (r *SwimDigitalNotamConsumerReconciler) ClientHPA(cr *appsv1alpha1.SwimDigitalNotamConsumer) *autoscalingv2.HorizontalPodAutoscaler {
	return consumer.BuildConsumerHPA(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue)
}

func (r *SwimDigitalNotamConsumerReconciler) KafkaConsole(cr *appsv1alpha1.SwimDigitalNotamConsumer) *unstructured.Unstructured {
	p := swimDigitalNotamConsumerBuildParams(cr)
	return consumer.BuildConsumerKafkaConsole(context.Background(), p, sharedManagedByValue, func(ctx context.Context, spec, ns string) string {
		return getOrDetectClusterDomain(ctx, r.Client, spec, ns)
	})
}

func (r *SwimDigitalNotamConsumerReconciler) KafkaNodePool(cr *appsv1alpha1.SwimDigitalNotamConsumer) *unstructured.Unstructured {
	return consumer.BuildConsumerKafkaNodePool(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue)
}

func (r *SwimDigitalNotamConsumerReconciler) KafkaCluster(cr *appsv1alpha1.SwimDigitalNotamConsumer) *unstructured.Unstructured {
	return consumer.BuildConsumerKafkaCluster(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue)
}

func (r *SwimDigitalNotamConsumerReconciler) KafkaTopic(cr *appsv1alpha1.SwimDigitalNotamConsumer, topicName string) *unstructured.Unstructured {
	return consumer.BuildConsumerKafkaTopic(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue, topicName)
}

func (r *SwimDigitalNotamConsumerReconciler) KafkaTopicDLQ(cr *appsv1alpha1.SwimDigitalNotamConsumer) *unstructured.Unstructured {
	return consumer.BuildConsumerKafkaTopic(swimDigitalNotamConsumerBuildParams(cr), sharedManagedByValue, "dnotam-events-dlq-topic")
}

func ocpFficeConsumerKafkaTopics() []string {
	return []string{
		"ffice-events-topic",
		"ffice-flight-plan-topic",
		"ffice-flight-update-topic",
		"ffice-filing-status-topic",
		"ffice-dlq-topic",
	}
}

func (r *SwimFficeConsumerReconciler) consumerPhaseConfig(req ctrl.Request, cr *appsv1alpha1.SwimFficeConsumer) consumer.ConsumerPhaseConfig {
	return consumer.ConsumerPhaseConfig{
		Client:         r.Client,
		Scheme:         r.Scheme,
		Owner:          cr,
		Request:        req,
		FinalizerName:  constants.FficeConsumerFinalizerName,
		CRKind:         "SwimFficeConsumer",
		BuildParams:    swimFficeConsumerBuildParams(cr),
		KafkaTopics:    ocpFficeConsumerKafkaTopics(),
		ManagedByLabel: sharedManagedByLabel,
		ManagedByValue: sharedManagedByValue,
		ResolveClusterDomain: func(ctx context.Context, specDomain, namespace string) string {
			return getOrDetectClusterDomain(ctx, r.Client, specDomain, namespace)
		},
		RemoveFinalizer: resources.MakeRemoveFinalizerFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimFficeConsumer { return &appsv1alpha1.SwimFficeConsumer{} },
			constants.FficeConsumerFinalizerName,
		),
		ApplyStatus: resources.MakeApplyStatusFunc(
			r.Client, req.NamespacedName,
			func() *appsv1alpha1.SwimFficeConsumer { return &appsv1alpha1.SwimFficeConsumer{} },
			func(o *appsv1alpha1.SwimFficeConsumer) *[]metav1.Condition { return &o.Status.Conditions },
		),
	}
}

func (r *SwimFficeConsumerReconciler) handleFficeConsumerFinalization(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimFficeConsumer) (ctrl.Result, error) {
	return consumer.HandleConsumerFinalization(ctx, r.consumerPhaseConfig(req, cr))
}

func (r *SwimFficeConsumerReconciler) ensureFficeConsumerFinalizer(ctx context.Context, cr *appsv1alpha1.SwimFficeConsumer, req ctrl.Request) (ctrl.Result, error) {
	return consumer.EnsureConsumerFinalizer(ctx, r.consumerPhaseConfig(req, cr))
}

func (r *SwimFficeConsumerReconciler) reconcileFficeConsumerKafka(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimFficeConsumer) (ctrl.Result, error) {
	return consumer.ReconcileConsumerKafkaPhase(ctx, r.consumerPhaseConfig(req, cr))
}

func (r *SwimFficeConsumerReconciler) reconcileFficeConsumerRBAC(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimFficeConsumer) error {
	return consumer.ReconcileConsumerRBACPhase(ctx, r.consumerPhaseConfig(req, cr))
}

func (r *SwimFficeConsumerReconciler) reconcileFficeConsumerSecrets(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimFficeConsumer) consumer.ConsumerSecretsBundle {
	return consumer.ReconcileConsumerSecretsBundle(ctx, r.consumerPhaseConfig(req, cr))
}

func (r *SwimFficeConsumerReconciler) reconcileFficeConsumerMongoDB(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimFficeConsumer) (ctrl.Result, error) {
	return consumer.ReconcileConsumerMongoPhase(ctx, r.consumerPhaseConfig(req, cr))
}

func (r *SwimFficeConsumerReconciler) reconcileFficeConsumerClient(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimFficeConsumer, bundle consumer.ConsumerSecretsBundle) (ctrl.Result, error) {
	return consumer.ReconcileConsumerClientPhase(ctx, r.consumerPhaseConfig(req, cr), bundle)
}

func (r *SwimFficeConsumerReconciler) updateFficeConsumerStatus(ctx context.Context, req ctrl.Request, cr *appsv1alpha1.SwimFficeConsumer, conditionType string, status metav1.ConditionStatus, reason, message string) error {
	cfg := r.consumerPhaseConfig(req, cr)
	if cfg.ApplyStatus == nil {
		return nil
	}
	return cfg.ApplyStatus(ctx, metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})
}
