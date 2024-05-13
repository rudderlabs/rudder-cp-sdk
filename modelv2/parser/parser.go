package parser

import (
	"io"

	jsoniter "github.com/json-iterator/go"

	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
)

var json = jsoniter.ConfigFastest

func Parse(reader io.Reader) (*modelv2.WorkspaceConfigs, error) {
	res := &modelv2.WorkspaceConfigs{}

	if err := json.NewDecoder(reader).Decode(res); err != nil {
		return nil, err
	}

	return res, nil
}
