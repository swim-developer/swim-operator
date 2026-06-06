package pv

import (
	"fmt"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func BuildPVConfigMap(p PVBuildParams, managedBy string) *corev1.ConfigMap {
	lbl := StandardLabels(p.CRName, managedBy)
	mariadbHost := resources.StrDefault(p.Spec.MariaDB.Host, MariaDBServiceName(p.CRName))
	mariadbPort := resources.Int32Default(p.Spec.MariaDB.Port, 3306)
	mariadbDatabase := resources.StrDefault(p.Spec.MariaDB.Database, "swim_provider_validator")
	amqpPort := resources.Int32Default(p.Spec.AMQP.Port, 443)
	data := map[string]string{
		"KEYCLOAK_URL":            p.Spec.Keycloak.URL,
		"KEYCLOAK_REALM":          p.Spec.Keycloak.Realm,
		"KEYCLOAK_CLIENT_ID":      p.Spec.Keycloak.ClientID,
		"SWIM_PROVIDER_API_URLS":  p.Spec.ProviderAPIURLs,
		"MARIADB_HOST":            mariadbHost,
		"MARIADB_PORT":            fmt.Sprintf("%d", mariadbPort),
		"MARIADB_DATABASE":        mariadbDatabase,
		"SWIM_PROVIDER_AMQP_HOST": p.Spec.AMQP.Host,
		"SWIM_PROVIDER_AMQP_PORT": fmt.Sprintf("%d", amqpPort),
	}
	if p.Spec.MTLS.Enabled {
		data["PROXY_MTLS_KEYSTORE_PATH"] = "/secrets/keystore.p12"
		data["PROXY_MTLS_KEYSTORE_TYPE"] = "PKCS12"
		data["PROXY_MTLS_TRUSTSTORE_PATH"] = "/secrets/truststore.p12"
		data["PROXY_MTLS_TRUSTSTORE_TYPE"] = "PKCS12"
	}
	return resources.ConfigMap(fmt.Sprintf("%s-config", p.CRName), p.Namespace, lbl, data)
}

func BuildPVDeployment(p PVBuildParams, managedBy, configHash string) *appsv1.Deployment {
	lbl := StandardLabels(p.CRName, managedBy)
	image := p.DefaultImage
	if image == "" {
		image = "quay.io/masales/swim-dnotam-provider-validator:latest"
	}
	replicas := int32(1)
	containerResources := resources.ResourcesOrDefault(corev1.ResourceRequirements{}, "128Mi", "100m", "512Mi", "500m")
	mariadbSecretName := GetMariaDBSecretName(p)
	mariadbHost := resources.StrDefault(p.Spec.MariaDB.Host, MariaDBServiceName(p.CRName))
	mariadbPort := resources.Int32Default(p.Spec.MariaDB.Port, 3306)
	envFrom := []corev1.EnvFromSource{
		{ConfigMapRef: &corev1.ConfigMapEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-config", p.CRName)}}},
		{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: mariadbSecretName}}},
	}
	volumeMounts := []corev1.VolumeMount{{Name: "data", MountPath: "/data"}}
	volumes := []corev1.Volume{{
		Name: "data",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: fmt.Sprintf("%s-data", p.CRName)},
		},
	}}
	if p.Spec.MTLS.Enabled && p.Spec.MTLS.PasswordsSecretName != "" {
		opt := true
		envFrom = append(envFrom, corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: p.Spec.MTLS.PasswordsSecretName},
				Optional:             &opt,
			},
		})
	}
	if mtlsSecret := MTLSSecretName(p); mtlsSecret != "" {
		opt := true
		mode := int32(420)
		volumeMounts = append(volumeMounts, corev1.VolumeMount{Name: "client-certs", MountPath: "/secrets", ReadOnly: true})
		volumes = append(volumes, corev1.Volume{
			Name: "client-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{SecretName: mtlsSecret, Optional: &opt, DefaultMode: &mode},
			},
		})
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: p.CRName, Namespace: p.Namespace, Labels: lbl},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Strategy: appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType},
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": p.CRName}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: lbl, Annotations: map[string]string{"config-hash": configHash}},
				Spec: corev1.PodSpec{
					ServiceAccountName: p.CRName,
					InitContainers: []corev1.Container{{
						Name:    "wait-for-mariadb",
						Image:   "docker.io/mariadb:12-ubi",
						EnvFrom: []corev1.EnvFromSource{{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: mariadbSecretName}}}},
						Command: []string{"bash", "-c", fmt.Sprintf(`for i in $(seq 1 60); do
  if mariadb-admin ping -h %s -P %d -u"$MARIADB_USERNAME" -p"$MARIADB_PASSWORD" --connect-timeout=2 >/dev/null 2>&1; then
    echo "MariaDB is ready"
    exit 0
  fi
  echo "Waiting for MariaDB... attempt $i/60"
  sleep 2
done
echo "MariaDB not available after 120 seconds"
exit 1`, mariadbHost, mariadbPort)},
					}},
					Containers: []corev1.Container{{
						Name:            constants.ProviderValidatorApp,
						Image:           image,
						ImagePullPolicy: corev1.PullAlways,
						Ports:           []corev1.ContainerPort{{ContainerPort: 8080, Protocol: corev1.ProtocolTCP, Name: "http"}},
						EnvFrom:         envFrom,
						Env:             []corev1.EnvVar{resources.EnvLiteral("TZ", "UTC")},
						VolumeMounts:    volumeMounts,
						Resources:       containerResources,
						LivenessProbe: &corev1.Probe{
							ProbeHandler:        corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{Path: "/q/health/live", Port: intstr.FromInt(8080)}},
							InitialDelaySeconds: 30, PeriodSeconds: 30, TimeoutSeconds: 10, FailureThreshold: 5,
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler:        corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{Path: "/q/health/ready", Port: intstr.FromInt(8080)}},
							InitialDelaySeconds: 10, PeriodSeconds: 10, TimeoutSeconds: 5, FailureThreshold: 5,
						},
					}},
					Volumes: volumes,
				},
			},
		},
	}
}

func BuildPVService(p PVBuildParams, managedBy string) *corev1.Service {
	lbl := StandardLabels(p.CRName, managedBy)
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: p.CRName, Namespace: p.Namespace, Labels: lbl},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": p.CRName},
			Ports:    []corev1.ServicePort{{Name: "http", Port: 8080, TargetPort: intstr.FromInt(8080), Protocol: corev1.ProtocolTCP}},
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
}

func BuildPVHPA(p PVBuildParams, managedBy string) *autoscalingv2.HorizontalPodAutoscaler {
	lbl := StandardLabels(p.CRName, managedBy)
	name := fmt.Sprintf("%s-hpa", p.CRName)
	hpa := p.Spec.HPA
	if !hpa.Enabled {
		return resources.BuildSingletonHPA(name, p.Namespace, p.CRName, lbl)
	}
	minReplicas := hpa.MinReplicas
	if minReplicas == nil {
		def := int32(1)
		minReplicas = &def
	}
	return resources.BuildHPA(resources.HPAParams{
		Name:                           name,
		Namespace:                      p.Namespace,
		Labels:                         lbl,
		TargetName:                     p.CRName,
		MinReplicas:                    minReplicas,
		MaxReplicas:                    resources.Int32Default(hpa.MaxReplicas, 5),
		CPUUtilization:                 hpa.TargetCPUUtilizationPercentage,
		TargetCPUUtilizationPercentage: 70,
		ScaleUpStabilization:           60,
		ScaleDownStabilization:         300,
	})
}
