package resources

import (
	"testing"

	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testHelpersCRShortName               = "swim-cr"
	testHelpersNamespace                 = "swim-ns"
	testHelpersQuarkusHealthLive         = "/q/health/live"
	testHelpersMySecret                  = "my-secret"
	testHelpersSecretKeyDatabaseName     = "database-name"
	testHelpersSecretKeyDatabaseUser     = "database-user"
	testHelpersSecretKeyDatabasePassword = "database-password"
	testHelpersSwimDnotam                = "swim-dnotam"
	testHelpersCustomDB                  = "custom-db"
	testHelpersSwimProviderCred          = "swim-provider"
	testHelpersMyProvider                = "my-provider"
	testHelpersTestLatestImage           = "test:latest"
)

func TestStrDefault_ReturnsValueWhenNonEmpty(t *testing.T) {
	g := NewWithT(t)
	g.Expect(StrDefault("hello", "fallback")).To(Equal("hello"))
}

func TestStrDefault_ReturnsFallbackWhenEmpty(t *testing.T) {
	g := NewWithT(t)
	g.Expect(StrDefault("", "fallback")).To(Equal("fallback"))
}

func TestInt32Default_ReturnsValueWhenPositive(t *testing.T) {
	g := NewWithT(t)
	g.Expect(Int32Default(42, 10)).To(Equal(int32(42)))
}

func TestInt32Default_ReturnsFallbackWhenZero(t *testing.T) {
	g := NewWithT(t)
	g.Expect(Int32Default(0, 10)).To(Equal(int32(10)))
}

func TestInt32Default_ReturnsFallbackWhenNegative(t *testing.T) {
	g := NewWithT(t)
	g.Expect(Int32Default(-1, 10)).To(Equal(int32(10)))
}

func TestInt64Default_ReturnsValueWhenNonZero(t *testing.T) {
	g := NewWithT(t)
	g.Expect(Int64Default(99, 5)).To(Equal(int64(99)))
}

func TestInt64Default_ReturnsFallbackWhenZero(t *testing.T) {
	g := NewWithT(t)
	g.Expect(Int64Default(0, 5)).To(Equal(int64(5)))
}

func TestResourcesOrDefault_ReturnsDefaultWhenEmpty(t *testing.T) {
	g := NewWithT(t)
	res := ResourcesOrDefault(corev1.ResourceRequirements{}, "512Mi", "250m", "1Gi", "1")
	g.Expect(res.Requests[corev1.ResourceMemory]).To(Equal(resource.MustParse("512Mi")))
	g.Expect(res.Requests[corev1.ResourceCPU]).To(Equal(resource.MustParse("250m")))
	g.Expect(res.Limits[corev1.ResourceMemory]).To(Equal(resource.MustParse("1Gi")))
	g.Expect(res.Limits[corev1.ResourceCPU]).To(Equal(resource.MustParse("1")))
}

func TestResourcesOrDefault_PreservesCustomValues(t *testing.T) {
	g := NewWithT(t)
	custom := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("2Gi")},
		Limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")},
	}
	res := ResourcesOrDefault(custom, "512Mi", "250m", "1Gi", "1")
	g.Expect(res.Requests[corev1.ResourceMemory]).To(Equal(resource.MustParse("2Gi")))
	g.Expect(res.Limits[corev1.ResourceMemory]).To(Equal(resource.MustParse("4Gi")))
}

func TestComputeConfigHash_DeterministicForSameInput(t *testing.T) {
	g := NewWithT(t)
	cm := &corev1.ConfigMap{Data: map[string]string{"KEY": "value"}}
	hash1 := ComputeConfigHash(cm)
	hash2 := ComputeConfigHash(cm)
	g.Expect(hash1).To(Equal(hash2))
}

func TestComputeConfigHash_DiffersForDifferentInput(t *testing.T) {
	g := NewWithT(t)
	cm1 := &corev1.ConfigMap{Data: map[string]string{"KEY": "value1"}}
	cm2 := &corev1.ConfigMap{Data: map[string]string{"KEY": "value2"}}
	g.Expect(ComputeConfigHash(cm1)).NotTo(Equal(ComputeConfigHash(cm2)))
}

func TestComputeConfigHash_IncludesSecrets(t *testing.T) {
	g := NewWithT(t)
	cm := &corev1.ConfigMap{Data: map[string]string{"K": "V"}}
	s1 := &corev1.Secret{Data: map[string][]byte{"pass": []byte("abc")}}
	s2 := &corev1.Secret{Data: map[string][]byte{"pass": []byte("xyz")}}
	hash1 := ComputeConfigHash(cm, s1)
	hash2 := ComputeConfigHash(cm, s2)
	g.Expect(hash1).NotTo(Equal(hash2))
}

