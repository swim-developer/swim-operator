package resources

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ProviderAppDeploymentParams struct {
	Name                  string
	Namespace             string
	Labels                map[string]string
	Image                 string
	Replicas              int32
	ContainerResources    corev1.ResourceRequirements
	ConfigHash            string
	ServerTLSSecretName   string
	CABundleConfigMapName string
}

func BuildProviderAppDeployment(p ProviderAppDeploymentParams) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Name,
			Namespace: p.Namespace,
			Labels:    p.Labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &p.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": p.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: p.Labels,
					Annotations: map[string]string{
						"config-hash": p.ConfigHash,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: p.Name,
					Containers: []corev1.Container{
						{
							Name:            p.Name,
							Image:           p.Image,
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{ContainerPort: 8080, Protocol: corev1.ProtocolTCP, Name: "http"},
								{ContainerPort: 8443, Protocol: corev1.ProtocolTCP, Name: "https"},
								{ContainerPort: 9000, Protocol: corev1.ProtocolTCP, Name: "management"},
								{ContainerPort: 9080, Protocol: corev1.ProtocolTCP, Name: "internal"},
							},
							EnvFrom: []corev1.EnvFromSource{
								{ConfigMapRef: &corev1.ConfigMapEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-config", p.Name)}}},
								{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-secret", p.Name)}}},
								{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-oidc-secret", p.Name)}}},
							},
							Env: []corev1.EnvVar{
								EnvLiteral("TZ", "UTC"),
								EnvFromFieldRef("KUBERNETES_NAMESPACE", "metadata.namespace"),
							},
							Resources: p.ContainerResources,
							VolumeMounts: []corev1.VolumeMount{
								{Name: "server-cert", MountPath: "/certs/server", ReadOnly: true},
								{Name: "ca-cert", MountPath: "/certs/ca", ReadOnly: true},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/q/health/live",
										Port: intstr.FromInt32(9000),
									},
								},
								InitialDelaySeconds: 3,
								PeriodSeconds:       30,
								TimeoutSeconds:      10,
								FailureThreshold:    5,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/q/health/ready",
										Port: intstr.FromInt32(9000),
									},
								},
								InitialDelaySeconds: 3,
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
								FailureThreshold:    5,
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "server-cert",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: p.ServerTLSSecretName,
								},
							},
						},
						{
							Name: "ca-cert",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: p.CABundleConfigMapName,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
