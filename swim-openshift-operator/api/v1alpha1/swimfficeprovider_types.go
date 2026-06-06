package v1alpha1

import (
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SwimFficeProviderSpec struct {
	commonapi.SwimFficeProviderBaseSpec `json:",inline"`
	// +optional
	Provider FficeProviderAppSpec `json:"provider,omitempty"`
}

type FficeProviderAppSpec struct {
	commonapi.FficeProviderAppBaseSpec `json:",inline"`
	// +optional
	Routes ProviderRoutesSpec `json:"routes,omitempty"`
}

type SwimFficeProviderStatus = commonapi.SwimStatus

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type SwimFficeProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SwimFficeProviderSpec   `json:"spec,omitempty"`
	Status SwimFficeProviderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type SwimFficeProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SwimFficeProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SwimFficeProvider{}, &SwimFficeProviderList{})
}
