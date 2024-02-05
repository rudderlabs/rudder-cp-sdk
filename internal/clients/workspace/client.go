package workspace

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/rudderlabs/rudder-cp-sdk/identity"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/base"
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

func (c *Client) GetWorkspaceConfigs(ctx context.Context) ([]byte, error) {
	req, err := c.Get(ctx, "/data-plane/v2/workspaceConfig")
	if err != nil {
		return nil, err
	}

	data, err := c.Send(req)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Client) GetUpdatedWorkspaceConfigs(ctx context.Context, updatedAfter time.Time) ([]byte, error) {
	return nil, errors.New("not implemented")
}
