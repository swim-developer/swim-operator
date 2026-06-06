package cv

import (
	"fmt"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/labels"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildCVAppService(p CVBuildParams, managedBy string) *corev1.Service {
	lbl := labels.StandardLabels(p.CRName, constants.ConsumerValidatorApp, p.CRName, managedBy)
	httpPort := resources.Int32Default(p.Spec.AppConfig.Quarkus.HTTPPort, 8080)
	sslPort := resources.Int32Default(p.Spec.AppConfig.Quarkus.SSLPort, 8443)
	return resources.ServiceClusterIP(p.CRName, p.Namespace, lbl, lbl, []corev1.ServicePort{
		resources.ServicePortTCP("http", httpPort, httpPort),
		resources.ServicePortTCP("https", sslPort, sslPort),
	})
}

func BuildCVDeployment(p CVBuildParams, managedBy string, configHash string) *appsv1.Deployment {
	lbl := labels.StandardLabels(p.CRName, constants.ConsumerValidatorApp, p.CRName, managedBy)
	imageRepo := resources.StrDefault(p.Spec.Image.Repository, p.DefaultImageRepo)
	imageTag := resources.StrDefault(p.Spec.Image.Tag, "latest")
	pullPolicy := p.Spec.Image.PullPolicy
	if pullPolicy == "" {
		pullPolicy = corev1.PullAlways
	}
	replicaCount := int32(1)
	if p.Spec.ReplicaCount != nil {
		replicaCount = *p.Spec.ReplicaCount
	}
	httpPort := resources.Int32Default(p.Spec.AppConfig.Quarkus.HTTPPort, 8080)
	sslPort := resources.Int32Default(p.Spec.AppConfig.Quarkus.SSLPort, 8443)
	mariadbPort := resources.Int32Default(p.Spec.MariaDB.Port, 3306)
	mariadbHost := MariaDBServiceName(p.CRName)
	secretName := cvGetMariaDBSecretName(p)
	serverTLS := fmt.Sprintf(constants.ServerTLSSuffix, p.CRName)
	annotations := map[string]string{}
	if configHash != "" {
		annotations["config-hash"] = configHash
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: p.CRName, Namespace: p.Namespace, Labels: lbl},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicaCount,
			Selector: &metav1.LabelSelector{MatchLabels: lbl},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: lbl, Annotations: annotations},
				Spec: corev1.PodSpec{
					ServiceAccountName: p.CRName,
					InitContainers: []corev1.Container{{
						Name:    "wait-for-mariadb",
						Image:   "docker.io/mariadb:12-ubi",
						EnvFrom: []corev1.EnvFromSource{{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: secretName}}}},
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
						Name:            p.CRName,
						Image:           fmt.Sprintf("%s:%s", imageRepo, imageTag),
						ImagePullPolicy: pullPolicy,
						Ports: []corev1.ContainerPort{
							{ContainerPort: httpPort, Name: "http"},
							{ContainerPort: sslPort, Name: "https"},
						},
						EnvFrom: []corev1.EnvFromSource{
							{ConfigMapRef: &corev1.ConfigMapEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-config", p.CRName)}}},
							{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-amqp-credentials", p.CRName)}}},
							{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: secretName}}},
						},
						Env:       []corev1.EnvVar{resources.EnvLiteral("TZ", "UTC")},
						Resources: resources.ResourcesOrDefault(corev1.ResourceRequirements{}, "256Mi", "250m", "1Gi", "2"),
						VolumeMounts: []corev1.VolumeMount{
							{Name: "server-cert", MountPath: "/certs/server", ReadOnly: true},
							{Name: "ca-cert", MountPath: "/certs/ca", ReadOnly: true},
						},
						StartupProbe:   resources.BuildHTTPProbe("/q/health/started", int(httpPort), resources.ProbeOverrides{PeriodSeconds: 2, FailureThreshold: 30}),
						LivenessProbe:  resources.BuildHTTPProbe("/q/health/live", int(httpPort), resources.ProbeOverrides{PeriodSeconds: 10, TimeoutSeconds: 5, FailureThreshold: 3}),
						ReadinessProbe: resources.BuildHTTPProbe("/q/health/ready", int(httpPort), resources.ProbeOverrides{PeriodSeconds: 5, TimeoutSeconds: 3, FailureThreshold: 3}),
					}},
					Volumes: []corev1.Volume{
						{Name: "server-cert", VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: serverTLS}}},
						{Name: "ca-cert", VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: serverTLS}}},
					},
				},
			},
		},
	}
}

func BuildCVHPA(p CVBuildParams, managedBy string) *autoscalingv2.HorizontalPodAutoscaler {
	lbl := labels.StandardLabels(p.CRName, constants.ConsumerValidatorApp, p.CRName, managedBy)
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
