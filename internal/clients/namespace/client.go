package namespace

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/rudderlabs/rudder-cp-sdk/identity"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/base"
	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
	"github.com/rudderlabs/rudder-cp-sdk/modelv2/parser"
)

var json = jsoniter.ConfigFastest

type Client struct {
	*base.Client

	Identity *identity.Namespace
}

func (c *Client) Get(ctx context.Context, path string) (*http.Request, error) {
	req, err := c.Client.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.Identity.Secret, "")

	return req, nil
}

func (c *Client) getWorkspaceConfigsReader(ctx context.Context) (io.ReadCloser, error) {
	req, err := c.Get(ctx, "/configuration/v2/namespaces/"+c.Identity.Namespace)
	if err != nil {
		return nil, err
	}

	return c.Send(req)
}

func (c *Client) GetWorkspaceConfigs(ctx context.Context) (*modelv2.WorkspaceConfigs, error) {
	reader, err := c.getWorkspaceConfigsReader(ctx)
	if err != nil {
		return nil, err
	}

	defer func() { _ = reader.Close() }()

	wcs, err := parser.Parse(reader)
	if err != nil {
		return nil, err
	}

	return wcs, nil
}

func (c *Client) GetCustomWorkspaceConfigs(ctx context.Context, object any) error {
	reader, err := c.getWorkspaceConfigsReader(ctx)
	if err != nil {
		return err
	}

	defer func() { _ = reader.Close() }()

	if err = json.NewDecoder(reader).Decode(object); err != nil {
		return fmt.Errorf("failed to decode workspace configs: %w", err)
	}

	return nil
}

func (c *Client) GetUpdatedWorkspaceConfigs(ctx context.Context, updatedAfter time.Time) (*modelv2.WorkspaceConfigs, error) {
	return nil, errors.New("not implemented")
}
