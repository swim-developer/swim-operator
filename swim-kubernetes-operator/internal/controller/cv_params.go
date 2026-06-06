package controller

import (
	appsv1alpha1 "github.com/swim-developer/swim-kubernetes-operator/api/v1alpha1"
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	"github.com/swim-developer/swim-operator-common/pkg/controller/cv"
)

func dnotamK8sCoreSpec(cr *appsv1alpha1.SwimDnotamConsumerValidator) commonapi.SwimDnotamConsumerValidatorSpec {
	return commonapi.SwimDnotamConsumerValidatorSpec{
		ReplicaCount: cr.Spec.ReplicaCount,
		Global:       cr.Spec.Global,
		AppConfig:    cr.Spec.AppConfig,
		MariaDB:      cr.Spec.MariaDB,
		Artemis:      cr.Spec.Artemis,
		CertManager:  cr.Spec.CertManager,
		Image:        cr.Spec.Image,
		HPA:          cr.Spec.HPA,
	}
}

func ed254K8sCoreSpec(cr *appsv1alpha1.SwimEd254ConsumerValidator) commonapi.SwimEd254ConsumerValidatorSpec {
	return commonapi.SwimEd254ConsumerValidatorSpec{
		ReplicaCount: cr.Spec.ReplicaCount,
		Global:       cr.Spec.Global,
		AppConfig:    cr.Spec.AppConfig,
		MariaDB:      cr.Spec.MariaDB,
		Artemis:      cr.Spec.Artemis,
		CertManager:  cr.Spec.CertManager,
		Image:        cr.Spec.Image,
		HPA:          cr.Spec.HPA,
	}
}

func dnotamK8sCVBuildParams(cr *appsv1alpha1.SwimDnotamConsumerValidator) cv.CVBuildParams {
	return cv.CVBuildParams{
		Flavor:            cv.CVFlavorDnotam,
		CRName:            cr.Name,
		Namespace:         cr.Namespace,
		Spec:              dnotamK8sCoreSpec(cr),
		DefaultImageRepo:  "quay.io/masales/swim-dnotam-consumer-validator",
		DefaultDatabase:   "swim_consumer_validator",
		ArtemisExposeMode: "ingress",
		Ingress: cv.CVIngressParams{
			Enabled:             cr.Spec.Ingress.Enabled,
			HostOverride:        cr.Spec.Ingress.Host,
			ArtemisHostOverride: cr.Spec.Ingress.ArtemisHost,
			APIHostOverride:     cr.Spec.Ingress.APIHost,
			TLSSecretName:       cr.Spec.Ingress.TLSSecretName,
			Annotations:         cr.Spec.Ingress.Annotations,
		},
	}
}

func fficeK8sCoreSpec(cr *appsv1alpha1.SwimFficeConsumerValidator) commonapi.SwimFficeConsumerValidatorSpec {
	return commonapi.SwimFficeConsumerValidatorSpec{
		ReplicaCount: cr.Spec.ReplicaCount,
		Global:       cr.Spec.Global,
		AppConfig:    cr.Spec.AppConfig,
		MariaDB:      cr.Spec.MariaDB,
		Artemis:      cr.Spec.Artemis,
		CertManager:  cr.Spec.CertManager,
		Image:        cr.Spec.Image,
		HPA:          cr.Spec.HPA,
	}
}

func fficeK8sCVBuildParams(cr *appsv1alpha1.SwimFficeConsumerValidator) cv.CVBuildParams {
	return cv.CVBuildParams{
		Flavor:            cv.CVFlavorFfice,
		CRName:            cr.Name,
		Namespace:         cr.Namespace,
		Spec:              cv.CvSpecFromFfice(fficeK8sCoreSpec(cr)),
		DefaultImageRepo:  "quay.io/masales/swim-ffice-consumer-validator",
		DefaultDatabase:   "swim_ffice_consumer_validator",
		ArtemisExposeMode: "ingress",
		Ingress: cv.CVIngressParams{
			Enabled:             cr.Spec.Ingress.Enabled,
			HostOverride:        cr.Spec.Ingress.Host,
			ArtemisHostOverride: cr.Spec.Ingress.ArtemisHost,
			APIHostOverride:     cr.Spec.Ingress.APIHost,
			TLSSecretName:       cr.Spec.Ingress.TLSSecretName,
			Annotations:         cr.Spec.Ingress.Annotations,
		},
	}
}

func ed254K8sCVBuildParams(cr *appsv1alpha1.SwimEd254ConsumerValidator) cv.CVBuildParams {
	return cv.CVBuildParams{
		Flavor:            cv.CVFlavorEd254,
		CRName:            cr.Name,
		Namespace:         cr.Namespace,
		Spec:              cv.CvSpecFromEd254(ed254K8sCoreSpec(cr)),
		DefaultImageRepo:  "quay.io/masales/swim-ed254-consumer-validator",
		DefaultDatabase:   "swim_ed254_consumer_validator",
		ArtemisExposeMode: "ingress",
		Ingress: cv.CVIngressParams{
			Enabled:             cr.Spec.Ingress.Enabled,
			HostOverride:        cr.Spec.Ingress.Host,
			ArtemisHostOverride: cr.Spec.Ingress.ArtemisHost,
			APIHostOverride:     cr.Spec.Ingress.APIHost,
			TLSSecretName:       cr.Spec.Ingress.TLSSecretName,
			Annotations:         cr.Spec.Ingress.Annotations,
		},
	}
}
