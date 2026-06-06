package v1alpha1

import (
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SwimEd254ConsumerValidatorSpec = commonapi.SwimEd254ConsumerValidatorSpec

type SwimEd254ConsumerValidatorStatus = commonapi.SwimStatus

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=secv

type SwimEd254ConsumerValidator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SwimEd254ConsumerValidatorSpec   `json:"spec,omitempty"`
	Status SwimEd254ConsumerValidatorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type SwimEd254ConsumerValidatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SwimEd254ConsumerValidator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SwimEd254ConsumerValidator{}, &SwimEd254ConsumerValidatorList{})
}
