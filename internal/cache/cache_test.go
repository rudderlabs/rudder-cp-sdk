package cache_test

import (
	"sync"
	"testing"
	"time"

	"github.com/rudderlabs/rudder-control-plane-sdk/internal/cache"
	"github.com/rudderlabs/rudder-control-plane-sdk/modelv2"
	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	c := &cache.WorkspaceConfigCache{}
	assert.Nil(t, c.Get(), "cache starts empty")

	c.Set(fullConfig)
	assert.Equal(t, fullConfig, c.Get(), "cache contains full config")

	c.Set(updatedConfig)
	assert.Equal(t, expectedMergedConfig, c.Get(), "cache contains merged config")
}

func TestCacheSubscriptions(t *testing.T) {
	c := &cache.WorkspaceConfigCache{}

	s := c.Subscribe()

	configs := make([]*modelv2.WorkspaceConfigs, 0)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for range s {
			configs = append(configs, c.Get())
			wg.Done()
			if len(configs) == 2 {
				return
			}
		}
	}()

	c.Set(fullConfig)
	c.Set(updatedConfig)

	wg.Wait()

	assert.Equal(t, fullConfig, configs[0])
	assert.Equal(t, expectedMergedConfig, configs[1])
}

var fullConfig = &modelv2.WorkspaceConfigs{
	SourceDefinitions: map[string]*modelv2.SourceDefinition{
		"src-def-1": {
			Name: "src-def-1",
		},
		"src-def-2": {
			Name: "src-def-2",
		},
	},
	DestinationDefinitions: map[string]*modelv2.DestinationDefinition{
		"dst-def-1": {
			Name: "dst-def-1",
		},
		"dst-def-2": {
			Name: "dst-def-2",
		},
	},
	Workspaces: map[string]*modelv2.WorkspaceConfig{
		"ws-1": {
			UpdatedAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		"ws-2": {
			UpdatedAt: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		"ws-3": {
			UpdatedAt: time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
		},
	},
}

var updatedConfig = &modelv2.WorkspaceConfigs{
	SourceDefinitions: map[string]*modelv2.SourceDefinition{
		"src-def-1": {
			Name: "src-def-1-updated",
		},
		"src-def-3": {
			Name: "src-def-3",
		},
	},
	DestinationDefinitions: map[string]*modelv2.DestinationDefinition{
		"dst-def-2": {
			Name: "dst-def-2-updated",
		},
		"dst-def-3": {
			Name: "dst-def-3",
		},
	},
	Workspaces: map[string]*modelv2.WorkspaceConfig{
		"ws-1": {
			UpdatedAt: time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
		},
		// no update for ws-2
		"ws-2": nil,
		// ws-3 removed
		"ws-4": {
			UpdatedAt: time.Date(2020, 1, 5, 0, 0, 0, 0, time.UTC),
		},
	},
}

var expectedMergedConfig = &modelv2.WorkspaceConfigs{
	SourceDefinitions: map[string]*modelv2.SourceDefinition{
		"src-def-1": updatedConfig.SourceDefinitions["src-def-1"],
		"src-def-2": fullConfig.SourceDefinitions["src-def-2"],
		"src-def-3": updatedConfig.SourceDefinitions["src-def-3"],
	},
	DestinationDefinitions: map[string]*modelv2.DestinationDefinition{
		"dst-def-1": fullConfig.DestinationDefinitions["dst-def-1"],
		"dst-def-2": updatedConfig.DestinationDefinitions["dst-def-2"],
		"dst-def-3": updatedConfig.DestinationDefinitions["dst-def-3"],
	},
	Workspaces: map[string]*modelv2.WorkspaceConfig{
		"ws-1": updatedConfig.Workspaces["ws-1"],
		"ws-2": fullConfig.Workspaces["ws-2"],
		"ws-4": updatedConfig.Workspaces["ws-4"],
	},
}
