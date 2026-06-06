package v1alpha1

import (
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SwimFficeConsumerSpec = commonapi.SwimFficeConsumerSpec

type SwimFficeConsumerStatus struct {
	commonapi.SwimStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type SwimFficeConsumer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SwimFficeConsumerSpec   `json:"spec,omitempty"`
	Status SwimFficeConsumerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type SwimFficeConsumerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SwimFficeConsumer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SwimFficeConsumer{}, &SwimFficeConsumerList{})
}
