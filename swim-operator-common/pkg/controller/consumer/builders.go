package consumer

import (
	"fmt"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/labels"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func consumerMongoDatabase(p ConsumerBuildParams) string {
	switch p.Flavor {
	case ConsumerFlavorDnotam:
		return p.Client.Mongo.Database
	case ConsumerFlavorFfice:
		return resources.StrDefault(p.FficeConsumer.Mongo.Database, "swim-ffice")
	default:
		return resources.StrDefault(p.Consumer.Mongo.Database, "swim-ed254")
	}
}

func consumerMongoUser(p ConsumerBuildParams) string {
	switch p.Flavor {
	case ConsumerFlavorDnotam:
		return p.Client.Mongo.User
	case ConsumerFlavorFfice:
		return p.FficeConsumer.Mongo.User
	default:
		return p.Consumer.Mongo.User
	}
}

func consumerMongoPassword(p ConsumerBuildParams) string {
	switch p.Flavor {
	case ConsumerFlavorDnotam:
		return p.Client.Mongo.Password
	case ConsumerFlavorFfice:
		return p.FficeConsumer.Mongo.Password
	default:
		return p.Consumer.Mongo.Password
	}
}

func consumerMongoStorageSize(p ConsumerBuildParams) string {
	switch p.Flavor {
	case ConsumerFlavorDnotam:
		return p.Client.Mongo.StorageSize
	case ConsumerFlavorFfice:
		return p.FficeConsumer.Mongo.StorageSize
	default:
		return p.Consumer.Mongo.StorageSize
	}
}

func consumerMongoResources(p ConsumerBuildParams) corev1.ResourceRequirements {
	switch p.Flavor {
	case ConsumerFlavorDnotam:
		return p.Client.Mongo.Resources
	case ConsumerFlavorFfice:
		return p.FficeConsumer.Mongo.Resources
	default:
		return p.Consumer.Mongo.Resources
	}
}

func consumerDefaultClientImage(p ConsumerBuildParams) string {
	switch p.Flavor {
	case ConsumerFlavorDnotam:
		return "quay.io/masales/swim-dnotam-consumer:latest"
	case ConsumerFlavorFfice:
		return "quay.io/masales/swim-ffice-consumer:latest"
	default:
		return "quay.io/masales/swim-ed254-consumer:latest"
	}
}

func consumerClientImage(p ConsumerBuildParams) string {
	switch p.Flavor {
	case ConsumerFlavorDnotam:
		return resources.StrDefault(p.Client.Image, consumerDefaultClientImage(p))
	case ConsumerFlavorFfice:
		return resources.StrDefault(p.FficeConsumer.Image, consumerDefaultClientImage(p))
	default:
		return resources.StrDefault(p.Consumer.Image, consumerDefaultClientImage(p))
	}
}

func consumerClientReplicas(p ConsumerBuildParams) int32 {
	switch p.Flavor {
	case ConsumerFlavorDnotam:
		return resources.Int32Default(p.Client.Replicas, 1)
	case ConsumerFlavorFfice:
		return resources.Int32Default(p.FficeConsumer.Replicas, 1)
	default:
		return resources.Int32Default(p.Consumer.Replicas, 1)
	}
}

func consumerClientResources(p ConsumerBuildParams) corev1.ResourceRequirements {
	switch p.Flavor {
	case ConsumerFlavorDnotam:
		return p.Client.Resources
	case ConsumerFlavorFfice:
		return p.FficeConsumer.Resources
	default:
		return p.Consumer.Resources
	}
}

func consumerClientProbe(p ConsumerBuildParams) commonapi.ProbeConfig {
	switch p.Flavor {
	case ConsumerFlavorDnotam:
		return p.Client.Probe
	case ConsumerFlavorFfice:
		return p.FficeConsumer.Probe
	default:
		return p.Consumer.Probe
	}
}

func consumerProviders(p ConsumerBuildParams) []commonapi.ProviderSpec {
	switch p.Flavor {
	case ConsumerFlavorDnotam:
		return p.Client.Config.Providers
	case ConsumerFlavorFfice:
		return p.FficeConsumer.Config.Providers
	default:
		return p.Consumer.Config.Providers
	}
}

func consumerServiceMonitorEnabled(p ConsumerBuildParams) bool {
	switch p.Flavor {
	case ConsumerFlavorDnotam:
		return p.Client.Config.Observability.ServiceMonitorEnabled
	case ConsumerFlavorFfice:
		return p.FficeConsumer.Config.Observability.ServiceMonitorEnabled
	default:
		return p.Consumer.Config.Observability.ServiceMonitorEnabled
	}
}