func TestBuildHTTPProbe_UsesDefaults(t *testing.T) {
	g := NewWithT(t)
	probe := BuildHTTPProbe(testHelpersQuarkusHealthLive, 9000, ProbeOverrides{
		DefaultInitialDelay: 3,
		DefaultPeriod:       30,
		DefaultTimeout:      10,
		DefaultFailure:      5,
	})
	g.Expect(probe.HTTPGet.Path).To(Equal(testHelpersQuarkusHealthLive))
	g.Expect(probe.InitialDelaySeconds).To(Equal(int32(3)))
	g.Expect(probe.PeriodSeconds).To(Equal(int32(30)))
	g.Expect(probe.TimeoutSeconds).To(Equal(int32(10)))
	g.Expect(probe.FailureThreshold).To(Equal(int32(5)))
}

func TestBuildHTTPProbe_OverridesDefaults(t *testing.T) {
	g := NewWithT(t)
	probe := BuildHTTPProbe("/q/health/ready", 9000, ProbeOverrides{
		InitialDelaySeconds: 15,
		PeriodSeconds:       60,
		TimeoutSeconds:      20,
		FailureThreshold:    10,
		DefaultInitialDelay: 3,
		DefaultPeriod:       30,
		DefaultTimeout:      10,
		DefaultFailure:      5,
	})
	g.Expect(probe.InitialDelaySeconds).To(Equal(int32(15)))
	g.Expect(probe.PeriodSeconds).To(Equal(int32(60)))
	g.Expect(probe.TimeoutSeconds).To(Equal(int32(20)))
	g.Expect(probe.FailureThreshold).To(Equal(int32(10)))
}

func TestFormatMongoURI(t *testing.T) {
	g := NewWithT(t)
	uri := FormatMongoURI("mongo-host", "27017")
	g.Expect(uri).To(Equal("mongodb://$(MONGODB_USERNAME):$(MONGODB_PASSWORD)@mongo-host:27017"))
}

func TestEnvLiteral(t *testing.T) {
	g := NewWithT(t)
	env := EnvLiteral("TZ", "UTC")
	g.Expect(env.Name).To(Equal("TZ"))
	g.Expect(env.Value).To(Equal("UTC"))
	g.Expect(env.ValueFrom).To(BeNil())
}

func TestEnvFromSecret(t *testing.T) {
	g := NewWithT(t)
	env := EnvFromSecret("DB_PASS", testHelpersMySecret, "password")
	g.Expect(env.Name).To(Equal("DB_PASS"))
	g.Expect(env.ValueFrom.SecretKeyRef.Name).To(Equal(testHelpersMySecret))
	g.Expect(env.ValueFrom.SecretKeyRef.Key).To(Equal("password"))
}

func TestEnvFromFieldRef(t *testing.T) {
	g := NewWithT(t)
	env := EnvFromFieldRef("POD_NAME", "metadata.name")
	g.Expect(env.Name).To(Equal("POD_NAME"))
	g.Expect(env.ValueFrom.FieldRef.FieldPath).To(Equal("metadata.name"))
}

func TestSecretStringData(t *testing.T) {
	g := NewWithT(t)
	labels := map[string]string{"app": "test"}
	data := map[string]string{"key": "val"}
	s := SecretStringData(testHelpersMySecret, "ns", labels, data)
	g.Expect(s.Name).To(Equal(testHelpersMySecret))
	g.Expect(s.Namespace).To(Equal("ns"))
	g.Expect(s.Type).To(Equal(corev1.SecretTypeOpaque))
	g.Expect(s.StringData["key"]).To(Equal("val"))
}

func TestConfigMap(t *testing.T) {
	g := NewWithT(t)
	labels := map[string]string{"app": "test"}
	data := map[string]string{"KEY": "VALUE"}
	cm := ConfigMap("my-cm", "ns", labels, data)
	g.Expect(cm.Name).To(Equal("my-cm"))
	g.Expect(cm.Data["KEY"]).To(Equal("VALUE"))
}

