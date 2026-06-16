package resources

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	defaultPostgresSecretDatabase = "swim-dnotam"
	defaultPostgresSecretUser     = "swim-provider"
	defaultPostgresSecretPassword = "swim-provider"
	postgresDataVolumeName        = "postgres-data"
	postgresContainerName         = "postgresql"
	pgIsreadyUser                 = "postgres"
	secretKeyDatabaseHost         = "database-host"
	secretKeyDatabaseName         = "database-name"
	secretKeyDatabaseUser         = "database-user"
	secretKeyDatabasePassword     = "database-password"
)

type PostgresParams struct {
	Name               string
	Namespace          string
	Labels             map[string]string
	Image              string
	StorageSize        string
	Database           string
	User               string
	Password           string
	Resources          corev1.ResourceRequirements
	ServiceAccountName string
	SecretName         string
	PVCName            string
}

func BuildPostgresSecret(p PostgresParams) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: p.SecretName, Namespace: p.Namespace, Labels: p.Labels},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{
			secretKeyDatabaseHost:     fmt.Sprintf("%s.%s.svc.cluster.local", p.Name, p.Namespace),
			secretKeyDatabaseName:     StrDefault(p.Database, defaultPostgresSecretDatabase),
			secretKeyDatabaseUser:     StrDefault(p.User, defaultPostgresSecretUser),
			secretKeyDatabasePassword: StrDefault(p.Password, defaultPostgresSecretPassword),
		},
	}
}

func BuildPostgresPVC(p PostgresParams) *corev1.PersistentVolumeClaim {
	return PVC(p.PVCName, p.Namespace, p.Labels, StrDefault(p.StorageSize, "5Gi"))
}

func BuildPostgresService(name, namespace string, labels map[string]string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: labels},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": name},
			Type:     corev1.ServiceTypeClusterIP,
			Ports:    []corev1.ServicePort{{Port: 5432, TargetPort: intstr.FromInt(5432), Protocol: corev1.ProtocolTCP}},
		},
	}
}

func BuildPostgresStatefulSet(p PostgresParams) *appsv1.StatefulSet {
	replicas := int32(1)
	image := StrDefault(p.Image, "registry.redhat.io/rhel9/postgresql-16:latest")
	res := p.Resources

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: p.Name, Namespace: p.Namespace, Labels: p.Labels},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: p.Name,
			Replicas:    &replicas,
			Selector:    &metav1.LabelSelector{MatchLabels: map[string]string{"app": p.Name}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: p.Labels},
				Spec: corev1.PodSpec{
					ServiceAccountName: p.ServiceAccountName,
					Containers: []corev1.Container{{
						Name:            postgresContainerName,
						Image:           image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Env: []corev1.EnvVar{
							EnvLiteral("TZ", "UTC"),
							EnvFromSecret("POSTGRESQL_USER", p.SecretName, secretKeyDatabaseUser),
							EnvFromSecret("POSTGRESQL_PASSWORD", p.SecretName, secretKeyDatabasePassword),
							EnvFromSecret("POSTGRESQL_ADMIN_PASSWORD", p.SecretName, secretKeyDatabasePassword),
							EnvFromSecret("POSTGRESQL_DATABASE", p.SecretName, secretKeyDatabaseName),
							EnvLiteral("POSTGRESQL_MAX_CONNECTIONS", "300"),
							EnvLiteral("POSTGRESQL_SHARED_BUFFERS", "256MB"),
							EnvLiteral("POSTGRESQL_EFFECTIVE_CACHE_SIZE", "512MB"),
							EnvLiteral("POSTGRESQL_WORK_MEM", "8MB"),
							EnvLiteral("POSTGRESQL_MAINTENANCE_WORK_MEM", "64MB"),
						},
						Ports:     []corev1.ContainerPort{{ContainerPort: 5432, Protocol: corev1.ProtocolTCP}},
						Resources: res,
						VolumeMounts: []corev1.VolumeMount{
							{Name: postgresDataVolumeName, MountPath: "/var/lib/pgsql/data"},
						},
						LivenessProbe: &corev1.Probe{
							ProbeHandler:        corev1.ProbeHandler{Exec: &corev1.ExecAction{Command: []string{"pg_isready", "-U", pgIsreadyUser}}},
							InitialDelaySeconds: 30, TimeoutSeconds: 5, PeriodSeconds: 10, FailureThreshold: 3,
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler:        corev1.ProbeHandler{Exec: &corev1.ExecAction{Command: []string{"pg_isready", "-U", pgIsreadyUser}}},
							InitialDelaySeconds: 10, TimeoutSeconds: 5, PeriodSeconds: 10, FailureThreshold: 3,
						},
					}},
					Volumes: []corev1.Volume{{
						Name:         postgresDataVolumeName,
						VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: p.PVCName}},
					}},
				},
			},
		},
	}
}