func BuildConsumerKeystorePasswordSecret(p ConsumerBuildParams, managedBy string) *corev1.Secret {
	return resources.SecretStringData(fmt.Sprintf(constants.KeystorePasswordSuffix, p.Name), p.Namespace, labels.StandardLabels(p.Name, "client", p.Name, managedBy), map[string]string{"password": "changeit"})
}

func BuildConsumerKafkaCredentialsSecret(p ConsumerBuildParams, managedBy string) *corev1.Secret {
	deploymentMode := p.Kafka.DeploymentMode
	if deploymentMode == "" {
		deploymentMode = "managed"
	}
	username := ""
	password := ""
	if deploymentMode == "external" {
		username = p.Kafka.Username
		password = p.Kafka.Password
	}
	return resources.SecretStringData(fmt.Sprintf("%s-kafka-credentials", p.Name), p.Namespace, labels.StandardLabels(p.Name, "kafka", p.Name, managedBy), map[string]string{
		"KAFKA_USERNAME": username,
		"KAFKA_PASSWORD": password,
	})
}

func BuildConsumerMongoSecret(p ConsumerBuildParams, managedBy string) *corev1.Secret {
	mongoName := fmt.Sprintf(constants.MongoDBSuffix, p.Name)
	return resources.BuildMongoSecret(resources.MongoParams{
		Name:               mongoName,
		Namespace:          p.Namespace,
		Labels:             labels.StandardLabels(mongoName, "mongodb", p.Name, managedBy),
		User:               consumerMongoUser(p),
		Password:           consumerMongoPassword(p),
		Database:           consumerMongoDatabase(p),
		StorageSize:        consumerMongoStorageSize(p),
		Resources:          consumerMongoResources(p),
		ServiceAccountName: p.Name,
		CredentialsSecret:  fmt.Sprintf(constants.MongoDBCredentialsSuffix, p.Name),
		DataPVCName:        fmt.Sprintf(constants.MongoDBDataSuffix, p.Name),
	})
}

func BuildConsumerMongoPVC(p ConsumerBuildParams, managedBy string) *corev1.PersistentVolumeClaim {
	mongoName := fmt.Sprintf(constants.MongoDBSuffix, p.Name)
	return resources.PVC(fmt.Sprintf(constants.MongoDBDataSuffix, p.Name), p.Namespace, labels.StandardLabels(mongoName, "mongodb", p.Name, managedBy), resources.StrDefault(consumerMongoStorageSize(p), "1Gi"))
}

func BuildConsumerMongoService(p ConsumerBuildParams, managedBy string) *corev1.Service {
	mongoName := fmt.Sprintf(constants.MongoDBSuffix, p.Name)
	return resources.ServiceClusterIP(mongoName, p.Namespace, labels.StandardLabels(mongoName, "mongodb", p.Name, managedBy), map[string]string{"app": mongoName}, []corev1.ServicePort{resources.ServicePort("mongo", 27017, 27017)})
}

func BuildConsumerMongoDeployment(p ConsumerBuildParams, managedBy string, configHash string) *appsv1.Deployment {
	mongoName := fmt.Sprintf(constants.MongoDBSuffix, p.Name)
	return resources.BuildMongoDeployment(resources.MongoParams{
		Name:               mongoName,
		Namespace:          p.Namespace,
		Labels:             labels.StandardLabels(mongoName, "mongodb", p.Name, managedBy),
		User:               consumerMongoUser(p),
		Password:           consumerMongoPassword(p),
		Database:           consumerMongoDatabase(p),
		StorageSize:        consumerMongoStorageSize(p),
		Resources:          consumerMongoResources(p),
		ServiceAccountName: p.Name,
		CredentialsSecret:  fmt.Sprintf(constants.MongoDBCredentialsSuffix, p.Name),
		DataPVCName:        fmt.Sprintf(constants.MongoDBDataSuffix, p.Name),
	}, configHash)
}

func BuildConsumerProvidersSecret(p ConsumerBuildParams, managedBy string) *corev1.Secret {
	keystorePassword := "changeit"
	return resources.SecretStringData(fmt.Sprintf("%s-providers", p.Name), p.Namespace, labels.StandardLabels(p.Name, "client", p.Name, managedBy), map[string]string{
		"SWIM_PROVIDERS": resources.SerializeProviders(consumerProviders(p), keystorePassword),
	})
}

