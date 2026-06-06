package controller

import (
	appsv1alpha1 "github.com/swim-developer/swim-kubernetes-operator/api/v1alpha1"
	"github.com/swim-developer/swim-operator-common/pkg/controller/consumer"
)

func swimDigitalNotamConsumerBuildParams(cr *appsv1alpha1.SwimDigitalNotamConsumer) consumer.ConsumerBuildParams {
	return consumer.ConsumerBuildParams{
		Flavor:              consumer.ConsumerFlavorDnotam,
		Name:                cr.Name,
		Namespace:           cr.Namespace,
		GlobalClusterDomain: cr.Spec.Global.ClusterDomain,
		Kafka:               cr.Spec.Kafka,
		CertManager:         cr.Spec.CertManager,
		Client:              cr.Spec.Client,
		HPA:                 cr.Spec.HPA,
	}
}

func swimEd254ConsumerBuildParams(cr *appsv1alpha1.SwimEd254Consumer) consumer.ConsumerBuildParams {
	return consumer.ConsumerBuildParams{
		Flavor:              consumer.ConsumerFlavorEd254,
		Name:                cr.Name,
		Namespace:           cr.Namespace,
		GlobalClusterDomain: cr.Spec.Global.ClusterDomain,
		Kafka:               cr.Spec.Kafka,
		CertManager:         cr.Spec.CertManager,
		Consumer:            cr.Spec.Consumer,
		HPA:                 cr.Spec.HPA,
	}
}

func swimFficeConsumerBuildParams(cr *appsv1alpha1.SwimFficeConsumer) consumer.ConsumerBuildParams {
	return consumer.ConsumerBuildParams{
		Flavor:              consumer.ConsumerFlavorFfice,
		Name:                cr.Name,
		Namespace:           cr.Namespace,
		GlobalClusterDomain: cr.Spec.Global.ClusterDomain,
		Kafka:               cr.Spec.Kafka,
		CertManager:         cr.Spec.CertManager,
		FficeConsumer:       cr.Spec.Consumer,
		HPA:                 cr.Spec.HPA,
	}
}
