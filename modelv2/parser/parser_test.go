package parser_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
	"github.com/rudderlabs/rudder-cp-sdk/modelv2/parser"
)

func TestParse(t *testing.T) {
	data, err := os.Open("testdata/workspace_configs.v2.json")
	require.NoError(t, err)

	wcs, err := parser.Parse(data)
	require.NoError(t, err)

	require.Equal(t, wcs, &modelv2.WorkspaceConfigs{
		Workspaces: map[string]*modelv2.WorkspaceConfig{
			"ws-1": {
				Sources: map[string]*modelv2.Source{
					"src-1-1": {},
					"src-1-2": {},
				},
				Destinations: map[string]*modelv2.Destination{
					"dst-1-1": {},
					"dst-1-2": {},
				},
			},
		},
		SourceDefinitions: map[string]*modelv2.SourceDefinition{
			"src-def-1": {},
		},
		DestinationDefinitions: map[string]*modelv2.DestinationDefinition{
			"src-def-2": {},
		},
	})
}

func TestParseError(t *testing.T) {
	wcs, err := parser.Parse(bytes.NewReader([]byte(`{ malformed json }`)))
	require.Nil(t, wcs)
	require.Error(t, err)
}