func BuildConsumerConfigMap(p ConsumerBuildParams, managedBy string) *corev1.ConfigMap {
	var data map[string]string
	switch p.Flavor {
	case ConsumerFlavorDnotam:
		data = resources.BuildDnotamConsumerConfigMapData(resources.DnotamConsumerConfigMapParams{
			Namespace:                     p.Namespace,
			CRName:                        p.Name,
			KafkaDeploymentMode:           p.Kafka.DeploymentMode,
			KafkaBootstrapExternal:        p.Kafka.BootstrapServers,
			MongoDatabase:                 p.Client.Mongo.Database,
			DnotamSubscriptionsSerialized: resources.SerializeDnotamSubscriptions(p.Client.Config.DnotamSubscriptions),
			SwimValidationEnabled:         p.Client.Config.SwimValidationEnabled,
			SwimValidationFailOnNullBody:  p.Client.Config.SwimValidationFailOnNullBody,
			DnotamDeleteAndRecreate:       p.Client.Config.DnotamDeleteAndRecreate,
			OpenTelemetryEnabled:          p.Client.Config.Observability.OpenTelemetryEnabled,
			OtelEndpoint:                  p.Client.Config.Observability.OtelEndpoint,
			OtelHeaders:                   p.Client.Config.Observability.OtelHeaders,
			PrometheusEnabled:             p.Client.Config.Observability.PrometheusEnabled,
		})
	case ConsumerFlavorEd254:
		data = resources.BuildEd254ConsumerConfigMapData(resources.Ed254ConsumerConfigMapParams{
			Namespace:                    p.Namespace,
			CRName:                       p.Name,
			KafkaDeploymentMode:          p.Kafka.DeploymentMode,
			KafkaBootstrapExternal:       p.Kafka.BootstrapServers,
			MongoDatabase:                resources.StrDefault(p.Consumer.Mongo.Database, "swim-ed254"),
			Ed254SubscriptionsSerialized: resources.SerializeEd254Subscriptions(p.Consumer.Config.Ed254Subscriptions),
			SwimValidationEnabled:        p.Consumer.Config.SwimValidationEnabled,
			OpenTelemetryEnabled:         p.Consumer.Config.Observability.OpenTelemetryEnabled,
			OtelEndpoint:                 p.Consumer.Config.Observability.OtelEndpoint,
			OtelHeaders:                  p.Consumer.Config.Observability.OtelHeaders,
			PrometheusEnabled:            p.Consumer.Config.Observability.PrometheusEnabled,
			HeartbeatTimeoutSeconds:      p.Consumer.Config.HeartbeatTimeoutSeconds,
		})
	case ConsumerFlavorFfice:
		data = resources.BuildFficeConsumerConfigMapData(resources.FficeConsumerConfigMapParams{
			Namespace:               p.Namespace,
			CRName:                  p.Name,
			KafkaDeploymentMode:     p.Kafka.DeploymentMode,
			KafkaBootstrapExternal:  p.Kafka.BootstrapServers,
			MongoDatabase:           resources.StrDefault(p.FficeConsumer.Mongo.Database, "swim-ffice"),
			SwimValidationEnabled:   p.FficeConsumer.Config.SwimValidationEnabled,
			OpenTelemetryEnabled:    p.FficeConsumer.Config.Observability.OpenTelemetryEnabled,
			OtelEndpoint:            p.FficeConsumer.Config.Observability.OtelEndpoint,
			OtelHeaders:             p.FficeConsumer.Config.Observability.OtelHeaders,
			PrometheusEnabled:       p.FficeConsumer.Config.Observability.PrometheusEnabled,
			HeartbeatTimeoutSeconds: p.FficeConsumer.Config.HeartbeatTimeoutSeconds,
		})
	}
	return resources.ConfigMap(fmt.Sprintf("%s-config", p.Name), p.Namespace, labels.StandardLabels(p.Name, "client", p.Name, managedBy), data)
}

func BuildConsumerCertificate(p ConsumerBuildParams, managedBy string) *certmanagerv1.Certificate {
	return resources.BuildMTLSCertificate(p.Name, p.Namespace, labels.StandardLabels(p.Name, "client", p.Name, managedBy), p.CertManager.IssuerName, p.CertManager.IssuerKind, fmt.Sprintf(constants.KeystorePasswordSuffix, p.Name))
}

func BuildConsumerClientService(p ConsumerBuildParams, managedBy string) *corev1.Service {
	lbl := labels.StandardLabels(p.Name, "client", p.Name, managedBy)
	return resources.ServiceClusterIP(p.Name, p.Namespace, lbl, map[string]string{"app": p.Name}, []corev1.ServicePort{
		resources.ServicePortTCP("http", 8080, 8080),
		resources.ServicePortTCP("management", 9000, 9000),
	})
}

