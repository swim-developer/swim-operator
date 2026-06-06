package v1alpha1

import (
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SwimDigitalNotamConsumerSpec = commonapi.SwimDigitalNotamConsumerSpec

type SwimDigitalNotamConsumerStatus struct {
	commonapi.SwimStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type SwimDigitalNotamConsumer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SwimDigitalNotamConsumerSpec   `json:"spec,omitempty"`
	Status SwimDigitalNotamConsumerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type SwimDigitalNotamConsumerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SwimDigitalNotamConsumer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SwimDigitalNotamConsumer{}, &SwimDigitalNotamConsumerList{})
}
