package modelv2

import "time"

// WorkspaceConfigs represents workspace configurations of one or more workspaces, as well as definitions shared by all of them.
type WorkspaceConfigs struct {
	// Workspaces is an map of workspace configurations. The key is a workspace ID.
	Workspaces map[string]*WorkspaceConfig `json:"workspaces"`
	// SourceDefinitions is a map of source definitions. The key is a source definition name.
	SourceDefinitions map[string]*SourceDefinition `json:"sourceDefinitions"`
	// DestinationDefinitions is a map of destination definitions. The key is a destination definition name.
	DestinationDefinitions map[string]*DestinationDefinition `json:"destinationDefinitions"`
}

type WorkspaceConfig struct {
	Settings                         *Settings                          `json:"settings"`
	Sources                          map[string]*Source                 `json:"sources"`
	Destinations                     map[string]*Destination            `json:"destinations"`
	Connections                      map[string]*Connection             `json:"connections"`
	Libraries                        []*Library                         `json:"libraries"`
	WHTProjects                      map[string]*WHTProject             `json:"whtProjects"`
	Accounts                         map[string]*Account                `json:"accounts"`
	Transformations                  map[string]*Transformation         `json:"transformations"`
	TrackingPlans                    map[string]*TrackingPlan           `json:"trackingPlans"`
	Resources                        map[string]*Resource               `json:"resources"`
	AudienceSources                  []*AudienceSource                  `json:"audienceSources"`
	SQLModelVersions                 map[string]*SQLModelVersion        `json:"sqlModelVersions"`
	SQLModelVersionSourceConnections []*SQLModelVersionSourceConnection `json:"sqlModelVersionSourceConnections"`
	UpdatedAt                        time.Time                          `json:"updatedAt"`
}

// UpdatedAt returns the maximum UpdatedAt value of all included workspace configs.
func (wcs *WorkspaceConfigs) UpdatedAt() time.Time {
	var updateAt time.Time
	for _, wc := range wcs.Workspaces {
		if wc != nil && wc.UpdatedAt.After(updateAt) {
			updateAt = wc.UpdatedAt
		}
	}

	return updateAt
}

// Empty returns an empty WorkspaceConfigs object.
func Empty() *WorkspaceConfigs {
	return &WorkspaceConfigs{
		Workspaces:             make(map[string]*WorkspaceConfig),
		SourceDefinitions:      make(map[string]*SourceDefinition),
		DestinationDefinitions: make(map[string]*DestinationDefinition),
	}
}
