package admin

import (
	"github.com/rudderlabs/rudder-cp-sdk/client/base"
)

type Client struct {
	*base.Client

	Username string
	Password string
}

func New() *Client {
	return &Client{}
}
