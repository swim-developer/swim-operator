package resources

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type MariaDBParams struct {
	Name               string
	Namespace          string
	Labels             map[string]string
	Database           string
	Username           string
	Password           string
	ServiceAccountName string
	SecretName         string
}

func BuildMariaDBSecret(p MariaDBParams) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: p.SecretName, Namespace: p.Namespace, Labels: p.Labels},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"MARIADB_USERNAME": StrDefault(p.Username, "swim"),
			"MARIADB_PASSWORD": StrDefault(p.Password, "swim"),
		},
	}
}

func BuildMariaDBService(name, namespace string, labels map[string]string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: labels},
		Spec: corev1.ServiceSpec{
			Selector:  map[string]string{"app": name},
			Ports:     []corev1.ServicePort{{Name: "mysql", Port: 3306, TargetPort: intstr.FromInt(3306), Protocol: corev1.ProtocolTCP}},
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
		},
	}
}

func BuildMariaDBStatefulSet(p MariaDBParams) *appsv1.StatefulSet {
	replicas := int32(1)
	database := StrDefault(p.Database, "swim_consumer_validator")
	username := StrDefault(p.Username, "swim")
	password := StrDefault(p.Password, "swim")

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: p.Name, Namespace: p.Namespace, Labels: p.Labels},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: p.Name,
			Replicas:    &replicas,
			Selector:    &metav1.LabelSelector{MatchLabels: map[string]string{"app": p.Name}},
			PersistentVolumeClaimRetentionPolicy: &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
				WhenDeleted: appsv1.DeletePersistentVolumeClaimRetentionPolicyType,
				WhenScaled:  appsv1.RetainPersistentVolumeClaimRetentionPolicyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: p.Labels},
				Spec: corev1.PodSpec{
					ServiceAccountName: p.ServiceAccountName,
					Containers: []corev1.Container{{
						Name:  "mariadb",
						Image: "docker.io/mariadb:12-ubi",
						Args: []string{
							"--innodb-buffer-pool-size=64M",
							"--character-set-server=utf8mb4",
							"--collation-server=utf8mb4_unicode_ci",
							"--socket=/tmp/mysql.sock",
						},
						Ports: []corev1.ContainerPort{{ContainerPort: 3306, Name: "mysql"}},
						Env: []corev1.EnvVar{
							EnvLiteral("TZ", "UTC"),
							EnvLiteral("MARIADB_ROOT_PASSWORD", password),
							EnvLiteral("MARIADB_DATABASE", database),
							EnvLiteral("MARIADB_USER", username),
							EnvLiteral("MARIADB_PASSWORD", password),
							EnvLiteral("MARIADB_MYSQL_LOCALHOST_USER", "1"),
						},
						VolumeMounts: []corev1.VolumeMount{{Name: "data", MountPath: "/var/lib/mysql", SubPath: "mysql"}},
					Resources: corev1.ResourceRequirements{},
						StartupProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								Exec: &corev1.ExecAction{
									Command: []string{"bash", "-c", "mariadb -u ${MARIADB_USER} -p${MARIADB_PASSWORD} --socket=/tmp/mysql.sock -e 'SELECT 1'"},
								},
							},
							PeriodSeconds: 10, FailureThreshold: 6,
						},
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								Exec: &corev1.ExecAction{
									Command: []string{"bash", "-c", "mariadb-admin -u ${MARIADB_USER} -p${MARIADB_PASSWORD} --socket=/tmp/mysql.sock ping"},
								},
							},
							PeriodSeconds: 10,
						},
					}},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{{
				ObjectMeta: metav1.ObjectMeta{Name: "data"},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("1Gi")},
					},
				},
			}},
		},
	}
}
