package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	model "github.com/rudderlabs/rudder-cp-sdk/model/wsconfigv2"
)

func TestParse(t *testing.T) {
	data, err := os.ReadFile("testdata/workspace_configs.v2.json")
	require.NoError(t, err)

	wcs, err := Parse(data)
	require.NoError(t, err)

	require.Equal(t, wcs, &model.WorkspaceConfigs{
		Workspaces: map[string]*model.WorkspaceConfig{
			"ws-1": {
				Sources: map[string]*model.Source{
					"src-1-1": {},
					"src-1-2": {},
				},
				Destinations: map[string]*model.Destination{
					"dst-1-1": {},
					"dst-1-2": {},
				},
			},
		},
		SourceDefinitions: map[string]*model.SourceDefinition{
			"src-def-1": {},
		},
		DestinationDefinitions: map[string]*model.DestinationDefinition{
			"src-def-2": {},
		},
	})
}

func TestParseError(t *testing.T) {
	wcs, err := Parse([]byte(`{ malformed json }`))
	require.Nil(t, wcs)
	require.Error(t, err)
}
