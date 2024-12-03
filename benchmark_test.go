package cpsdk

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

func BenchmarkGetWorkspaceConfigs(b *testing.B) {
	conf := config.New()
	baseURL := conf.GetString("BASE_URL", "https://dp.api.rudderstack.com/")
	namespace := conf.GetString("NAMESPACE", "free-us-1")
	identity := conf.GetString("IDENTITY", "")
	if identity == "" {
		b.Skip("IDENTITY is not set")
		return
	}

	cpSDK, err := New(
		WithBaseUrl(baseURL),
		WithLogger(logger.NOP),
		WithPollingInterval(0), // Setting the poller interval to 0 to disable the poller
		WithNamespaceIdentity(namespace, identity),
	)
	require.NoError(b, err)
	defer cpSDK.Close()

	var workspaceConfigs WorkspaceConfigs
	err = cpSDK.GetWorkspaceConfigs(context.Background(), &workspaceConfigs, time.Time{})
	require.NoError(b, err)
	require.NotNil(b, workspaceConfigs)
	require.Greater(b, len(workspaceConfigs.Workspaces), 0)
}

type WorkspaceConfigs struct {
	Workspaces        map[string]*WorkspaceConfig  `json:"workspaces"`
	SourceDefinitions map[string]*SourceDefinition `json:"sourceDefinitions"`
}

type WorkspaceConfig struct {
	Sources      map[string]*Source      `json:"sources"`
	Destinations map[string]*Destination `json:"destinations"`
	Connections  map[string]*Connection  `json:"connections"`
}

type Source struct {
	Name           string `json:"name"`
	WriteKey       string `json:"writeKey"`
	Enabled        bool   `json:"enabled"`
	DefinitionName string `json:"sourceDefinitionName"`
	Deleted        bool   `json:"deleted"`
}

type Destination struct {
	Enabled bool `json:"enabled"`
}

type Connection struct {
	SourceID         string `json:"sourceId"`
	DestinationID    string `json:"destinationId"`
	Enabled          bool   `json:"enabled"`
	ProcessorEnabled bool   `json:"processorEnabled"`
}

type SourceDefinition struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}
