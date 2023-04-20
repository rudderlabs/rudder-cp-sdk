package model_test

import (
	"io/ioutil"
	"testing"

	"github.com/rudderlabs/rudder-control-plane-sdk/model"
	"github.com/stretchr/testify/assert"
)

func TestParseV2WorkspaceConfigs(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/workspace_configs.v2.json")
	if err != nil {
		t.Fatal(err)
	}

	wcs, err := model.ParseV2WorkspaceConfigs(data)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, wcs, &model.WorkspaceConfigs{
		Version: 2,
		SourceDefinitions: map[string]*model.SourceDefinition{
			"src-def-1": {},
		},
		DestinationDefinitions: map[string]*model.DestinationDefinition{
			"src-def-2": {},
		},
	})
}
