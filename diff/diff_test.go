package diff

import (
	stdjson "encoding/json"
	"iter"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUpdateable(t *testing.T) {
	// Get the data from the first call and make sure it unmarshals correctly
	firstCall, err := os.ReadFile("./testdata/call_01.json")
	require.NoError(t, err)

	var response UpdateableObject[string] = &WorkspaceConfigs{}
	err = stdjson.Unmarshal(firstCall, &response)
	require.NoError(t, err)
	{
		workspaces := getWorkspaces(response)
		require.Len(t, workspaces, 2)
		require.Contains(t, workspaces, "workspace1")
		require.Contains(t, workspaces, "workspace2")
		require.Equal(t, goldenWorkspace1, workspaces["workspace1"])
		require.Equal(t, goldenWorkspace2, workspaces["workspace2"])
		srcDefinitions := getSourceDefinitions(response)
		require.Len(t, srcDefinitions, 1)
		require.Contains(t, srcDefinitions, "close_crm")
		require.Equal(t, &SourceDefinition{Name: "Close CRM"}, srcDefinitions["close_crm"])
		dstDefinitions := getDestinationDefinitions(response)
		require.Len(t, dstDefinitions, 0)
	}

	// Update the cache with the first call response and make sure the workspaces are correct
	var cache UpdateableObject[string] = &WorkspaceConfigs{}
	updater := &Updater[string]{}
	updateAfter, updated, err := updater.UpdateCache(response, cache)
	require.NoError(t, err)
	require.True(t, updated)
	require.Equal(t, time.Date(2021, 9, 1, 6, 6, 6, 0, time.UTC), updateAfter)
	{
		workspaces := getWorkspaces(cache)
		require.Len(t, workspaces, 2)
		require.Contains(t, workspaces, "workspace1")
		require.Contains(t, workspaces, "workspace2")
		require.Equal(t, goldenWorkspace1, workspaces["workspace1"])
		require.Equal(t, goldenWorkspace2, workspaces["workspace2"])
		srcDefinitions := getSourceDefinitions(cache)
		require.Len(t, srcDefinitions, 1)
		require.Contains(t, srcDefinitions, "close_crm")
		require.Equal(t, &SourceDefinition{Name: "Close CRM"}, srcDefinitions["close_crm"])
		dstDefinitions := getDestinationDefinitions(cache)
		require.Len(t, dstDefinitions, 0)
	}

	// in the second call we get the same two workspaces but with no updates so they'll both be null.
	// therefore nothing should change in the cache.
	secondCall, err := os.ReadFile("./testdata/call_02.json")
	require.NoError(t, err)

	response = &WorkspaceConfigs{}
	err = stdjson.Unmarshal(secondCall, &response)
	require.NoError(t, err)
	{
		workspaces := getWorkspaces(response)
		require.Len(t, workspaces, 2)
		require.Contains(t, workspaces, "workspace1")
		require.Contains(t, workspaces, "workspace2")
		require.Nil(t, workspaces["workspace1"])
		require.Nil(t, workspaces["workspace2"])
		srcDefinitions := getSourceDefinitions(response)
		require.Len(t, srcDefinitions, 1)
		require.Contains(t, srcDefinitions, "close_crm")
		require.Equal(t, &SourceDefinition{Name: "Close CRM"}, srcDefinitions["close_crm"])
		dstDefinitions := getDestinationDefinitions(response)
		require.Len(t, dstDefinitions, 0)
	}

	updateAfter, updated, err = updater.UpdateCache(response, cache)
	require.NoError(t, err)
	require.False(t, updated)
	require.Equal(t, time.Date(2021, 9, 1, 6, 6, 6, 0, time.UTC), updateAfter)
	{
		workspaces := getWorkspaces(cache)
		require.Len(t, workspaces, 2)
		require.Contains(t, workspaces, "workspace1")
		require.Contains(t, workspaces, "workspace2")
		require.Equal(t, goldenWorkspace1, workspaces["workspace1"])
		require.Equal(t, goldenWorkspace2, workspaces["workspace2"])
		srcDefinitions := getSourceDefinitions(cache)
		require.Len(t, srcDefinitions, 1)
		require.Contains(t, srcDefinitions, "close_crm")
		require.Equal(t, &SourceDefinition{Name: "Close CRM"}, srcDefinitions["close_crm"])
		dstDefinitions := getDestinationDefinitions(cache)
		require.Len(t, dstDefinitions, 0)
	}

	// in the third call workspace1 is not updated, workspace2 is deleted, and we receive a new workspace3.
	thirdCall, err := os.ReadFile("./testdata/call_03.json")
	require.NoError(t, err)

	response = &WorkspaceConfigs{}
	err = stdjson.Unmarshal(thirdCall, &response)
	require.NoError(t, err)
	{
		workspaces := getWorkspaces(response)
		require.Len(t, workspaces, 2)
		require.Contains(t, workspaces, "workspace1")
		require.NotContains(t, workspaces, "workspace2")
		require.Contains(t, workspaces, "workspace3")
		require.Nil(t, workspaces["workspace1"])
		require.Equal(t, goldenWorkspace3, workspaces["workspace3"])
		srcDefinitions := getSourceDefinitions(response)
		require.Len(t, srcDefinitions, 1)
		require.Contains(t, srcDefinitions, "close_crm")
		require.Equal(t, &SourceDefinition{Name: "Close CRM"}, srcDefinitions["close_crm"])
		dstDefinitions := getDestinationDefinitions(response)
		require.Len(t, dstDefinitions, 0)
	}

	updateAfter, updated, err = updater.UpdateCache(response, cache)
	require.NoError(t, err)
	require.True(t, updated)
	require.Equal(t, time.Date(2021, 9, 1, 6, 6, 7, 0, time.UTC), updateAfter)
	{
		workspaces := getWorkspaces(cache)
		require.Len(t, workspaces, 2)
		require.Contains(t, workspaces, "workspace1")
		require.Contains(t, workspaces, "workspace3")
		require.Equal(t, goldenWorkspace1, workspaces["workspace1"])
		require.Equal(t, goldenWorkspace3, workspaces["workspace3"])
	}

	// in the fourth call workspace1 is not updated but workspace3 is.
	fourthCall, err := os.ReadFile("./testdata/call_04.json")
	require.NoError(t, err)

	response = &WorkspaceConfigs{}
	err = stdjson.Unmarshal(fourthCall, &response)
	require.NoError(t, err)
	{
		workspaces := getWorkspaces(response)
		require.Len(t, workspaces, 2)
		require.Contains(t, workspaces, "workspace1")
		require.Contains(t, workspaces, "workspace3")
		require.Nil(t, workspaces["workspace1"])
		require.Equal(t, goldenUpdatedWorkspace3, workspaces["workspace3"])
		srcDefinitions := getSourceDefinitions(response)
		require.Len(t, srcDefinitions, 2)
		require.Contains(t, srcDefinitions, "close_crm")
		require.Equal(t, &SourceDefinition{Name: "Close CRM"}, srcDefinitions["close_crm"])
		require.Contains(t, srcDefinitions, "singer-klaviyo")
		require.Equal(t, &SourceDefinition{Name: "Klaviyo"}, srcDefinitions["singer-klaviyo"])
		dstDefinitions := getDestinationDefinitions(response)
		require.Len(t, dstDefinitions, 1)
		require.Contains(t, dstDefinitions, "LINKEDIN_ADS")
		require.Equal(t, &DestinationDefinition{Name: "LinkedIn Ads"}, dstDefinitions["LINKEDIN_ADS"])
	}

	updateAfter, updated, err = updater.UpdateCache(response, cache)
	require.NoError(t, err)
	require.True(t, updated)
	require.Equal(t, time.Date(2021, 9, 1, 6, 6, 8, 0, time.UTC), updateAfter)
	{
		workspaces := getWorkspaces(cache)
		require.Len(t, workspaces, 2)
		require.Contains(t, workspaces, "workspace1")
		require.Contains(t, workspaces, "workspace3")
		require.Equal(t, goldenWorkspace1, workspaces["workspace1"])
		require.Equal(t, goldenUpdatedWorkspace3, workspaces["workspace3"])
		srcDefinitions := getSourceDefinitions(cache)
		require.Len(t, srcDefinitions, 2)
		require.Contains(t, srcDefinitions, "close_crm")
		require.Equal(t, &SourceDefinition{Name: "Close CRM"}, srcDefinitions["close_crm"])
		require.Contains(t, srcDefinitions, "singer-klaviyo")
		require.Equal(t, &SourceDefinition{Name: "Klaviyo"}, srcDefinitions["singer-klaviyo"])
		dstDefinitions := getDestinationDefinitions(cache)
		require.Len(t, dstDefinitions, 1)
		require.Contains(t, dstDefinitions, "LINKEDIN_ADS")
		require.Equal(t, &DestinationDefinition{Name: "LinkedIn Ads"}, dstDefinitions["LINKEDIN_ADS"])
	}

	// in the fifth call we didn't receive any updates, so the cache should remain the same, and we should receive an error
	fifthCall := []byte(`{}`)
	response = &WorkspaceConfigs{}
	err = stdjson.Unmarshal(fifthCall, &response)
	require.NoError(t, err)
	_, _, err = updater.UpdateCache(response, cache)
	require.Error(t, err)
}

type WorkspaceConfigs struct {
	Workspaces             Workspaces             `json:"workspaces"`
	SourceDefinitions      SourceDefinitions      `json:"sourceDefinitions"`
	DestinationDefinitions DestinationDefinitions `json:"destinationDefinitions"`
}

func (wcs *WorkspaceConfigs) Updateables() iter.Seq[UpdateableList[string, UpdateableElement]] {
	return func(yield func(UpdateableList[string, UpdateableElement]) bool) {
		yield(&wcs.Workspaces)
	}
}

func (wcs *WorkspaceConfigs) NonUpdateables() iter.Seq[NonUpdateablesList[string, any]] {
	return func(yield func(NonUpdateablesList[string, any]) bool) {
		if !yield(&wcs.SourceDefinitions) {
			return
		}
		if !yield(&wcs.DestinationDefinitions) {
			return
		}
	}
}

type Workspaces map[string]*WorkspaceConfig

func (ws *Workspaces) Type() string { return "Workspaces" }
func (ws *Workspaces) Length() int  { return len(*ws) }
func (ws *Workspaces) Reset()       { *ws = make(map[string]*WorkspaceConfig) }

func (ws *Workspaces) List() iter.Seq2[string, UpdateableElement] {
	return func(yield func(string, UpdateableElement) bool) {
		for key, wc := range *ws {
			if !yield(key, wc) {
				break
			}
		}
	}
}

func (ws *Workspaces) GetElementByKey(id string) (UpdateableElement, bool) {
	wc, ok := (*ws)[id]
	return wc, ok
}

func (ws *Workspaces) SetElementByKey(id string, object UpdateableElement) {
	if *ws == nil {
		*ws = make(map[string]*WorkspaceConfig)
	}
	(*ws)[id] = object.(*WorkspaceConfig)
}

type WorkspaceConfig struct {
	Sources      map[string]*Source      `json:"sources"`
	Destinations map[string]*Destination `json:"destinations"`
	Connections  map[string]*Connection  `json:"connections"`
	EventReplays map[string]*EventReplay `json:"eventReplays"`
	UpdatedAt    time.Time               `json:"updatedAt"`
}

func (wc *WorkspaceConfig) GetUpdatedAt() time.Time { return wc.UpdatedAt }
func (wc *WorkspaceConfig) IsNil() bool             { return wc == nil }

type Source struct {
	Name           string             `json:"name"`
	WriteKey       string             `json:"writeKey"`
	Enabled        bool               `json:"enabled"`
	DefinitionName string             `json:"sourceDefinitionName"`
	Config         stdjson.RawMessage `json:"config"`
	Deleted        bool               `json:"deleted"`
}

type Destination struct {
	Enabled bool `json:"enabled"`
}

type Connection struct {
	SourceID         string `json:"sourceId"`
	DestinationID    string `json:"destinationId"`
	ProcessorEnabled bool   `json:"processorEnabled"`
}

type EventReplay struct {
	Sources      map[string]*SourceReplay      `json:"sources"`
	Destinations map[string]*DestinationReplay `json:"destinations"`
	Connections  []ConnectionReplay            `json:"connections"`
}

type SourceReplay struct {
	OriginalID string `json:"originalId"`
}

type DestinationReplay struct {
	OriginalID string `json:"originalId"`
}

type ConnectionReplay struct {
	SourceID      string `json:"sourceId"`
	DestinationID string `json:"destinationId"`
}

type SourceDefinition struct {
	Name string `json:"name"`
}

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

type DestinationDefinition struct {
	Name string `json:"name"`
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

func getWorkspaces(v UpdateableObject[string]) map[string]*WorkspaceConfig {
	return v.(*WorkspaceConfigs).Workspaces
}

func getSourceDefinitions(v UpdateableObject[string]) map[string]*SourceDefinition {
	return v.(*WorkspaceConfigs).SourceDefinitions
}

func getDestinationDefinitions(v UpdateableObject[string]) map[string]*DestinationDefinition {
	return v.(*WorkspaceConfigs).DestinationDefinitions
}
