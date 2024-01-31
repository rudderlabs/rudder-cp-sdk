package modelv2_test

import (
	"testing"
	"time"

	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
	"github.com/stretchr/testify/assert"
)

func TestWorkspaceConfigsUpdatedAt(t *testing.T) {
	wcs := &modelv2.WorkspaceConfigs{
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

	updatedAt := wcs.UpdatedAt()
	assert.Equal(t, time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC), updatedAt, "UpdatedAt should return the maximum of all workspace's UpdatedAt")
}
