package v1alpha1

import (
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProviderValidatorSpec struct {
	commonapi.ProviderValidatorBaseSpec `json:",inline"`
	// +optional
	Route ProviderValidatorRouteSpec `json:"route,omitempty"`
}

type ProviderValidatorRouteSpec struct {
	// +optional
	Host string `json:"host,omitempty"`
}

type SwimDnotamProviderValidatorStatus = commonapi.SwimStatus

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=sdpv
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='Available')].status"

type SwimDnotamProviderValidator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderValidatorSpec             `json:"spec,omitempty"`
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
