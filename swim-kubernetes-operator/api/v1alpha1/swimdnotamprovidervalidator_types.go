package v1alpha1

import (
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SwimDnotamProviderValidatorSpec struct {
	commonapi.ProviderValidatorBaseSpec `json:",inline"`
	// +optional
	CertManager commonapi.CertManagerSpec `json:"certManager,omitempty"`
	// +optional
	Ingress ProviderValidatorIngressSpec `json:"ingress,omitempty"`
}

type ProviderValidatorIngressSpec struct {
	commonapi.IngressSpec `json:",inline"`
}

type SwimDnotamProviderValidatorStatus struct {
	commonapi.SwimStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=sdpv

type SwimDnotamProviderValidator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SwimDnotamProviderValidatorSpec   `json:"spec,omitempty"`
	Status SwimDnotamProviderValidatorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type SwimDnotamProviderValidatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SwimDnotamProviderValidator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SwimDnotamProviderValidator{}, &SwimDnotamProviderValidatorList{})
}