func BuildConsumerServiceMonitor(p ConsumerBuildParams, managedBy string) *monitoringv1.ServiceMonitor {
	return resources.BuildServiceMonitor(resources.ServiceMonitorParams{
		Name:        fmt.Sprintf("%s-metrics", p.Name),
		Namespace:   p.Namespace,
		Labels:      labels.StandardLabels(p.Name, "monitoring", p.Name, managedBy),
		SelectorApp: p.Name,
		PortName:    "management",
		Path:        "/q/metrics",
		Interval:    "30s",
	})
}

func BuildConsumerServiceAccount(p ConsumerBuildParams, managedBy string) *corev1.ServiceAccount {
	return resources.StandardServiceAccount(p.Name, p.Namespace, labels.StandardLabels(p.Name, "client", p.Name, managedBy))
}

func BuildConsumerRole(p ConsumerBuildParams, managedBy string) *rbacv1.Role {
	return resources.BuildLeaseRole(p.Name, p.Namespace, labels.StandardLabels(p.Name, "client", p.Name, managedBy))
}

func BuildConsumerRoleBinding(p ConsumerBuildParams, managedBy string) *rbacv1.RoleBinding {
	return resources.BuildRoleBinding(p.Name, p.Namespace, p.Name, fmt.Sprintf("%s-role", p.Name), labels.StandardLabels(p.Name, "client", p.Name, managedBy))
}

func BuildConsumerClientDeployment(p ConsumerBuildParams, managedBy string, configHash string) *appsv1.Deployment {
	pr := consumerClientProbe(p)
	return resources.BuildSwimConsumerClientDeployment(resources.SwimConsumerClientDeploymentParams{
		Name:                   p.Name,
		Namespace:              p.Namespace,
		Labels:                 labels.StandardLabels(p.Name, "client", p.Name, managedBy),
		ServiceAccountName:     p.Name,
		Replicas:               consumerClientReplicas(p),
		Image:                  consumerClientImage(p),
		ContainerResources:     resources.ResourcesOrDefault(consumerClientResources(p), "512Mi", "500m", "1Gi", "2"),
		ConfigHash:             configHash,
		ConfigMapName:          fmt.Sprintf("%s-config", p.Name),
		ProvidersSecretName:    fmt.Sprintf("%s-providers", p.Name),
		KafkaCredentialsSecret: fmt.Sprintf("%s-kafka-credentials", p.Name),
		MongoCredentialsSecret: fmt.Sprintf(constants.MongoDBCredentialsSuffix, p.Name),
		KeystorePasswordSecret: fmt.Sprintf(constants.KeystorePasswordSuffix, p.Name),
		MTLSSecretName:         fmt.Sprintf("%s-mtls", p.Name),
		MtlsVolumeName:         constants.MTLSCertsVolume,
		LivenessProbe: resources.ProbeOverrides{
			InitialDelaySeconds: pr.InitialDelaySeconds,
			PeriodSeconds:       pr.PeriodSeconds,
			TimeoutSeconds:      pr.TimeoutSeconds,
			FailureThreshold:    pr.FailureThreshold,
			DefaultInitialDelay: 3,
			DefaultPeriod:       30,
			DefaultTimeout:      10,
			DefaultFailure:      5,
		},
		ReadinessProbe: resources.ProbeOverrides{
			InitialDelaySeconds: pr.InitialDelaySeconds,
			PeriodSeconds:       pr.PeriodSeconds,
			TimeoutSeconds:      pr.TimeoutSeconds,
			FailureThreshold:    pr.FailureThreshold,
			DefaultInitialDelay: 3,
			DefaultPeriod:       10,
			DefaultTimeout:      5,
			DefaultFailure:      5,
		},
	})
}

func BuildConsumerHPA(p ConsumerBuildParams, managedBy string) *autoscalingv2.HorizontalPodAutoscaler {
	minReplicas := p.HPA.MinReplicas
	if minReplicas == nil {
		def := int32(1)
		minReplicas = &def
	}
	return resources.BuildHPA(resources.HPAParams{
		Name:                           fmt.Sprintf("%s-hpa", p.Name),
		Namespace:                      p.Namespace,
		Labels:                         labels.StandardLabels(p.Name, "client", p.Name, managedBy),
		TargetName:                     p.Name,
		MinReplicas:                    minReplicas,
		MaxReplicas:                    resources.Int32Default(p.HPA.MaxReplicas, 5),
		CPUUtilization:                 p.HPA.TargetCPUUtilizationPercentage,
		TargetCPUUtilizationPercentage: 70,
		ScaleUpStabilization:           60,
		ScaleDownStabilization:         300,
	})
}
