package modelv2

type Source struct {
	Name               string                `json:"name"`
	WriteKey           string                `json:"writeKey"`
	Enabled            bool                  `json:"enabled"`
	Deleted            bool                  `json:"deleted"`
	Config             map[string]any        `json:"config"`
	Transient          bool                  `json:"transient"`
	DefinitionName     string                `json:"sourceDefinitionName"`
	SecretVersion      int                   `json:"secretVersion"`
	AccountID          string                `json:"accountId"`
	TrackingPlanConfig *DGSourceTrackingPlan `json:"dgSourceTrackingPlanConfig"`
	// TODO: consider adding the following fields
	// CreatedBy string `json:"createdBy"`
	// CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt time.Time `json:"updatedAt"`
	// LiveEventsConfig ...
}

type SourceDefinition struct {
	Name        string         `json:"name"`
	DisplayName string         `json:"displayName"`
	Category    string         `json:"category"`
	Options     map[string]any `json:"options"`
	Config      map[string]any `json:"config"`
	// TODO: consider adding the following fields
	// CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt time.Time `json:"updatedAt"`
}