func TestServiceClusterIP(t *testing.T) {
	g := NewWithT(t)
	labels := map[string]string{"app": "test"}
	selector := map[string]string{"app": "test"}
	ports := []corev1.ServicePort{ServicePort("http", 8080, 8080)}
	svc := ServiceClusterIP("my-svc", "ns", labels, selector, ports)
	g.Expect(svc.Name).To(Equal("my-svc"))
	g.Expect(svc.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
	g.Expect(svc.Spec.Ports).To(HaveLen(1))
	g.Expect(svc.Spec.Ports[0].Port).To(Equal(int32(8080)))
}

func TestPVC_DefaultValues(t *testing.T) {
	g := NewWithT(t)
	labels := map[string]string{"app": "test"}
	pvc := PVC("my-pvc", "ns", labels, "5Gi")
	g.Expect(pvc.Name).To(Equal("my-pvc"))
	g.Expect(pvc.Spec.AccessModes).To(ContainElement(corev1.ReadWriteOnce))
	g.Expect(pvc.Spec.Resources.Requests[corev1.ResourceStorage]).To(Equal(resource.MustParse("5Gi")))
}

func TestStandardServiceAccount(t *testing.T) {
	g := NewWithT(t)
	labels := map[string]string{"app": "test"}
	sa := StandardServiceAccount("my-sa", "ns", labels)
	g.Expect(sa.Name).To(Equal("my-sa"))
	g.Expect(sa.Namespace).To(Equal("ns"))
	g.Expect(sa.Labels["app"]).To(Equal("test"))
}

func TestPostgresHostFQDN(t *testing.T) {
	g := NewWithT(t)
	g.Expect(PostgresHostFQDN("my-pg", testHelpersNamespace)).To(Equal("my-pg.swim-ns.svc.cluster.local"))
}

func TestPostgresJDBCURL(t *testing.T) {
	g := NewWithT(t)
	g.Expect(PostgresJDBCURL("host", "5432", "mydb")).To(Equal("jdbc:postgresql://host:5432/mydb"))
}

func TestPostgresR2DBCURL(t *testing.T) {
	g := NewWithT(t)
	g.Expect(PostgresR2DBCURL("host", "5432", "mydb")).To(Equal("r2dbc:postgresql://host:5432/mydb"))
}

func TestBoolPtr(t *testing.T) {
	g := NewWithT(t)
	g.Expect(*BoolPtr(true)).To(BeTrue())
	g.Expect(*BoolPtr(false)).To(BeFalse())
}

func TestDefaultArtemisName_UsesSpecWhenProvided(t *testing.T) {
	g := NewWithT(t)
	g.Expect(DefaultArtemisName("custom-artemis", "my-cr")).To(Equal("custom-artemis"))
}

func TestDefaultArtemisName_DerivesFromCRNameWhenEmpty(t *testing.T) {
	g := NewWithT(t)
	name := DefaultArtemisName("", "my-cr")
	g.Expect(name).To(ContainSubstring("my-cr"))
}

func TestMustParseResource(t *testing.T) {
	g := NewWithT(t)
	q := MustParseResource("512Mi")
	g.Expect(q).To(Equal(resource.MustParse("512Mi")))
}

func minimalLabels() map[string]string {
	return map[string]string{"app": "test"}
}

func minimalMongoParams(name, namespace string) MongoParams {
	return MongoParams{
		Name:               name,
		Namespace:          namespace,
		Labels:             minimalLabels(),
		CredentialsSecret:  name + "-mongodb-credentials",
		DataPVCName:        name + "-mongodb-data",
		ServiceAccountName: name,
	}
}

func minimalPostgresParams(name, namespace string) PostgresParams {
	return PostgresParams{
		Name:               name,
		Namespace:          namespace,
		Labels:             minimalLabels(),
		SecretName:         name + "-postgres-secret",
		PVCName:            name + "-postgres-pvc",
		ServiceAccountName: name,
	}
}

// ---------------------------------------------------------------------------
// MongoDB builders
// ---------------------------------------------------------------------------

func TestBuildMongoSecret_DefaultCredentials(t *testing.T) {
	g := NewWithT(t)
	s := BuildMongoSecret(minimalMongoParams(testHelpersCRShortName, testHelpersNamespace))
	g.Expect(s.Name).To(Equal("swim-cr-mongodb-credentials"))
	g.Expect(s.StringData[testHelpersSecretKeyDatabaseName]).To(Equal(testHelpersSwimDnotam))
	g.Expect(s.StringData[testHelpersSecretKeyDatabaseUser]).To(Equal("swim"))
	g.Expect(s.StringData[testHelpersSecretKeyDatabasePassword]).To(Equal("swim123"))
	g.Expect(s.StringData["database-admin-password"]).To(Equal("admin"))
}

func TestBuildMongoSecret_CustomCredentials(t *testing.T) {
	g := NewWithT(t)
	p := minimalMongoParams(testHelpersCRShortName, testHelpersNamespace)
	p.Database = testHelpersCustomDB
	p.User = "admin"
	p.Password = "s3cret"
	s := BuildMongoSecret(p)
	g.Expect(s.StringData[testHelpersSecretKeyDatabaseName]).To(Equal(testHelpersCustomDB))
	g.Expect(s.StringData[testHelpersSecretKeyDatabaseUser]).To(Equal("admin"))
	g.Expect(s.StringData[testHelpersSecretKeyDatabasePassword]).To(Equal("s3cret"))
}

func TestBuildMongoPVC_DefaultSize(t *testing.T) {
	g := NewWithT(t)
	pvc := BuildMongoPVC(minimalMongoParams(testHelpersCRShortName, testHelpersNamespace))
	g.Expect(pvc.Spec.Resources.Requests[corev1.ResourceStorage]).To(Equal(resource.MustParse("1Gi")))
}

func TestBuildMongoPVC_CustomSize(t *testing.T) {
	g := NewWithT(t)
	p := minimalMongoParams(testHelpersCRShortName, testHelpersNamespace)
	p.StorageSize = "10Gi"
	pvc := BuildMongoPVC(p)
	g.Expect(pvc.Spec.Resources.Requests[corev1.ResourceStorage]).To(Equal(resource.MustParse("10Gi")))
}

func TestBuildMongoService_Port(t *testing.T) {
	g := NewWithT(t)
	svc := BuildMongoService(testHelpersCRShortName+"-mongodb", testHelpersNamespace, minimalLabels())
	g.Expect(svc.Spec.Ports[0].Port).To(Equal(int32(27017)))
	g.Expect(svc.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
}

func TestBuildMongoDeployment_Defaults(t *testing.T) {
	g := NewWithT(t)
	d := BuildMongoDeployment(minimalMongoParams(testHelpersCRShortName, testHelpersNamespace), "abc123")
	g.Expect(d.Name).To(Equal(testHelpersCRShortName))
	g.Expect(*d.Spec.Replicas).To(Equal(int32(1)))
	g.Expect(d.Spec.Strategy.Type).To(Equal(appsv1.RecreateDeploymentStrategyType))
	g.Expect(d.Spec.Template.Spec.Containers[0].Image).To(Equal("quay.io/mongodb/mongodb-community-server:8.0-ubi8"))
	g.Expect(d.Spec.Template.Annotations["config-hash"]).To(Equal("abc123"))
}

func TestBuildMongoDeployment_EnvVars(t *testing.T) {
	g := NewWithT(t)
	d := BuildMongoDeployment(minimalMongoParams(testHelpersCRShortName, testHelpersNamespace), "hash")
	envs := d.Spec.Template.Spec.Containers[0].Env
	envNames := make([]string, len(envs))
	for i, e := range envs {
		envNames[i] = e.Name
	}
	g.Expect(envNames).To(ContainElements("TZ", "MONGO_INITDB_DATABASE", "MONGO_INITDB_ROOT_USERNAME", "MONGO_INITDB_ROOT_PASSWORD"))
}

func TestBuildMongoDeployment_HealthProbes(t *testing.T) {
	g := NewWithT(t)
	d := BuildMongoDeployment(minimalMongoParams(testHelpersCRShortName, testHelpersNamespace), "hash")
	c := d.Spec.Template.Spec.Containers[0]
	g.Expect(c.LivenessProbe).NotTo(BeNil())
	g.Expect(c.ReadinessProbe).NotTo(BeNil())
	g.Expect(c.LivenessProbe.Exec.Command).To(ContainElement("mongosh"))
}

// ---------------------------------------------------------------------------
// PostgreSQL builders
// ---------------------------------------------------------------------------

func TestBuildPostgresSecret_DefaultCredentials(t *testing.T) {
	g := NewWithT(t)
	s := BuildPostgresSecret(minimalPostgresParams(testHelpersCRShortName, testHelpersNamespace))
	g.Expect(s.StringData["database-host"]).To(Equal("swim-cr.swim-ns.svc.cluster.local"))
	g.Expect(s.StringData[testHelpersSecretKeyDatabaseName]).To(Equal(testHelpersSwimDnotam))
	g.Expect(s.StringData[testHelpersSecretKeyDatabaseUser]).To(Equal(testHelpersSwimProviderCred))
	g.Expect(s.StringData[testHelpersSecretKeyDatabasePassword]).To(Equal(testHelpersSwimProviderCred))
}

func TestBuildPostgresSecret_CustomCredentials(t *testing.T) {
	g := NewWithT(t)
	p := minimalPostgresParams(testHelpersCRShortName, testHelpersNamespace)
	p.Database = testHelpersCustomDB
	p.User = "admin"
	p.Password = "s3cret"
	s := BuildPostgresSecret(p)
	g.Expect(s.StringData[testHelpersSecretKeyDatabaseName]).To(Equal(testHelpersCustomDB))
	g.Expect(s.StringData[testHelpersSecretKeyDatabaseUser]).To(Equal("admin"))
	g.Expect(s.StringData[testHelpersSecretKeyDatabasePassword]).To(Equal("s3cret"))
}

func TestBuildPostgresPVC_DefaultSize(t *testing.T) {
	g := NewWithT(t)
	pvc := BuildPostgresPVC(minimalPostgresParams(testHelpersCRShortName, testHelpersNamespace))
	g.Expect(pvc.Spec.Resources.Requests[corev1.ResourceStorage]).To(Equal(resource.MustParse("5Gi")))
}

func TestBuildPostgresService_Port(t *testing.T) {
	g := NewWithT(t)
	svc := BuildPostgresService(testHelpersCRShortName+"-postgres", testHelpersNamespace, minimalLabels())
	g.Expect(svc.Spec.Ports[0].Port).To(Equal(int32(5432)))
	g.Expect(svc.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
}

func TestBuildPostgresServiceHeadless_ClusterIPNone(t *testing.T) {
	g := NewWithT(t)
	svc := BuildPostgresServiceHeadless(testHelpersCRShortName+"-postgres", testHelpersNamespace, minimalLabels())
	g.Expect(svc.Spec.ClusterIP).To(Equal(corev1.ClusterIPNone))
}

func TestBuildPostgresStatefulSet_Defaults(t *testing.T) {
	g := NewWithT(t)
	ss := BuildPostgresStatefulSet(minimalPostgresParams(testHelpersCRShortName, testHelpersNamespace))
	g.Expect(*ss.Spec.Replicas).To(Equal(int32(1)))
	g.Expect(ss.Spec.Template.Spec.Containers[0].Image).To(Equal("registry.redhat.io/rhel9/postgresql-16:latest"))
}

func TestBuildPostgresStatefulSet_CustomImage(t *testing.T) {
	g := NewWithT(t)
	p := minimalPostgresParams(testHelpersCRShortName, testHelpersNamespace)
	p.Image = "my-registry/postgres:15"
	ss := BuildPostgresStatefulSet(p)
	g.Expect(ss.Spec.Template.Spec.Containers[0].Image).To(Equal("my-registry/postgres:15"))
}

func TestBuildPostgresStatefulSet_EnvVars(t *testing.T) {
	g := NewWithT(t)
	ss := BuildPostgresStatefulSet(minimalPostgresParams(testHelpersCRShortName, testHelpersNamespace))
	envs := ss.Spec.Template.Spec.Containers[0].Env
	envNames := make([]string, len(envs))
	for i, e := range envs {
		envNames[i] = e.Name
	}
	g.Expect(envNames).To(ContainElements("TZ", "POSTGRESQL_USER", "POSTGRESQL_PASSWORD", "POSTGRESQL_DATABASE",
		"POSTGRESQL_MAX_CONNECTIONS", "POSTGRESQL_SHARED_BUFFERS"))
}

func TestBuildPostgresStatefulSet_HealthProbes(t *testing.T) {
	g := NewWithT(t)
	ss := BuildPostgresStatefulSet(minimalPostgresParams(testHelpersCRShortName, testHelpersNamespace))
	c := ss.Spec.Template.Spec.Containers[0]
	g.Expect(c.LivenessProbe.Exec.Command).To(ContainElement("pg_isready"))
	g.Expect(c.ReadinessProbe.Exec.Command).To(ContainElement("pg_isready"))
}

func TestBuildUpstreamPostgresStatefulSet_DefaultImage(t *testing.T) {
	g := NewWithT(t)
	ss := BuildUpstreamPostgresStatefulSet(minimalPostgresParams(testHelpersCRShortName, testHelpersNamespace))
	g.Expect(ss.Spec.Template.Spec.Containers[0].Image).To(Equal("docker.io/postgres:16"))
}

func TestBuildUpstreamPostgresStatefulSet_UsesDirectEnvVars(t *testing.T) {
	g := NewWithT(t)
	ss := BuildUpstreamPostgresStatefulSet(minimalPostgresParams(testHelpersCRShortName, testHelpersNamespace))
	envs := ss.Spec.Template.Spec.Containers[0].Env
	envNames := make([]string, len(envs))
	for i, e := range envs {
		envNames[i] = e.Name
	}
	g.Expect(envNames).To(ContainElements("POSTGRES_DB", "POSTGRES_USER", "POSTGRES_PASSWORD", "PGDATA"))
}

// ---------------------------------------------------------------------------
// Provider app deployment
// ---------------------------------------------------------------------------

func TestBuildProviderAppDeployment_Structure(t *testing.T) {
	g := NewWithT(t)
	d := BuildProviderAppDeployment(ProviderAppDeploymentParams{
		Name:                  testHelpersMyProvider,
		Namespace:             testHelpersNamespace,
		Labels:                minimalLabels(),
		Image:                 "quay.io/masales/swim-dnotam-provider:latest",
		Replicas:              2,
		ConfigHash:            "deadbeef",
		ServerTLSSecretName:   "my-provider-server-tls",
		CABundleConfigMapName: "my-provider-ca-bundle",
	})
	g.Expect(d.Name).To(Equal(testHelpersMyProvider))
	g.Expect(*d.Spec.Replicas).To(Equal(int32(2)))
	g.Expect(d.Spec.Template.Annotations["config-hash"]).To(Equal("deadbeef"))
}

func TestBuildProviderAppDeployment_Ports(t *testing.T) {
	g := NewWithT(t)
	d := BuildProviderAppDeployment(ProviderAppDeploymentParams{
		Name:      testHelpersMyProvider,
		Namespace: testHelpersNamespace,
		Labels:    minimalLabels(),
		Image:     testHelpersTestLatestImage,
		Replicas:  1,
	})
	ports := d.Spec.Template.Spec.Containers[0].Ports
	portNames := make([]string, len(ports))
	for i, p := range ports {
		portNames[i] = p.Name
	}
	g.Expect(portNames).To(ContainElements("http", "https", "management", "internal"))
	g.Expect(ports).To(HaveLen(4))
}

func TestBuildProviderAppDeployment_EnvFrom(t *testing.T) {
	g := NewWithT(t)
	d := BuildProviderAppDeployment(ProviderAppDeploymentParams{
		Name:      testHelpersMyProvider,
		Namespace: testHelpersNamespace,
		Labels:    minimalLabels(),
		Image:     testHelpersTestLatestImage,
		Replicas:  1,
	})
	envFrom := d.Spec.Template.Spec.Containers[0].EnvFrom
	g.Expect(envFrom).To(HaveLen(3))
	g.Expect(envFrom[0].ConfigMapRef.Name).To(Equal("my-provider-config"))
	g.Expect(envFrom[1].SecretRef.Name).To(Equal("my-provider-secret"))
	g.Expect(envFrom[2].SecretRef.Name).To(Equal("my-provider-oidc-secret"))
}

func TestBuildProviderAppDeployment_DirectEnvVars(t *testing.T) {
	g := NewWithT(t)
	d := BuildProviderAppDeployment(ProviderAppDeploymentParams{
		Name:      testHelpersMyProvider,
		Namespace: testHelpersNamespace,
		Labels:    minimalLabels(),
		Image:     testHelpersTestLatestImage,
		Replicas:  1,
	})
	envs := d.Spec.Template.Spec.Containers[0].Env
	g.Expect(envs).To(HaveLen(2))
	g.Expect(envs[0].Name).To(Equal("TZ"))
	g.Expect(envs[0].Value).To(Equal("UTC"))
	g.Expect(envs[1].Name).To(Equal("KUBERNETES_NAMESPACE"))
	g.Expect(envs[1].ValueFrom.FieldRef.FieldPath).To(Equal("metadata.namespace"))
}

func TestBuildProviderAppDeployment_HealthProbes(t *testing.T) {
	g := NewWithT(t)
	d := BuildProviderAppDeployment(ProviderAppDeploymentParams{
		Name:      testHelpersMyProvider,
		Namespace: testHelpersNamespace,
		Labels:    minimalLabels(),
		Image:     testHelpersTestLatestImage,
		Replicas:  1,
	})
	c := d.Spec.Template.Spec.Containers[0]
	g.Expect(c.LivenessProbe.HTTPGet.Path).To(Equal(testHelpersQuarkusHealthLive))
	g.Expect(c.ReadinessProbe.HTTPGet.Path).To(Equal("/q/health/ready"))
}

func TestBuildProviderAppDeployment_TLSVolumes(t *testing.T) {
	g := NewWithT(t)
	d := BuildProviderAppDeployment(ProviderAppDeploymentParams{
		Name:                  testHelpersMyProvider,
		Namespace:             testHelpersNamespace,
		Labels:                minimalLabels(),
		Image:                 testHelpersTestLatestImage,
		Replicas:              1,
		ServerTLSSecretName:   "my-tls-secret",
		CABundleConfigMapName: "my-ca-bundle",
	})
	volumes := d.Spec.Template.Spec.Volumes
	g.Expect(volumes).To(HaveLen(2))
	g.Expect(volumes[0].Name).To(Equal("server-cert"))
	g.Expect(volumes[0].Secret.SecretName).To(Equal("my-tls-secret"))
	g.Expect(volumes[1].Name).To(Equal("ca-cert"))
	g.Expect(volumes[1].ConfigMap.Name).To(Equal("my-ca-bundle"))
}

// ---------------------------------------------------------------------------
// Consumer ConfigMap data — DNOTAM
// ---------------------------------------------------------------------------

func TestDnotamConsumerConfigMap_DefaultValues(t *testing.T) {
	g := NewWithT(t)
	data := BuildDnotamConsumerConfigMapData(DnotamConsumerConfigMapParams{
		CRName:    testHelpersCRShortName,
		Namespace: testHelpersNamespace,
	})
	g.Expect(data["MONGODB_HOST"]).To(Equal("swim-cr-mongodb.swim-ns.svc.cluster.local"))
	g.Expect(data["MONGODB_PORT"]).To(Equal("27017"))
	g.Expect(data["MONGODB_DATABASE"]).To(Equal(testHelpersSwimDnotam))
	g.Expect(data["SWIM_VALIDATION_ENABLED"]).To(Equal("true"))
	g.Expect(data["SWIM_VALIDATION_FAIL_ON_NULLBODY"]).To(Equal("false"))
	g.Expect(data["DNOTAM_DELETE_AND_RECREATE"]).To(Equal("true"))
	g.Expect(data["OTEL_ENABLED"]).To(Equal("false"))
	g.Expect(data["OTEL_SDK_DISABLED"]).To(Equal("true"))
	g.Expect(data["PROMETHEUS_ENABLED"]).To(Equal("false"))
}

func TestDnotamConsumerConfigMap_ManagedKafka(t *testing.T) {
	g := NewWithT(t)
	data := BuildDnotamConsumerConfigMapData(DnotamConsumerConfigMapParams{
		CRName:    testHelpersCRShortName,
		Namespace: testHelpersNamespace,
	})
	g.Expect(data["KAFKA_BOOTSTRAP_SERVERS"]).To(Equal("kafka-kafka-bootstrap.swim-ns.svc.cluster.local:9092"))
}

func TestDnotamConsumerConfigMap_ExternalKafka(t *testing.T) {
	g := NewWithT(t)
	data := BuildDnotamConsumerConfigMapData(DnotamConsumerConfigMapParams{
		CRName:                 testHelpersCRShortName,
		Namespace:              testHelpersNamespace,
		KafkaDeploymentMode:    "external",
		KafkaBootstrapExternal: "kafka.company.com:9092",
	})
	g.Expect(data["KAFKA_BOOTSTRAP_SERVERS"]).To(Equal("kafka.company.com:9092"))
}

func TestDnotamConsumerConfigMap_CustomDatabase(t *testing.T) {
	g := NewWithT(t)
	data := BuildDnotamConsumerConfigMapData(DnotamConsumerConfigMapParams{
		CRName:        testHelpersCRShortName,
		Namespace:     testHelpersNamespace,
		MongoDatabase: testHelpersCustomDB,
	})
	g.Expect(data["MONGODB_DATABASE"]).To(Equal(testHelpersCustomDB))
}

func TestDnotamConsumerConfigMap_ObservabilityEnabled(t *testing.T) {
	g := NewWithT(t)
	data := BuildDnotamConsumerConfigMapData(DnotamConsumerConfigMapParams{
		CRName:               testHelpersCRShortName,
		Namespace:            testHelpersNamespace,
		OpenTelemetryEnabled: true,
		OtelEndpoint:         "http://otel:4317",
		OtelHeaders:          "key=value",
		PrometheusEnabled:    true,
	})
	g.Expect(data["OTEL_ENABLED"]).To(Equal("true"))
	g.Expect(data["OTEL_SDK_DISABLED"]).To(Equal("false"))
	g.Expect(data["OTEL_ENDPOINT"]).To(Equal("http://otel:4317"))
	g.Expect(data["OTEL_HEADERS"]).To(Equal("key=value"))
	g.Expect(data["PROMETHEUS_ENABLED"]).To(Equal("true"))
}

// ---------------------------------------------------------------------------
// Consumer ConfigMap data — ED-254
// ---------------------------------------------------------------------------

func TestEd254ConsumerConfigMap_DefaultValues(t *testing.T) {
	g := NewWithT(t)
	data := BuildEd254ConsumerConfigMapData(Ed254ConsumerConfigMapParams{
		CRName:    testHelpersCRShortName,
		Namespace: testHelpersNamespace,
	})
	g.Expect(data["MONGODB_HOST"]).To(Equal("swim-cr-mongodb.swim-ns.svc.cluster.local"))
	g.Expect(data["MONGODB_PORT"]).To(Equal("27017"))
	g.Expect(data["MONGODB_DATABASE"]).To(Equal("swim-ed254"))
	g.Expect(data["SWIM_VALIDATION_ENABLED"]).To(Equal("true"))
	g.Expect(data["OTEL_ENABLED"]).To(Equal("false"))
	g.Expect(data["OTEL_SDK_DISABLED"]).To(Equal("true"))
	g.Expect(data["PROMETHEUS_ENABLED"]).To(Equal("false"))
}

func TestEd254ConsumerConfigMap_HasNoHeartbeatWhenZero(t *testing.T) {
	g := NewWithT(t)
	data := BuildEd254ConsumerConfigMapData(Ed254ConsumerConfigMapParams{
		CRName:    testHelpersCRShortName,
		Namespace: testHelpersNamespace,
	})
	_, exists := data["HEARTBEAT_TIMEOUT_SECONDS"]
	g.Expect(exists).To(BeFalse())
}

func TestEd254ConsumerConfigMap_HeartbeatTimeoutWhenPositive(t *testing.T) {
	g := NewWithT(t)
	data := BuildEd254ConsumerConfigMapData(Ed254ConsumerConfigMapParams{
		CRName:                  testHelpersCRShortName,
		Namespace:               testHelpersNamespace,
		HeartbeatTimeoutSeconds: 60,
	})
	g.Expect(data["HEARTBEAT_TIMEOUT_SECONDS"]).To(Equal("60"))
}

func TestEd254ConsumerConfigMap_ExternalKafka(t *testing.T) {
	g := NewWithT(t)
	data := BuildEd254ConsumerConfigMapData(Ed254ConsumerConfigMapParams{
		CRName:                 testHelpersCRShortName,
		Namespace:              testHelpersNamespace,
		KafkaDeploymentMode:    "external",
		KafkaBootstrapExternal: "kafka.ext.com:9092",
	})
	g.Expect(data["KAFKA_BOOTSTRAP_SERVERS"]).To(Equal("kafka.ext.com:9092"))
}

func TestEd254ConsumerConfigMap_DiffersFromDnotam(t *testing.T) {
	g := NewWithT(t)
	dnotam := BuildDnotamConsumerConfigMapData(DnotamConsumerConfigMapParams{CRName: "cr", Namespace: "ns"})
	ed254 := BuildEd254ConsumerConfigMapData(Ed254ConsumerConfigMapParams{CRName: "cr", Namespace: "ns"})
	g.Expect(dnotam["MONGODB_DATABASE"]).To(Equal(testHelpersSwimDnotam))
	g.Expect(ed254["MONGODB_DATABASE"]).To(Equal("swim-ed254"))
	_, hasDnotamSubs := dnotam["DNOTAM_SUBSCRIPTIONS"]
	_, hasEd254Subs := ed254["ED254_SUBSCRIPTIONS"]
	g.Expect(hasDnotamSubs).To(BeTrue())
	g.Expect(hasEd254Subs).To(BeTrue())
}

// ---------------------------------------------------------------------------
// Consumer client deployment
// ---------------------------------------------------------------------------

func TestConsumerClientDeployment_EnvVars(t *testing.T) {
	g := NewWithT(t)
	d := BuildSwimConsumerClientDeployment(SwimConsumerClientDeploymentParams{
		Name:                   testHelpersCRShortName,
		Namespace:              testHelpersNamespace,
		Labels:                 minimalLabels(),
		Image:                  "quay.io/masales/swim-dnotam-consumer:latest",
		Replicas:               1,
		ConfigMapName:          "swim-cr-config",
		ProvidersSecretName:    "swim-cr-providers",
		KafkaCredentialsSecret: "swim-cr-kafka-credentials",
		MongoCredentialsSecret: "swim-cr-mongodb-credentials",
		KeystorePasswordSecret: "swim-cr-keystore-password",
		MTLSSecretName:         "swim-cr-mtls",
	})
	envs := d.Spec.Template.Spec.Containers[0].Env
	envNames := make([]string, len(envs))
	for i, e := range envs {
		envNames[i] = e.Name
	}
	g.Expect(envNames).To(ContainElements(
		"TZ", "MONGODB_USERNAME", "MONGODB_PASSWORD", "MONGODB_URI",
		"SWIM_TRUSTSTORE_PATH", "SWIM_TRUSTSTORE_PASSWORD",
		"SWIM_CLIENT_KEYSTORE_PATH", "SWIM_CLIENT_KEYSTORE_PASSWORD",
		"POD_NAME", "POD_NAMESPACE",
	))
}

func TestConsumerClientDeployment_InitContainer(t *testing.T) {
	g := NewWithT(t)
	d := BuildSwimConsumerClientDeployment(SwimConsumerClientDeploymentParams{
		Name:      testHelpersCRShortName,
		Namespace: testHelpersNamespace,
		Labels:    minimalLabels(),
		Image:     testHelpersTestLatestImage,
		Replicas:  1,
	})
	g.Expect(d.Spec.Template.Spec.InitContainers).To(HaveLen(1))
	g.Expect(d.Spec.Template.Spec.InitContainers[0].Name).To(Equal("validate-secrets"))
}

func TestConsumerClientDeployment_EnvFromSources(t *testing.T) {
	g := NewWithT(t)
	d := BuildSwimConsumerClientDeployment(SwimConsumerClientDeploymentParams{
		Name:                   testHelpersCRShortName,
		Namespace:              testHelpersNamespace,
		Labels:                 minimalLabels(),
		Image:                  testHelpersTestLatestImage,
		Replicas:               1,
		ConfigMapName:          "my-config",
		ProvidersSecretName:    "my-providers",
		KafkaCredentialsSecret: "my-kafka",
	})
	envFrom := d.Spec.Template.Spec.Containers[0].EnvFrom
	g.Expect(envFrom).To(HaveLen(3))
	g.Expect(envFrom[0].ConfigMapRef.Name).To(Equal("my-config"))
	g.Expect(envFrom[1].SecretRef.Name).To(Equal("my-providers"))
	g.Expect(envFrom[2].SecretRef.Name).To(Equal("my-kafka"))
}

func init() {
	// ensure the init is a no-op
	_ = metav1.ObjectMeta{}
}
