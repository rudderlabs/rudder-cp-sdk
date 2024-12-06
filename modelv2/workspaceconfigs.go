package modelv2

import (
	"iter"
	"time"

	"github.com/rudderlabs/rudder-cp-sdk/diff"
)

var _ diff.UpdateableList[string, *WorkspaceConfig] = &WorkspaceConfigs[string, *WorkspaceConfig]{}

// WorkspaceConfigs represents workspace configurations of one or more workspaces, as well as definitions shared by all of them.
type WorkspaceConfigs[K string, T *WorkspaceConfig] struct {
	// Workspaces is a map of workspace configurations. The key is a workspace ID.
	Workspaces map[K]T `json:"workspaces"`
	// SourceDefinitions is a map of source definitions. The key is a source definition name.
	SourceDefinitions map[string]*SourceDefinition `json:"sourceDefinitions"`
	// DestinationDefinitions is a map of destination definitions. The key is a destination definition name.
	DestinationDefinitions map[string]*DestinationDefinition `json:"destinationDefinitions"`
}

func (wcs *WorkspaceConfigs[K, T]) Length() int { return len(wcs.Workspaces) }

func (wcs *WorkspaceConfigs[K, T]) List() iter.Seq2[K, T] {
	return func(yield func(K, T) bool) {
		for key, wc := range wcs.Workspaces {
			if !yield(key, wc) {
				break
			}
		}
	}
}

func (wcs *WorkspaceConfigs[K, T]) GetElementByKey(id K) (T, bool) {
	wc, ok := wcs.Workspaces[id]
	return wc, ok
}

func (wcs *WorkspaceConfigs[K, T]) SetElementByKey(id K, object T) {
	if wcs.Workspaces == nil {
		wcs.Workspaces = make(map[K]T)
	}
	wcs.Workspaces[id] = object
}

func (wcs *WorkspaceConfigs[K, T]) Reset() {
	wcs.Workspaces = make(map[K]T)
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

func (wc *WorkspaceConfig) GetUpdatedAt() time.Time { return wc.UpdatedAt }
func (wc *WorkspaceConfig) IsNil() bool             { return wc == nil }
