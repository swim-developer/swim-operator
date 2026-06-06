package cv

import (
	"fmt"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/labels"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func MariaDBServiceName(crName string) string {
	return fmt.Sprintf("%s-mariadb", crName)
}

func cvGetMariaDBSecretName(p CVBuildParams) string {
	if p.Spec.MariaDB.ExistingSecret != "" {
		return p.Spec.MariaDB.ExistingSecret
	}
	return fmt.Sprintf("%s-mariadb-credentials", p.CRName)
}

func BuildCVServiceAccount(p CVBuildParams, managedBy string) *corev1.ServiceAccount {
	return resources.StandardServiceAccount(p.CRName, p.Namespace, labels.StandardLabels(p.CRName, constants.ConsumerValidatorApp, p.CRName, managedBy))
}

func BuildCVRole(p CVBuildParams, managedBy string) *rbacv1.Role {
	return resources.BuildSecretsConfigMapsRole(p.CRName, p.Namespace, labels.StandardLabels(p.CRName, constants.ConsumerValidatorApp, p.CRName, managedBy))
}

func BuildCVRoleBinding(p CVBuildParams, managedBy string) *rbacv1.RoleBinding {
	return resources.BuildRoleBinding(p.CRName, p.Namespace, p.CRName, fmt.Sprintf("%s-role", p.CRName), labels.StandardLabels(p.CRName, constants.ConsumerValidatorApp, p.CRName, managedBy))
}

func BuildCVAmqpSecret(p CVBuildParams, managedBy string) *corev1.Secret {
	lbl := labels.StandardLabels(p.CRName, constants.ConsumerValidatorApp, p.CRName, managedBy)
	username := resources.StrDefault(p.Spec.AppConfig.Amqp.Username, "admin")
	password := resources.StrDefault(p.Spec.AppConfig.Amqp.Password, "admin")
	return resources.SecretStringData(fmt.Sprintf("%s-amqp-credentials", p.CRName), p.Namespace, lbl, map[string]string{
		"AMQP_BROKER_USERNAME": username,
		"AMQP_BROKER_PASSWORD": password,
	})
}

func BuildCVMariaDBSecret(p CVBuildParams, managedBy string) *corev1.Secret {
	svc := MariaDBServiceName(p.CRName)
	return resources.BuildMariaDBSecret(resources.MariaDBParams{
		Name:       svc,
		Namespace:  p.Namespace,
		Labels:     labels.StandardLabels(svc, "mariadb", p.CRName, managedBy),
		Database:   p.Spec.MariaDB.Database,
		Username:   p.Spec.MariaDB.Username,
		Password:   p.Spec.MariaDB.Password,
		SecretName: cvGetMariaDBSecretName(p),
	})
}

func BuildCVMariaDBStatefulSet(p CVBuildParams, managedBy string) *appsv1.StatefulSet {
	svc := MariaDBServiceName(p.CRName)
	return resources.BuildMariaDBStatefulSet(resources.MariaDBParams{
		Name:               svc,
		Namespace:          p.Namespace,
		Labels:             labels.StandardLabels(svc, "mariadb", p.CRName, managedBy),
		Database:           resources.StrDefault(p.Spec.MariaDB.Database, p.DefaultDatabase),
		Username:           p.Spec.MariaDB.Username,
		Password:           p.Spec.MariaDB.Password,
		ServiceAccountName: p.CRName,
		SecretName:         cvGetMariaDBSecretName(p),
	})
}

func BuildCVMariaDBService(p CVBuildParams, managedBy string) *corev1.Service {
	svc := MariaDBServiceName(p.CRName)
	return resources.BuildMariaDBService(svc, p.Namespace, labels.StandardLabels(svc, "mariadb", p.CRName, managedBy))
}