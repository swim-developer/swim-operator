package provider

import (
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

type ProviderStrategy interface {
	CRKind() string
	ArtemisBrokerCleanupPrefix() string
	ArtemisSpecName() string
	ServiceMonitorEnabled() bool
	ReadyMessage() string
	NotReadyMessage() string
	AdditionalRoleRules() []rbacv1.PolicyRule
	Exposure() ProviderExposureSpec
	ConfigMapData(p ProviderBuildParams, clusterDomain string) map[string]string
	AppSecretData() map[string]string
	OIDCSecretData() map[string]string
	AppImage() string
	AppReplicas() int32
	AppResources() corev1.ResourceRequirements
	PostgresParams(p ProviderBuildParams, managedBy string) resources.PostgresParams
	ArtemisBaseParams(p ProviderBuildParams, ingressHost string) resources.ArtemisProviderParams
	ArtemisOIDCSecret(p ProviderBuildParams, managedBy string) *corev1.Secret
	ArtemisAddressBPSecret(p ProviderBuildParams, managedBy string) *corev1.Secret
	ArtemisSecurityBPSecret(p ProviderBuildParams, managedBy string) *corev1.Secret
	KafkaTopicAllName() string
	KafkaTopicDLQName() string
	KafkaTopicPartitions(topicName string) int64
	SkipIfKafkaExists() bool
}
