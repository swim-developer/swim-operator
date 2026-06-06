package v1alpha1

import (
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SwimDnotamConsumerValidatorSpec = commonapi.SwimDnotamConsumerValidatorSpec

type SwimDnotamConsumerValidatorStatus = commonapi.SwimStatus

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=sdcv

type SwimDnotamConsumerValidator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SwimDnotamConsumerValidatorSpec   `json:"spec,omitempty"`
	Status SwimDnotamConsumerValidatorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type SwimDnotamConsumerValidatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SwimDnotamConsumerValidator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SwimDnotamConsumerValidator{}, &SwimDnotamConsumerValidatorList{})
}
