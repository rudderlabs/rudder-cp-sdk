package workspace

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rudderlabs/rudder-go-kit/jsonrs"

	"github.com/rudderlabs/rudder-cp-sdk/identity"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/base"
)

type Client struct {
	*base.Client

	Identity *identity.Workspace
}

func (c *Client) Get(ctx context.Context, path string, opts ...base.QueryOption) (*http.Request, error) {
	req, err := c.Client.Get(ctx, path, opts...)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.Identity.WorkspaceToken, "")

	return req, nil
}

func (c *Client) getWorkspaceConfigsReader(ctx context.Context, updatedAfter time.Time) (io.ReadCloser, error) {
	req, err := c.Get(ctx, "/data-plane/v2/workspaceConfig", base.WithUpdatedAfter(updatedAfter))
	if err != nil {
		return nil, err
	}

	return c.Send(req)
}

func (c *Client) GetWorkspaceConfigs(ctx context.Context, object any, updatedAfter time.Time) error {
	reader, err := c.getWorkspaceConfigsReader(ctx, updatedAfter)
	if err != nil {
		return err
	}

	defer func() { _ = reader.Close() }()

	if err = jsonrs.NewDecoder(reader).Decode(object); err != nil {
		return fmt.Errorf("failed to decode workspace configs: %w", err)
	}

	return nil
}

func (c *Client) GetNamespaceWorkspaces(ctx context.Context) (workspaceIDs []string, err error) {
	return nil, base.ErrUnsupportedOperation
}
