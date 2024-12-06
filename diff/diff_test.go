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

	var response UpdateableList[string, *WorkspaceConfig] = &WorkspaceConfigs[string, *WorkspaceConfig]{}
	err = stdjson.Unmarshal(firstCall, &response)
	require.NoError(t, err)
	{
		workspaces := getWorkspaces(response)
		require.Len(t, workspaces, 2)
		require.Contains(t, workspaces, "workspace1")
		require.Contains(t, workspaces, "workspace2")
		require.Equal(t, goldenWorkspace1, workspaces["workspace1"])
		require.Equal(t, goldenWorkspace2, workspaces["workspace2"])
	}

	// Update the cache with the first call response and make sure the workspaces are correct
	var cache UpdateableList[string, *WorkspaceConfig] = &WorkspaceConfigs[string, *WorkspaceConfig]{}
	updater := &Updater[string, *WorkspaceConfig]{}
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
	}

	// in the second call we get the same two workspaces but with no updates so they'll both be null.
	// therefore nothing should change in the cache.
	secondCall, err := os.ReadFile("./testdata/call_02.json")
	require.NoError(t, err)

	response = &WorkspaceConfigs[string, *WorkspaceConfig]{}
	err = stdjson.Unmarshal(secondCall, &response)
	require.NoError(t, err)
	{
		workspaces := getWorkspaces(response)
		require.Len(t, workspaces, 2)
		require.Contains(t, workspaces, "workspace1")
		require.Contains(t, workspaces, "workspace2")
		require.Nil(t, workspaces["workspace1"])
		require.Nil(t, workspaces["workspace2"])
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
	}

	// in the third call workspace1 is not updated, workspace2 is deleted, and we receive a new workspace3.
	thirdCall, err := os.ReadFile("./testdata/call_03.json")
	require.NoError(t, err)

	response = &WorkspaceConfigs[string, *WorkspaceConfig]{}
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

	response = &WorkspaceConfigs[string, *WorkspaceConfig]{}
	err = stdjson.Unmarshal(fourthCall, &response)
	require.NoError(t, err)
	{
		workspaces := getWorkspaces(response)
		require.Len(t, workspaces, 2)
		require.Contains(t, workspaces, "workspace1")
		require.Contains(t, workspaces, "workspace3")
		require.Nil(t, workspaces["workspace1"])
		require.Equal(t, goldenUpdatedWorkspace3, workspaces["workspace3"])
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
	}
}

type WorkspaceConfigs[K string, T *WorkspaceConfig] struct {
	Workspaces map[string]*WorkspaceConfig `json:"workspaces"`
}

func (wcs *WorkspaceConfigs[K, T]) Length() int { return len(wcs.Workspaces) }

func (wcs *WorkspaceConfigs[K, T]) List() iter.Seq2[K, T] {
	return func(yield func(K, T) bool) {
		for key, wc := range wcs.Workspaces {
			if !yield(K(key), wc) {
				break
			}
		}
	}
}

func (wcs *WorkspaceConfigs[K, T]) GetElementByKey(id string) (T, bool) {
	wc, ok := wcs.Workspaces[id]
	return wc, ok
}

func (wcs *WorkspaceConfigs[K, T]) SetElementByKey(id string, object T) {
	if wcs.Workspaces == nil {
		wcs.Workspaces = make(map[string]*WorkspaceConfig)
	}
	wcs.Workspaces[id] = object
}

func (wcs *WorkspaceConfigs[K, T]) Reset() {
	wcs.Workspaces = make(map[string]*WorkspaceConfig)
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

func getWorkspaces(v UpdateableList[string, *WorkspaceConfig]) map[string]*WorkspaceConfig {
	return v.(*WorkspaceConfigs[string, *WorkspaceConfig]).Workspaces
}
