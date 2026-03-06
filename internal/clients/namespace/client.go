package namespace

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

	Identity *identity.Namespace
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

func (c *Client) GetNamespaceWorkspaces(ctx context.Context) ([]string, error) {
	type response struct {
		Data []string `json:"data"`
	}
	req, err := c.GetWithAuth(ctx, "/configuration/v2/namespaces/"+c.Identity.Namespace+"/workspace-ids")
	if err != nil {
		return nil, fmt.Errorf("creating get request: %w", err)
	}
	reader, err := c.Send(req)
	if err != nil {
		return nil, fmt.Errorf("sending get request: %w", err)
	}
	defer func() { _ = reader.Close() }()

	var res response
	if err = jsonrs.NewDecoder(reader).Decode(&res); err != nil {
		return nil, fmt.Errorf("decoding namespace workspaces response: %w", err)
	}
	return res.Data, nil
}

func (c *Client) GetWithAuth(ctx context.Context, path string, opts ...base.QueryOption) (*http.Request, error) {
	req, err := c.Get(ctx, path, opts...)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Identity.Secret, "")
	return req, nil
}

func (c *Client) getWorkspaceConfigsReader(ctx context.Context, updatedAfter time.Time) (io.ReadCloser, error) {
	req, err := c.GetWithAuth(ctx, "/configuration/v2/namespaces/"+c.Identity.Namespace, base.WithUpdatedAfter(updatedAfter))
	if err != nil {
		return nil, err
	}
	return c.Send(req)
}