func BuildUpstreamPostgresStatefulSet(p PostgresParams) *appsv1.StatefulSet {
	replicas := int32(1)
	image := StrDefault(p.Image, "docker.io/postgres:16")
	res := p.Resources
	database := StrDefault(p.Database, defaultPostgresSecretDatabase)
	user := StrDefault(p.User, defaultPostgresSecretUser)
	password := StrDefault(p.Password, defaultPostgresSecretPassword)

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: p.Name, Namespace: p.Namespace, Labels: p.Labels},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: p.Name,
			Replicas:    &replicas,
			Selector:    &metav1.LabelSelector{MatchLabels: map[string]string{"app": p.Name}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: p.Labels},
				Spec: corev1.PodSpec{
					ServiceAccountName: p.ServiceAccountName,
					Containers: []corev1.Container{{
						Name:            postgresContainerName,
						Image:           image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Env: []corev1.EnvVar{
							EnvLiteral("TZ", "UTC"),
							EnvLiteral("POSTGRES_DB", database),
							EnvLiteral("POSTGRES_USER", user),
							EnvLiteral("POSTGRES_PASSWORD", password),
							EnvLiteral("PGDATA", "/var/lib/postgresql/data/pgdata"),
						},
						Ports:     []corev1.ContainerPort{{ContainerPort: 5432, Protocol: corev1.ProtocolTCP}},
						Resources: res,
						VolumeMounts: []corev1.VolumeMount{
							{Name: "postgres-data", MountPath: "/var/lib/postgresql/data"},
						},
						LivenessProbe: &corev1.Probe{
							ProbeHandler:        corev1.ProbeHandler{Exec: &corev1.ExecAction{Command: []string{"pg_isready", "-U", user}}},
							InitialDelaySeconds: 30, TimeoutSeconds: 5, PeriodSeconds: 10, FailureThreshold: 3,
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler:        corev1.ProbeHandler{Exec: &corev1.ExecAction{Command: []string{"pg_isready", "-U", user}}},
							InitialDelaySeconds: 10, TimeoutSeconds: 5, PeriodSeconds: 10, FailureThreshold: 3,
						},
					}},
					Volumes: []corev1.Volume{{
						Name:         postgresDataVolumeName,
						VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: p.PVCName}},
					}},
				},
			},
		},
	}
}

func BuildPostgresServiceHeadless(name, namespace string, labels map[string]string) *corev1.Service {
	svc := BuildPostgresService(name, namespace, labels)
	svc.Spec.ClusterIP = corev1.ClusterIPNone
	return svc
}

func PostgresHostFQDN(name, namespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace)
}

func PostgresJDBCURL(host, port, database string) string {
	return fmt.Sprintf("jdbc:postgresql://%s:%s/%s", host, port, database)
}

func PostgresR2DBCURL(host, port, database string) string {
	return fmt.Sprintf("r2dbc:postgresql://%s:%s/%s", host, port, database)
}

func MustParseResource(s string) resource.Quantity {
	return resource.MustParse(s)
}
