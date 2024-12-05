package modelv2_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
)

func TestWorkspaceConfigsUpdatedAt(t *testing.T) {
	wcs := &modelv2.WorkspaceConfigs[string, *modelv2.WorkspaceConfig]{
		Workspaces: map[string]*modelv2.WorkspaceConfig{
			"ws-1": {
				UpdatedAt: time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			"ws-2": {
				UpdatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			"ws-3": nil,
		},
	}

	m := make(map[string]*modelv2.WorkspaceConfig, len(wcs.Workspaces))
	for id, wc := range wcs.List() {
		require.Equal(t, wcs.Workspaces[id], wc)
		m[id] = wc
	}

	require.Len(t, m, 3)
	require.Equal(t, wcs.Workspaces, m)

	require.Equal(t, time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC), m["ws-1"].GetUpdatedAt())
	require.Equal(t, time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC), m["ws-2"].GetUpdatedAt())
	require.True(t, m["ws-3"].IsNil())

	ws, ok := wcs.GetElementByKey("ws-1")
	require.True(t, ok)
	require.Equal(t, wcs.Workspaces["ws-1"], ws)

	ws, ok = wcs.GetElementByKey("ws-2")
	require.True(t, ok)
	require.Equal(t, wcs.Workspaces["ws-2"], ws)

	ws, ok = wcs.GetElementByKey("ws-3")
	require.True(t, ok)
	require.Nil(t, ws)

	wcs.SetElementByKey("ws-3", &modelv2.WorkspaceConfig{
		UpdatedAt: time.Date(2021, 1, 3, 0, 0, 0, 0, time.UTC),
	})
	ws, ok = wcs.GetElementByKey("ws-3")
	require.True(t, ok)
	require.Equal(t, wcs.Workspaces["ws-3"], ws)
	require.Equal(t, time.Date(2021, 1, 3, 0, 0, 0, 0, time.UTC), wcs.Workspaces["ws-3"].GetUpdatedAt())
}
