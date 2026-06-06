package controller

import (
	"github.com/swim-developer/swim-operator-common/pkg/controller/cv"
	appsv1alpha1 "github.com/swim-developer/swim-openshift-operator/api/v1alpha1"
)

func dnotamCVBuildParams(cr *appsv1alpha1.SwimDnotamConsumerValidator) cv.CVBuildParams {
	return cv.CVBuildParams{
		Flavor:            cv.CVFlavorDnotam,
		CRName:            cr.Name,
		Namespace:         cr.Namespace,
		Spec:              cr.Spec,
		DefaultImageRepo:  "quay.io/masales/swim-dnotam-consumer-validator",
		DefaultDatabase:   "swim_consumer_validator",
		ArtemisExposeMode: "route",
	}
}

func ed254CVBuildParams(cr *appsv1alpha1.SwimEd254ConsumerValidator) cv.CVBuildParams {
	return cv.CVBuildParams{
		Flavor:            cv.CVFlavorEd254,
		CRName:            cr.Name,
		Namespace:         cr.Namespace,
		Spec:              cv.CvSpecFromEd254(cr.Spec),
		DefaultImageRepo:  "quay.io/masales/swim-ed254-consumer-validator",
		DefaultDatabase:   "swim_ed254_consumer_validator",
		ArtemisExposeMode: "route",
	}
}

func fficeCVBuildParams(cr *appsv1alpha1.SwimFficeConsumerValidator) cv.CVBuildParams {
	return cv.CVBuildParams{
		Flavor:            cv.CVFlavorFfice,
		CRName:            cr.Name,
		Namespace:         cr.Namespace,
		Spec:              cv.CvSpecFromFfice(cr.Spec),
		DefaultImageRepo:  "quay.io/masales/swim-ffice-consumer-validator",
		DefaultDatabase:   "swim_ffice_consumer_validator",
		ArtemisExposeMode: "route",
	}
}
