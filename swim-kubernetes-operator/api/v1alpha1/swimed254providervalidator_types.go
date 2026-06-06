package v1alpha1

import (
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SwimEd254ProviderValidatorSpec struct {
	commonapi.ProviderValidatorBaseSpec `json:",inline"`
	// +optional
	CertManager commonapi.CertManagerSpec `json:"certManager,omitempty"`
	// +optional
	Ingress ProviderValidatorIngressSpec `json:"ingress,omitempty"`
}

type SwimEd254ProviderValidatorStatus struct {
	commonapi.SwimStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=sepv

type SwimEd254ProviderValidator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SwimEd254ProviderValidatorSpec   `json:"spec,omitempty"`
	Status SwimEd254ProviderValidatorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type SwimEd254ProviderValidatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SwimEd254ProviderValidator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SwimEd254ProviderValidator{}, &SwimEd254ProviderValidatorList{})
}
