package parser

import (
	jsoniter "github.com/json-iterator/go"

	model "github.com/rudderlabs/rudder-cp-sdk/model/wsconfigv2"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func Parse(data []byte) (*model.WorkspaceConfigs, error) {
	res := &model.WorkspaceConfigs{}

	if err := json.Unmarshal(data, res); err != nil {
		return nil, err
	}

	return res, nil
}
