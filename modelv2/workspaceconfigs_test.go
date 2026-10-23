package modelv2_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/jsonrs"

	"github.com/rudderlabs/rudder-cp-sdk/diff"
	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
)

func TestWorkspaceConfigsUpdatedAt(t *testing.T) {
	wcs := &modelv2.WorkspaceConfigs{
		Workspaces: modelv2.Workspaces{
			"ws-1": {
				UpdatedAt: time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			"ws-2": {
				UpdatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			"ws-3": nil,
		},
	}

	m := make(modelv2.Workspaces, len(wcs.Workspaces))
	for uo := range wcs.Updateables() {
		for id, wc := range uo.List() {
			require.Equal(t, wcs.Workspaces[id], wc)
			m[id] = wc.(*modelv2.WorkspaceConfig)
		}
	}

	require.Len(t, m, 3)
	require.Equal(t, wcs.Workspaces, m)

	require.Equal(t, time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC), m["ws-1"].GetUpdatedAt())
	require.Equal(t, time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC), m["ws-2"].GetUpdatedAt())
	require.True(t, m["ws-3"].IsNil())

	var updateableList diff.UpdateableList[string, diff.UpdateableElement] = &wcs.Workspaces
	ws, ok := updateableList.GetElementByKey("ws-1")
	require.True(t, ok)
	require.Equal(t, wcs.Workspaces["ws-1"], ws)

	ws, ok = updateableList.GetElementByKey("ws-2")
	require.True(t, ok)
	require.Equal(t, wcs.Workspaces["ws-2"], ws)

	ws, ok = updateableList.GetElementByKey("ws-3")
	require.True(t, ok)
	require.Nil(t, ws)

	updateableList.SetElementByKey("ws-3", &modelv2.WorkspaceConfig{
		UpdatedAt: time.Date(2021, 1, 3, 0, 0, 0, 0, time.UTC),
	})
	ws, ok = updateableList.GetElementByKey("ws-3")
	require.True(t, ok)
	require.Equal(t, wcs.Workspaces["ws-3"], ws)
	require.Equal(t, time.Date(2021, 1, 3, 0, 0, 0, 0, time.UTC), wcs.Workspaces["ws-3"].GetUpdatedAt())
}

func TestNonUpdateables(t *testing.T) {
	wcs := &modelv2.WorkspaceConfigs{
		Workspaces: modelv2.Workspaces{
			"ws-1": {
				UpdatedAt: time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			"ws-2": {
				UpdatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			"ws-3": nil,
		},
		SourceDefinitions: modelv2.SourceDefinitions{
			"source-1": &modelv2.SourceDefinition{
				Name: "Source Definition 1",
			},
		},
		DestinationDefinitions: modelv2.DestinationDefinitions{
			"destination-1": &modelv2.DestinationDefinition{
				Name: "Destination Definition 1",
			},
		},
	}

	nonUpdateables := make([]diff.NonUpdateablesList[string, any], 0, 2)
	for nu := range wcs.NonUpdateables() {
		nonUpdateables = append(nonUpdateables, nu)
	}

	require.Len(t, nonUpdateables, 2)
	require.Equal(t, &wcs.SourceDefinitions, nonUpdateables[0])
	require.Equal(t, &wcs.DestinationDefinitions, nonUpdateables[1])
}

func TestOmittedDefinitions(t *testing.T) {
	for _, lib := range []string{jsonrs.StdLib, jsonrs.SonnetLib, jsonrs.JsoniterLib} {
		t.Run(lib, func(t *testing.T) {
			for _, tc := range []struct {
				name     string
				data     string
				omitted  bool
				expected modelv2.SourceDefinitions
			}{
				{name: "null", data: `null`, omitted: true},
				{name: "empty", data: `{}`, expected: modelv2.SourceDefinitions{}},
				{name: "sent", data: `{"webhook":{"name":"webhook"}}`, expected: modelv2.SourceDefinitions{"webhook": {Name: "webhook"}}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					var wcs modelv2.WorkspaceConfigs
					data := `{"sourceDefinitions":` + tc.data + `,"destinationDefinitions":` + tc.data + `,"IncrementalDefinitionsState":{"confirmed":true}}`
					require.NoError(t, jsonrs.NewWithLibrary(lib).Unmarshal([]byte(data), &wcs))

					require.Equal(t, tc.expected, wcs.SourceDefinitions)
					require.Equal(t, tc.omitted, wcs.SourceDefinitions.Omitted())
					require.Equal(t, tc.omitted, wcs.DestinationDefinitions.Omitted())
					confirmed, _ := wcs.IncrementalDefinitions()
					require.False(t, confirmed, "the state should not be decoded")
				})
			}
		})
	}
}
