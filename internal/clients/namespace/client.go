package namespace

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rudderlabs/rudder-cp-sdk/identity"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/base"
	"github.com/rudderlabs/rudder-go-kit/jsonrs"
)

type Client struct {
	*base.Client

	Identity *identity.Namespace
}

func (c *Client) Get(ctx context.Context, path string, updatedAfter time.Time) (*http.Request, error) {
	req, err := c.Client.Get(ctx, path, updatedAfter)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.Identity.Secret, "")

	return req, nil
}

func (c *Client) getWorkspaceConfigsReader(ctx context.Context, updatedAfter time.Time) (io.ReadCloser, error) {
	req, err := c.Get(ctx, "/configuration/v2/namespaces/"+c.Identity.Namespace, updatedAfter)
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
