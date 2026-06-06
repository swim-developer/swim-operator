package v1alpha1

type Ed254SubscriptionSpec struct {
	// +optional
	Provider string `json:"provider,omitempty"`
	// +optional
	DestinationAerodrome []Ed254DestinationAerodrome `json:"destinationAerodrome,omitempty"`
	// +optional
	PointName []string `json:"pointName,omitempty"`
	// +optional
	FlightSelector []Ed254FlightSelector `json:"flightSelector,omitempty"`
	// +optional
	SupplementaryData *Ed254SupplementaryData `json:"supplementaryData"`
	// +optional
	Description string `json:"description,omitempty"`
}

type Ed254DestinationAerodrome struct {
	// +kubebuilder:validation:Required
	AerodromeDesignator string `json:"aerodromeDesignator"`
	// +optional
	AssignedArrivalRunway []string `json:"assignedArrivalRunway,omitempty"`
}

type Ed254FlightSelector struct {
	// +optional
	Arcid string `json:"arcid,omitempty"`
	// +optional
	Ades string `json:"ades,omitempty"`
	// +optional
	Adep string `json:"adep,omitempty"`
	// +optional
	Eobt string `json:"eobt,omitempty"`
	// +optional
	Eobd string `json:"eobd,omitempty"`
	// +optional
	IfplId string `json:"ifplId,omitempty"`
}

type Ed254SupplementaryData struct {
	// +optional
	Delay bool `json:"delay,omitempty"`
	// +optional
	LandingSequencePosition bool `json:"landingSequencePosition,omitempty"`
	// +optional
	AmanStrategy bool `json:"amanStrategy,omitempty"`
	// +optional
	DepartureAerodrome bool `json:"departureAerodrome,omitempty"`
	// +optional
	ProposedProcedure bool `json:"proposedProcedure,omitempty"`
}
