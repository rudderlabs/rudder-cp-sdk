package workspace

import (
	"context"
	"net/http"

	"github.com/rudderlabs/rudder-control-plane-sdk/identity"
	"github.com/rudderlabs/rudder-control-plane-sdk/internal/clients/base"
	"github.com/rudderlabs/rudder-control-plane-sdk/model"
)

type Client struct {
	*base.Client

	Identity *identity.Workspace
}

func (c *Client) Get(ctx context.Context, path string) (*http.Request, error) {
	req, err := c.Client.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.Identity.WorkspaceToken, "")
	return req, nil
}

func (c *Client) GetWorkspaceConfigs(ctx context.Context) (*model.WorkspaceConfigs, error) {
	req, err := c.Get(ctx, "/data-plane/v2/workspaceConfig")
	if err != nil {
		return nil, err
	}

	data, err := c.Send(req)
	if err != nil {
		return nil, err
	}

	wcs, err := model.ParseV2WorkspaceConfigs(data)
	if err != nil {
		return nil, err
	}

	return wcs, nil
}
