package parser

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func Parse(data []byte) (*modelv2.WorkspaceConfigs, error) {
	res := &modelv2.WorkspaceConfigs{}

	if err := json.Unmarshal(data, res); err != nil {
		return nil, err
	}

	return res, nil
}
