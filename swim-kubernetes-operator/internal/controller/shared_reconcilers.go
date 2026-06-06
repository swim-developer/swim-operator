package controller

import (
	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
)

const sharedManagedByLabel = constants.SharedManagedByLabel
const sharedManagedByValue = "swim-kubernetes-operator"

func providerRequeueResult(result ctrl.Result) bool {
	return result.RequeueAfter > 0
}

func shouldHaltAfterProviderStep(result ctrl.Result, err error) (ctrl.Result, bool, error) {
	if err != nil {
		return result, true, err
	}
	if providerRequeueResult(result) {
		return result, true, nil
	}
	return ctrl.Result{}, false, nil
}

func shouldHaltAfterProviderStepEmptyOnError(result ctrl.Result, err error) (ctrl.Result, bool, error) {
	if err != nil {
		return ctrl.Result{}, true, err
	}
	if providerRequeueResult(result) {
		return result, true, nil
	}
	return ctrl.Result{}, false, nil
}

func serviceMonitorAvailable(mgr ctrl.Manager) bool {
	_, err := mgr.GetRESTMapper().RESTMapping(
		schema.GroupKind{Group: "monitoring.coreos.com", Kind: "ServiceMonitor"},
	)
	return err == nil
}
