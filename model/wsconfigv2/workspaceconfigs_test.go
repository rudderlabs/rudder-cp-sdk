package wsconfigv2_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	model "github.com/rudderlabs/rudder-cp-sdk/model/wsconfigv2"
)

func TestWorkspaceConfigsUpdatedAt(t *testing.T) {
	wcs := &model.WorkspaceConfigs{
		Workspaces: map[string]*model.WorkspaceConfig{
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
