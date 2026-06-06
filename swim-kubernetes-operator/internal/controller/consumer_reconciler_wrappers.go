package controller

import (
	"context"

	appsv1alpha1 "github.com/swim-developer/swim-kubernetes-operator/api/v1alpha1"
	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/controller/consumer"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func k8sEd254ConsumerKafkaTopics() []string {
	return []string{
		"ed254-events-all-topic",
		"ed254-events-arrival-sequence-topic",
		"ed254-events-dlq-topic",
	}
}

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

func (r *SwimEd254ConsumerReconciler) consumerPhaseConfig(req ctrl.Request, cr *appsv1alpha1.SwimEd254Consumer) consumer.ConsumerPhaseConfig {
	return consumer.ConsumerPhaseConfig{
		Client:         r.Client,
		Scheme:         r.Scheme,
		Owner:          cr,
		Request:        req,
		FinalizerName:  constants.Ed254ConsumerFinalizerName,
		CRKind:         "SwimEd254Consumer",
		BuildParams:    swimEd254ConsumerBuildParams(cr),
		KafkaTopics:    k8sEd254ConsumerKafkaTopics(),
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

func k8sFficeConsumerKafkaTopics() []string {
	return consumer.FficeConsumerKafkaTopics()
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
		KafkaTopics:    k8sFficeConsumerKafkaTopics(),
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
