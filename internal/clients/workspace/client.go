package workspace

import (
	"context"
	"net/http"
	"time"

	"github.com/rudderlabs/rudder-cp-sdk/identity"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/base"
)

type Client struct {
	*base.Client

	Identity *identity.Workspace
	// Secrets controls whether account secrets are embedded in workspace configs responses.
	// If empty, the control plane's default applies.
	Secrets string
}

func (c *Client) Get(ctx context.Context, path string, queryOpts ...base.QueryOption) (*http.Request, error) {
	req, err := c.Client.Get(ctx, path, queryOpts...)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.Identity.WorkspaceToken, "")

	return req, nil
}

func (c *Client) GetWorkspaceConfigs(ctx context.Context, object any, updatedAfter time.Time) error {
	req, err := c.Get(ctx, "/data-plane/v2/workspaceConfig",
		base.WithUpdatedAfter(updatedAfter), base.WithSecrets(c.Secrets))
	if err != nil {
		return err
	}
	return c.SendWorkspaceConfigs(req, object)
}

func (c *Client) GetNamespaceWorkspaces(ctx context.Context) (workspaceIDs []string, err error) {
	return nil, base.ErrUnsupportedOperation
}
