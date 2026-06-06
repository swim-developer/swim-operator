package pv

import (
	"fmt"

	"github.com/swim-developer/swim-operator-common/pkg/constants"
	"github.com/swim-developer/swim-operator-common/pkg/labels"
	"github.com/swim-developer/swim-operator-common/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func StandardLabels(crName, managedByValue string) map[string]string {
	return labels.StandardLabels(crName, constants.ProviderValidatorApp, crName, managedByValue)
}

func GetMariaDBSecretName(p PVBuildParams) string {
	if p.Spec.MariaDB.ExistingSecret != "" {
		return p.Spec.MariaDB.ExistingSecret
	}
	return fmt.Sprintf("%s-mariadb-credentials", p.CRName)
}

func BuildPVServiceAccount(p PVBuildParams, managedBy string) *corev1.ServiceAccount {
	return resources.StandardServiceAccount(p.CRName, p.Namespace, StandardLabels(p.CRName, managedBy))
}

func BuildPVRole(p PVBuildParams, managedBy string) *rbacv1.Role {
	return resources.BuildSecretsConfigMapsRole(p.CRName, p.Namespace, StandardLabels(p.CRName, managedBy))
}

func BuildPVRoleBinding(p PVBuildParams, managedBy string) *rbacv1.RoleBinding {
	return resources.BuildRoleBinding(p.CRName, p.Namespace, p.CRName, fmt.Sprintf("%s-role", p.CRName), StandardLabels(p.CRName, managedBy))
}

func BuildPVAppPVC(p PVBuildParams, managedBy string) *corev1.PersistentVolumeClaim {
	return resources.PVC(fmt.Sprintf("%s-data", p.CRName), p.Namespace, StandardLabels(p.CRName, managedBy), "1Gi")
}

func BuildPVMariaDBSecret(p PVBuildParams, managedBy string) *corev1.Secret {
	mdName := MariaDBServiceName(p.CRName)
	return resources.BuildMariaDBSecret(resources.MariaDBParams{
		Name:       mdName,
		Namespace:  p.Namespace,
		Labels:     StandardLabels(mdName, managedBy),
		Database:   resources.StrDefault(p.Spec.MariaDB.Database, "swim_provider_validator"),
		Username:   resources.StrDefault(p.Spec.MariaDB.Username, "swim"),
		Password:   resources.StrDefault(p.Spec.MariaDB.Password, "swim"),
		SecretName: fmt.Sprintf("%s-mariadb-credentials", p.CRName),
	})
}

func BuildPVMariaDBService(p PVBuildParams, managedBy string) *corev1.Service {
	mdName := MariaDBServiceName(p.CRName)
	return resources.BuildMariaDBService(mdName, p.Namespace, StandardLabels(mdName, managedBy))
}

func BuildPVMariaDBStatefulSet(p PVBuildParams, managedBy string) *appsv1.StatefulSet {
	mdName := MariaDBServiceName(p.CRName)
	return resources.BuildMariaDBStatefulSet(resources.MariaDBParams{
		Name:               mdName,
		Namespace:          p.Namespace,
		Labels:             StandardLabels(mdName, managedBy),
		Database:           resources.StrDefault(p.Spec.MariaDB.Database, "swim_provider_validator"),
		Username:           resources.StrDefault(p.Spec.MariaDB.Username, "swim"),
		Password:           resources.StrDefault(p.Spec.MariaDB.Password, "swim"),
		ServiceAccountName: p.CRName,
		SecretName:         GetMariaDBSecretName(p),
	})
}

func BuildPVKeystorePasswordSecret(p PVBuildParams, managedBy string) *corev1.Secret {
	return resources.SecretStringData(fmt.Sprintf(constants.KeystorePasswordSuffix, p.CRName), p.Namespace, StandardLabels(p.CRName, managedBy), map[string]string{"password": "changeit"})
}

func MTLSSecretName(p PVBuildParams) string {
	if !p.Spec.MTLS.Enabled {
		return ""
	}
	if p.Spec.MTLS.CertsSecretName != "" {
		return p.Spec.MTLS.CertsSecretName
	}
	if p.CertManager.Enabled {
		return fmt.Sprintf("%s-mtls", p.CRName)
	}
	return ""
}
