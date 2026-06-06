package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func ReconcileConsumerMongoPhase(ctx context.Context, cfg ConsumerPhaseConfig) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	p := cfg.BuildParams
	mb := cfg.ManagedByValue
	desiredMongoSecret := BuildConsumerMongoSecret(p, mb)
	if err := consumerReconcileSecret(ctx, cfg, desiredMongoSecret); err != nil {
		return ctrl.Result{}, err
	}
	if err := consumerReconcilePVC(ctx, cfg, BuildConsumerMongoPVC(p, mb)); err != nil {
		return ctrl.Result{}, err
	}
	mongoHash := resources.ComputeConfigHash(nil, desiredMongoSecret)
	if err := consumerReconcileDeployment(ctx, cfg, BuildConsumerMongoDeployment(p, mb, mongoHash)); err != nil {
		return ctrl.Result{}, err
	}
	if err := consumerReconcileService(ctx, cfg, BuildConsumerMongoService(p, mb)); err != nil {
		return ctrl.Result{}, err
	}
	mongoReady, err := consumerIsMongoReady(ctx, cfg)
	if err != nil {
		return ctrl.Result{}, err
	}
	if !mongoReady {
		logger.Info("MongoDB not ready yet, requeuing in 10 seconds")
		applyConsumerCondition(ctx, cfg, "MongoDBReady", metav1.ConditionFalse, "Provisioning", "Waiting for MongoDB to be ready")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	applyConsumerCondition(ctx, cfg, "MongoDBReady", metav1.ConditionTrue, "Ready", "MongoDB is ready and accepting connections")
	return ctrl.Result{}, nil
}

func consumerIsMongoReady(ctx context.Context, cfg ConsumerPhaseConfig) (bool, error) {
	p := cfg.BuildParams
	deployment := &appsv1.Deployment{}
	err := cfg.Client.Get(ctx, client.ObjectKey{Name: fmt.Sprintf(constants.MongoDBSuffix, p.Name), Namespace: p.Namespace}, deployment)
	if err != nil {
		return false, client.IgnoreNotFound(err)
	}
	if deployment.Status.Replicas == 0 {
		return false, nil
	}
	return deployment.Status.ReadyReplicas > 0 && deployment.Status.ReadyReplicas == deployment.Status.Replicas, nil
}
