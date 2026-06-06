package resources

import (
	"fmt"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SwimConsumerClientDeploymentParams struct {
	Name                   string
	Namespace              string
	Labels                 map[string]string
	ServiceAccountName     string
	Replicas               int32
	Image                  string
	ContainerResources     corev1.ResourceRequirements
	ConfigHash             string
	ConfigMapName         string
	ProvidersSecretName    string
	KafkaCredentialsSecret string
	MongoCredentialsSecret string
	KeystorePasswordSecret string
	MTLSSecretName         string
	MtlsVolumeName         string
	LivenessProbe          ProbeOverrides
	ReadinessProbe         ProbeOverrides
}

func BuildSwimConsumerClientDeployment(p SwimConsumerClientDeploymentParams) *appsv1.Deployment {
	volName := p.MtlsVolumeName
	if volName == "" {
		volName = constants.MTLSCertsVolume
	}
	annotations := map[string]string{"config-hash": p.ConfigHash}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: p.Name, Namespace: p.Namespace, Labels: p.Labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: &p.Replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": p.Name}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      p.Labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: p.ServiceAccountName,
					InitContainers: []corev1.Container{{
						Name:  "validate-secrets",
						Image: "registry.access.redhat.com/ubi9/ubi-minimal:latest",
						Command: []string{
							"sh", "-c",
							"echo 'Validating certificate files...' && " +
								"test -f /secrets/truststore.p12 && echo '✓ truststore.p12 found' && " +
								"test -f /secrets/truststore.jks && echo '✓ truststore.jks found' && " +
								"test -f /secrets/keystore.p12 && echo '✓ keystore.p12 found' && " +
								"test -f /secrets/keystore.jks && echo '✓ keystore.jks found' && " +
								"test -f /secrets/ca.crt && echo '✓ ca.crt found' && " +
								"echo 'All certificate files validated successfully'",
						},
						VolumeMounts: []corev1.VolumeMount{{Name: volName, MountPath: "/secrets", ReadOnly: true}},
					}},
					Containers: []corev1.Container{{
						Name:            p.Name,
						Image:           p.Image,
						ImagePullPolicy: corev1.PullAlways,
						Ports: []corev1.ContainerPort{
							{Name: "http", ContainerPort: 8080},
							{Name: "management", ContainerPort: 9000},
						},
						Resources: p.ContainerResources,
						EnvFrom: []corev1.EnvFromSource{
							{ConfigMapRef: &corev1.ConfigMapEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: p.ConfigMapName}}},
							{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: p.ProvidersSecretName}}},
							{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: p.KafkaCredentialsSecret}}},
						},
						Env: []corev1.EnvVar{
							EnvLiteral("TZ", "UTC"),
							EnvFromSecret("MONGODB_USERNAME", p.MongoCredentialsSecret, constants.DatabaseUserKey),
							EnvFromSecret("MONGODB_PASSWORD", p.MongoCredentialsSecret, constants.DatabasePasswordKey),
							EnvLiteral("MONGODB_URI", FormatMongoURI("$(MONGODB_HOST)", "$(MONGODB_PORT)")),
							EnvLiteral("SWIM_TRUSTSTORE_PATH", "/secrets/truststore.p12"),
							EnvFromSecret("SWIM_TRUSTSTORE_PASSWORD", p.KeystorePasswordSecret, "password"),
							EnvLiteral("SWIM_CLIENT_KEYSTORE_PATH", "/secrets/keystore.p12"),
							EnvFromSecret("SWIM_CLIENT_KEYSTORE_PASSWORD", p.KeystorePasswordSecret, "password"),
							EnvFromFieldRef("POD_NAME", "metadata.name"),
							EnvFromFieldRef("POD_NAMESPACE", "metadata.namespace"),
						},
						VolumeMounts:   []corev1.VolumeMount{{Name: volName, MountPath: "/secrets", ReadOnly: true}},
						LivenessProbe:  BuildHTTPProbe("/q/health/live", 9000, p.LivenessProbe),
						ReadinessProbe: BuildHTTPProbe("/q/health/ready", 9000, p.ReadinessProbe),
					}},
					Volumes: []corev1.Volume{{
						Name:         volName,
						VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: p.MTLSSecretName}},
					}},
				},
			},
		},
	}
}

type DnotamConsumerConfigMapParams struct {
	Namespace                    string
	CRName                       string
	KafkaDeploymentMode          string
	KafkaBootstrapExternal       string
	MongoDatabase                string
	DnotamSubscriptionsSerialized string
	SwimValidationEnabled        string
	SwimValidationFailOnNullBody string
	DnotamDeleteAndRecreate      string
	OpenTelemetryEnabled         bool
	OtelEndpoint                 string
	OtelHeaders                  string
	PrometheusEnabled            bool
}

