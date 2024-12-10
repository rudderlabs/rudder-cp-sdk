package modelv2

import (
	"iter"
	"time"

	"github.com/rudderlabs/rudder-cp-sdk/diff"
)

var (
	_ diff.UpdateableObject[string]                       = &WorkspaceConfigs{}
	_ diff.UpdateableList[string, diff.UpdateableElement] = &Workspaces{}
	_ diff.NonUpdateablesList[string, any]                = &SourceDefinitions{}
	_ diff.NonUpdateablesList[string, any]                = &DestinationDefinitions{}
)

// WorkspaceConfigs represents workspace configurations of one or more workspaces, as well as definitions shared by all of them.
type WorkspaceConfigs struct {
	// Workspaces is a map of workspace configurations. The key is a workspace ID.
	Workspaces Workspaces `json:"workspaces"`
	// SourceDefinitions is a map of source definitions. The key is a source definition name.
	SourceDefinitions SourceDefinitions `json:"sourceDefinitions"`
	// DestinationDefinitions is a map of destination definitions. The key is a destination definition name.
	DestinationDefinitions DestinationDefinitions `json:"destinationDefinitions"`
}

func (wcs *WorkspaceConfigs) Updateables() iter.Seq[diff.UpdateableList[string, diff.UpdateableElement]] {
	return func(yield func(diff.UpdateableList[string, diff.UpdateableElement]) bool) {
		yield(&wcs.Workspaces)
	}
}

func (wcs *WorkspaceConfigs) NonUpdateables() iter.Seq[diff.NonUpdateablesList[string, any]] {
	return func(yield func(diff.NonUpdateablesList[string, any]) bool) {
		yield(&wcs.SourceDefinitions)
		yield(&wcs.DestinationDefinitions)
	}
}

type Workspaces map[string]*WorkspaceConfig

func (ws *Workspaces) Type() string { return "Workspaces" }
func (ws *Workspaces) Length() int  { return len(*ws) }
func (ws *Workspaces) Reset()       { *ws = make(map[string]*WorkspaceConfig) }

func (ws *Workspaces) List() iter.Seq2[string, diff.UpdateableElement] {
	return func(yield func(string, diff.UpdateableElement) bool) {
		for key, wc := range *ws {
			if !yield(key, wc) {
				break
			}
		}
	}
}

func (ws *Workspaces) GetElementByKey(id string) (diff.UpdateableElement, bool) {
	wc, ok := (*ws)[id]
	return wc, ok
}

func (ws *Workspaces) SetElementByKey(id string, object diff.UpdateableElement) {
	if *ws == nil {
		*ws = make(map[string]*WorkspaceConfig)
	}
	(*ws)[id] = object.(*WorkspaceConfig)
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

type SourceDefinitions map[string]*SourceDefinition

func (sd *SourceDefinitions) Type() string { return "SourceDefinitions" }
func (sd *SourceDefinitions) Reset()       { *sd = make(map[string]*SourceDefinition) }
func (sd *SourceDefinitions) SetElementByKey(id string, object any) {
	(*sd)[id] = object.(*SourceDefinition)
}

func (sd *SourceDefinitions) List() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		for key, d := range *sd {
			if !yield(key, d) {
				break
			}
		}
	}
}

type DestinationDefinitions map[string]*DestinationDefinition

func (dd *DestinationDefinitions) Type() string { return "DestinationDefinitions" }
func (dd *DestinationDefinitions) Reset()       { *dd = make(map[string]*DestinationDefinition) }
func (dd *DestinationDefinitions) SetElementByKey(id string, object any) {
	(*dd)[id] = object.(*DestinationDefinition)
}

func (dd *DestinationDefinitions) List() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		for key, d := range *dd {
			if !yield(key, d) {
				break
			}
		}
	}
}
