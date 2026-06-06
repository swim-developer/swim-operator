package v1alpha1

type FficeSubscriptionSpec struct {
	// +optional
	Provider string `json:"provider,omitempty"`
	// +optional
	Topic string `json:"topic,omitempty"`
	// +optional
	Description string `json:"description,omitempty"`
}
