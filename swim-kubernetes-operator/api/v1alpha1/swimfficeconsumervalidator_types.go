package v1alpha1

import (
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SwimFficeConsumerValidatorSpec struct {
	commonapi.SwimFficeConsumerValidatorSpec `json:",inline"`
	// +optional
	Ingress ConsumerValidatorIngressSpec `json:"ingress,omitempty"`
	// +optional
	Observability commonapi.ObservabilitySpec `json:"observability,omitempty"`
}

type SwimFficeConsumerValidatorStatus struct {
	commonapi.SwimStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=sfcv

type SwimFficeConsumerValidator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SwimFficeConsumerValidatorSpec   `json:"spec,omitempty"`
	Status SwimFficeConsumerValidatorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type SwimFficeConsumerValidatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SwimFficeConsumerValidator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SwimFficeConsumerValidator{}, &SwimFficeConsumerValidatorList{})
}
