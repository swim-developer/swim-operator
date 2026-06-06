package v1alpha1

import (
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SwimDigitalNotamProviderSpec struct {
	commonapi.SwimDigitalNotamProviderBaseSpec `json:",inline"`
	// +optional
	Provider ProviderAppSpec `json:"provider,omitempty"`
}

type ProviderAppSpec struct {
	commonapi.ProviderAppBaseSpec `json:",inline"`
	// +optional
	Routes ProviderRoutesSpec `json:"routes,omitempty"`
}

type ProviderRoutesSpec struct {
	// +kubebuilder:default=true
	HTTPSEdge bool `json:"httpsEdge,omitempty"`
	// +kubebuilder:default=true
	HTTPSPassthrough bool `json:"httpsPassthrough,omitempty"`
	// +optional
	HTTPSEdgeHost string `json:"httpsEdgeHost,omitempty"`
	// +optional
	HTTPSPassthroughHost string `json:"httpsPassthroughHost,omitempty"`
}

type SwimDigitalNotamProviderStatus = commonapi.SwimStatus

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type SwimDigitalNotamProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SwimDigitalNotamProviderSpec   `json:"spec,omitempty"`
	Status SwimDigitalNotamProviderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type SwimDigitalNotamProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SwimDigitalNotamProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SwimDigitalNotamProvider{}, &SwimDigitalNotamProviderList{})
}
