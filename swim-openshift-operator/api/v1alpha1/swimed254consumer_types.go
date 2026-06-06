package v1alpha1

import (
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SwimEd254ConsumerSpec = commonapi.SwimEd254ConsumerSpec

type SwimEd254ConsumerStatus = commonapi.SwimStatus

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type SwimEd254Consumer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SwimEd254ConsumerSpec   `json:"spec,omitempty"`
	Status SwimEd254ConsumerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type SwimEd254ConsumerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SwimEd254Consumer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SwimEd254Consumer{}, &SwimEd254ConsumerList{})
}
