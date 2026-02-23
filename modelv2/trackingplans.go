package modelv2

type TrackingPlan struct {
	Version int `json:"version"`
}

type DGSourceTrackingPlanConfig struct {
	Global   *DGSourceTrackingPlanOptions `json:"global"`
	Track    *DGSourceTrackingPlanOptions `json:"track"`
	Identify *DGSourceTrackingPlanOptions `json:"identify"`
	Group    *DGSourceTrackingPlanOptions `json:"group"`
}

type DGSourceTrackingPlan struct {
	TrackingPlanID string                     `json:"trackingPlanId"`
	Version        int                        `json:"version"`
	Config         DGSourceTrackingPlanConfig `json:"config"`
	Deleted        bool                       `json:"deleted"`
	// TODO: consider adding the following fields
	// CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt time.Time `json:"updatedAt"`
}

type DGSourceTrackingPlanOptions struct {
	AllowUnplannedEvents any            `json:"allowUnplannedEvents"`
	SendViolatedEventsTo string         `json:"sendViolatedEventsTo"`
	AJVOptions           map[string]any `json:"ajvOptions"`
}
