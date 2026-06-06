package v1alpha1

import (
	commonapi "github.com/swim-developer/swim-operator-common/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SwimEd254ProviderSpec struct {
	commonapi.SwimEd254ProviderBaseSpec `json:",inline"`
	// +optional
	Provider Ed254ProviderAppSpec `json:"provider,omitempty"`
}

type Ed254ProviderAppSpec struct {
	commonapi.Ed254ProviderAppBaseSpec `json:",inline"`
	// +optional
	Routes ProviderRoutesSpec `json:"routes,omitempty"`
}

type SwimEd254ProviderStatus = commonapi.SwimStatus

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type SwimEd254Provider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SwimEd254ProviderSpec   `json:"spec,omitempty"`
	Status SwimEd254ProviderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type SwimEd254ProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SwimEd254Provider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SwimEd254Provider{}, &SwimEd254ProviderList{})
}
