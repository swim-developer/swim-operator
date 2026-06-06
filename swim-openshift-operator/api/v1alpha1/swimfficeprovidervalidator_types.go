package v1alpha1

import (
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SwimFficeProviderValidatorStatus = commonapi.SwimStatus

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=sfpv

type SwimFficeProviderValidator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderValidatorSpec             `json:"spec,omitempty"`
	Status SwimFficeProviderValidatorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type SwimFficeProviderValidatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SwimFficeProviderValidator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SwimFficeProviderValidator{}, &SwimFficeProviderValidatorList{})
}
