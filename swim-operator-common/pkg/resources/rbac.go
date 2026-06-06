package resources

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildServiceAccount(name, namespace string, labels map[string]string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: labels},
	}
}

func BuildLeaseRole(name, namespace string, labels map[string]string) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-role", name), Namespace: namespace, Labels: labels},
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{"coordination.k8s.io"},
			Resources: []string{"leases"},
			Verbs:     []string{"get", "list", "watch", "create", "update", "patch"},
		}},
	}
}

func BuildSecretsConfigMapsRole(name, namespace string, labels map[string]string) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-role", name), Namespace: namespace, Labels: labels},
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{""},
			Resources: []string{"secrets", "configmaps"},
			Verbs:     []string{"get", "list", "watch"},
		}},
	}
}

func BuildRoleBinding(name, namespace, serviceAccountName, roleName string, labels map[string]string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-rolebinding", name), Namespace: namespace, Labels: labels},
		Subjects: []rbacv1.Subject{{
			Kind: "ServiceAccount", Name: serviceAccountName, Namespace: namespace,
		}},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     roleName,
		},
	}
}
