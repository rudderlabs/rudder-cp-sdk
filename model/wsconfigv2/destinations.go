package wsconfigv2

type Destination struct {
	Name              string                 `json:"name"`
	Enabled           bool                   `json:"enabled"`
	Config            map[string]interface{} `json:"config"`
	DefinitionName    string                 `json:"destinationDefinitionName"`
	Deleted           bool                   `json:"deleted"`
	TransformationIDs []string               `json:"transformationIds"`
	RevisionID        string                 `json:"revisionId"`
	SecretVersion     int                    `json:"secretVersion"`
	Regions           []*DestinationRegion   `json:"regions"`
	// TODO: consider adding the following fields
	// CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt time.Time `json:"updatedAt"`
	// LiveEventsConfig ...
}

type DestinationRegion struct {
	Region        Region                 `json:"region"`
	Config        map[string]interface{} `json:"config"`
	RevisionID    string                 `json:"revisionId"`
	SecretVersion int                    `json:"secretVersion"`
	// TODO: consider adding the following fields
	// CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt time.Time `json:"updatedAt"`
}

type DestinationDefinition struct {
	Name        string                 `json:"name"`
	DisplayName string                 `json:"displayName"`
	Category    string                 `json:"category"`
	Options     map[string]interface{} `json:"options"`
	Config      map[string]interface{} `json:"config"`
	// TODO: consider adding the following fields
	// ResponseRules map[string]interface{} `json:"responseRules"`
	// CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt time.Time `json:"updatedAt"`
}