func BuildDnotamConsumerConfigMapData(p DnotamConsumerConfigMapParams) map[string]string {
	mode := StrDefault(p.KafkaDeploymentMode, "managed")
	var kafkaBootstrap string
	if mode == "external" {
		kafkaBootstrap = p.KafkaBootstrapExternal
	} else {
		kafkaBootstrap = fmt.Sprintf("kafka-kafka-bootstrap.%s.svc.cluster.local:9092", p.Namespace)
	}
	return map[string]string{
		"MONGODB_HOST":                     fmt.Sprintf("%s-mongodb.%s.svc.cluster.local", p.CRName, p.Namespace),
		"MONGODB_PORT":                     "27017",
		"MONGODB_DATABASE":               StrDefault(p.MongoDatabase, "swim-dnotam"),
		"KAFKA_BOOTSTRAP_SERVERS":        kafkaBootstrap,
		"DNOTAM_SUBSCRIPTIONS":           p.DnotamSubscriptionsSerialized,
		"SWIM_VALIDATION_ENABLED":        StrDefault(p.SwimValidationEnabled, "true"),
		"SWIM_VALIDATION_FAIL_ON_NULLBODY": StrDefault(p.SwimValidationFailOnNullBody, "false"),
		"DNOTAM_DELETE_AND_RECREATE":     StrDefault(p.DnotamDeleteAndRecreate, "true"),
		"OTEL_ENABLED":                   fmt.Sprintf("%t", p.OpenTelemetryEnabled),
		"OTEL_SDK_DISABLED":              fmt.Sprintf("%t", !p.OpenTelemetryEnabled),
		"OTEL_ENDPOINT":                  p.OtelEndpoint,
		"OTEL_HEADERS":                   StrDefault(p.OtelHeaders, ""),
		"PROMETHEUS_ENABLED":             fmt.Sprintf("%t", p.PrometheusEnabled),
	}
}

type Ed254ConsumerConfigMapParams struct {
	Namespace                     string
	CRName                        string
	KafkaDeploymentMode           string
	KafkaBootstrapExternal        string
	MongoDatabase                 string
	Ed254SubscriptionsSerialized  string
	SwimValidationEnabled         string
	OpenTelemetryEnabled          bool
	OtelEndpoint                  string
	OtelHeaders                   string
	PrometheusEnabled             bool
	HeartbeatTimeoutSeconds       int32
}

type FficeConsumerConfigMapParams struct {
	Namespace               string
	CRName                  string
	KafkaDeploymentMode     string
	KafkaBootstrapExternal  string
	MongoDatabase           string
	SwimValidationEnabled   string
	OpenTelemetryEnabled    bool
	OtelEndpoint            string
	OtelHeaders             string
	PrometheusEnabled       bool
	HeartbeatTimeoutSeconds int32
}

func BuildFficeConsumerConfigMapData(p FficeConsumerConfigMapParams) map[string]string {
	mode := StrDefault(p.KafkaDeploymentMode, "managed")
	var kafkaBootstrap string
	if mode == "external" {
		kafkaBootstrap = p.KafkaBootstrapExternal
	} else {
		kafkaBootstrap = fmt.Sprintf("kafka-kafka-bootstrap.%s.svc.cluster.local:9092", p.Namespace)
	}
	data := map[string]string{
		"MONGODB_HOST":            fmt.Sprintf("%s-mongodb.%s.svc.cluster.local", p.CRName, p.Namespace),
		"MONGODB_PORT":            "27017",
		"MONGODB_DATABASE":        StrDefault(p.MongoDatabase, "swim-ffice"),
		"KAFKA_BOOTSTRAP_SERVERS": kafkaBootstrap,
		"SWIM_VALIDATION_ENABLED": StrDefault(p.SwimValidationEnabled, "true"),
		"OTEL_ENABLED":            fmt.Sprintf("%t", p.OpenTelemetryEnabled),
		"OTEL_SDK_DISABLED":       fmt.Sprintf("%t", !p.OpenTelemetryEnabled),
		"OTEL_ENDPOINT":           p.OtelEndpoint,
		"OTEL_HEADERS":            StrDefault(p.OtelHeaders, ""),
		"PROMETHEUS_ENABLED":      fmt.Sprintf("%t", p.PrometheusEnabled),
	}
	if p.HeartbeatTimeoutSeconds > 0 {
		data["HEARTBEAT_TIMEOUT_SECONDS"] = fmt.Sprintf("%d", p.HeartbeatTimeoutSeconds)
	}
	return data
}

func BuildEd254ConsumerConfigMapData(p Ed254ConsumerConfigMapParams) map[string]string {
	mode := StrDefault(p.KafkaDeploymentMode, "managed")
	var kafkaBootstrap string
	if mode == "external" {
		kafkaBootstrap = p.KafkaBootstrapExternal
	} else {
		kafkaBootstrap = fmt.Sprintf("kafka-kafka-bootstrap.%s.svc.cluster.local:9092", p.Namespace)
	}
	data := map[string]string{
		"MONGODB_HOST":            fmt.Sprintf("%s-mongodb.%s.svc.cluster.local", p.CRName, p.Namespace),
		"MONGODB_PORT":            "27017",
		"MONGODB_DATABASE":        StrDefault(p.MongoDatabase, "swim-ed254"),
		"KAFKA_BOOTSTRAP_SERVERS": kafkaBootstrap,
		"ED254_SUBSCRIPTIONS":     p.Ed254SubscriptionsSerialized,
		"SWIM_VALIDATION_ENABLED": StrDefault(p.SwimValidationEnabled, "true"),
		"OTEL_ENABLED":            fmt.Sprintf("%t", p.OpenTelemetryEnabled),
		"OTEL_SDK_DISABLED":       fmt.Sprintf("%t", !p.OpenTelemetryEnabled),
		"OTEL_ENDPOINT":           p.OtelEndpoint,
		"OTEL_HEADERS":            StrDefault(p.OtelHeaders, ""),
		"PROMETHEUS_ENABLED":      fmt.Sprintf("%t", p.PrometheusEnabled),
	}
	if p.HeartbeatTimeoutSeconds > 0 {
		data["HEARTBEAT_TIMEOUT_SECONDS"] = fmt.Sprintf("%d", p.HeartbeatTimeoutSeconds)
	}
	return data
}
