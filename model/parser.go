package model

import (
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func ParseV2WorkspaceConfigs(data []byte) (*WorkspaceConfigs, error) {
	res := &WorkspaceConfigs{}

	if err := json.Unmarshal(data, res); err != nil {
		return nil, err
	}

	return res, nil
}
