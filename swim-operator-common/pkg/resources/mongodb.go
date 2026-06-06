package resources

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type MongoParams struct {
	Name               string
	Namespace          string
	Labels             map[string]string
	User               string
	Password           string
	Database           string
	StorageSize        string
	Resources          corev1.ResourceRequirements
	ServiceAccountName string
	CredentialsSecret  string
	DataPVCName        string
}

func BuildMongoSecret(p MongoParams) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: p.CredentialsSecret, Namespace: p.Namespace, Labels: p.Labels},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"database-name":           StrDefault(p.Database, "swim-dnotam"),
			"database-user":           StrDefault(p.User, "swim"),
			"database-password":       StrDefault(p.Password, "swim123"),
			"database-admin-password": "admin",
		},
	}
}

func BuildMongoPVC(p MongoParams) *corev1.PersistentVolumeClaim {
	size := StrDefault(p.StorageSize, "1Gi")
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: p.DataPVCName, Namespace: p.Namespace, Labels: p.Labels},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(size)},
			},
		},
	}
}

func BuildMongoService(name, namespace string, labels map[string]string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: labels},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": name},
			Ports:    []corev1.ServicePort{{Name: "mongo", Port: 27017, TargetPort: intstr.FromInt(27017)}},
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
}

func BuildMongoDeployment(p MongoParams, configHash string) *appsv1.Deployment {
	replicas := int32(1)
	res := ResourcesOrDefault(p.Resources, "512Mi", "250m", "1Gi", "500m")

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: p.Name, Namespace: p.Namespace, Labels: p.Labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": p.Name}},
			Strategy: appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      p.Labels,
					Annotations: map[string]string{"config-hash": configHash},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: p.ServiceAccountName,
					Containers: []corev1.Container{{
						Name:      p.Name,
						Image:     "quay.io/mongodb/mongodb-community-server:8.0-ubi8",
						Ports:     []corev1.ContainerPort{{ContainerPort: 27017}},
						Resources: res,
						Env: []corev1.EnvVar{
							{Name: "TZ", Value: "UTC"},
							envFromSecret("MONGO_INITDB_DATABASE", p.CredentialsSecret, "database-name"),
							envFromSecret("MONGO_INITDB_ROOT_USERNAME", p.CredentialsSecret, "database-user"),
							envFromSecret("MONGO_INITDB_ROOT_PASSWORD", p.CredentialsSecret, "database-password"),
						},
						VolumeMounts: []corev1.VolumeMount{{Name: p.DataPVCName, MountPath: "/data/db"}},
						LivenessProbe: &corev1.Probe{
							ProbeHandler:        corev1.ProbeHandler{Exec: &corev1.ExecAction{Command: []string{"mongosh", "--eval", "db.adminCommand('ping')"}}},
							InitialDelaySeconds: 30, TimeoutSeconds: 5, PeriodSeconds: 10, FailureThreshold: 3,
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler:        corev1.ProbeHandler{Exec: &corev1.ExecAction{Command: []string{"mongosh", "--eval", "db.adminCommand('ping')"}}},
							InitialDelaySeconds: 10, TimeoutSeconds: 5, PeriodSeconds: 10, FailureThreshold: 3,
						},
					}},
					Volumes: []corev1.Volume{{Name: p.DataPVCName, VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: p.DataPVCName}}}},
				},
			},
		},
	}
}

func envFromSecret(envName, secretName, key string) corev1.EnvVar {
	return corev1.EnvVar{
		Name: envName,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
				Key:                  key,
			},
		},
	}
}

func EnvFromSecret(envName, secretName, key string) corev1.EnvVar {
	return envFromSecret(envName, secretName, key)
}

func EnvLiteral(name, value string) corev1.EnvVar {
	return corev1.EnvVar{Name: name, Value: value}
}

func EnvFromFieldRef(name, fieldPath string) corev1.EnvVar {
	return corev1.EnvVar{
		Name:      name,
		ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: fieldPath}},
	}
}

func SecretStringData(name, namespace string, labels map[string]string, data map[string]string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: labels},
		Type:       corev1.SecretTypeOpaque,
		StringData: data,
	}
}

func ConfigMap(name, namespace string, labels map[string]string, data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: labels},
		Data:       data,
	}
}

func ServiceClusterIP(name, namespace string, labels map[string]string, selector map[string]string, ports []corev1.ServicePort) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: labels},
		Spec: corev1.ServiceSpec{
			Selector: selector,
			Ports:    ports,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
}

func ServicePort(name string, port, targetPort int32) corev1.ServicePort {
	return corev1.ServicePort{
		Name: name, Port: port, TargetPort: intstr.FromInt32(targetPort),
	}
}

func ServicePortTCP(name string, port, targetPort int32) corev1.ServicePort {
	return corev1.ServicePort{
		Name: name, Port: port, TargetPort: intstr.FromInt32(targetPort), Protocol: corev1.ProtocolTCP,
	}
}

func PVC(name, namespace string, labels map[string]string, size string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: labels},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(size)},
			},
			VolumeMode: func() *corev1.PersistentVolumeMode { v := corev1.PersistentVolumeFilesystem; return &v }(),
		},
	}
}

func StandardServiceAccount(name, namespace string, labels map[string]string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: labels},
	}
}

func FormatMongoURI(host, port string) string {
	return fmt.Sprintf("mongodb://$(MONGODB_USERNAME):$(MONGODB_PASSWORD)@%s:%s", host, port)
}
