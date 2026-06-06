package cv

import (
	"context"
	"fmt"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/helpers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func EnsureCVFinalizer(ctx context.Context, cfg CVPhaseConfig) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(cfg.Owner, cfg.FinalizerName) {
		return ctrl.Result{}, nil
	}
	controllerutil.AddFinalizer(cfg.Owner, cfg.FinalizerName)
	if err := cfg.Client.Update(ctx, cfg.Owner); err != nil {
		if errors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}
	return ctrl.Result{Requeue: true}, nil
}

func HandleCVFinalization(ctx context.Context, cfg CVPhaseConfig) (ctrl.Result, error) {
	metaObj, ok := cfg.Owner.(metav1.Object)
	if !ok {
		return ctrl.Result{}, fmt.Errorf("owner does not implement metav1.Object")
	}
	if metaObj.GetDeletionTimestamp() == nil || metaObj.GetDeletionTimestamp().IsZero() {
		return ctrl.Result{}, nil
	}
	if !controllerutil.ContainsFinalizer(cfg.Owner, cfg.FinalizerName) {
		return ctrl.Result{}, nil
	}
	CleanupArtemisPVCs(ctx, cfg.Client, cfg.BuildParams.Namespace, cfg.BuildParams.CRName)
	if cfg.RemoveFinalizer != nil {
		if err := cfg.RemoveFinalizer(ctx); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}
	return ctrl.Result{}, nil
}

func CleanupArtemisPVCs(ctx context.Context, c client.Client, namespace, crName string) {
	logger := log.FromContext(ctx)
	artemisName := fmt.Sprintf(constants.ArtemisSuffix, crName)
	pvcList := &corev1.PersistentVolumeClaimList{}
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels{"ActiveMQArtemis": artemisName},
	}
	if err := c.List(ctx, pvcList, listOpts...); err != nil {
		logger.V(1).Info("Failed to list Artemis PVCs (best effort cleanup)", "error", err.Error())
		return
	}
	for i := range pvcList.Items {
		pvc := &pvcList.Items[i]
		if err := c.Delete(ctx, pvc); err != nil && !errors.IsNotFound(err) {
			logger.V(1).Info("Failed to delete Artemis PVC (best effort cleanup)", "pvc", pvc.Name, "error", err.Error())
		} else {
			logger.Info("Deleted Artemis PVC", "pvc", pvc.Name)
		}
	}
}

func IsCVPodReady(ctx context.Context, c client.Client, namespace string, lbls map[string]string) bool {
	return helpers.IsPodReady(ctx, c, namespace, lbls)
}
