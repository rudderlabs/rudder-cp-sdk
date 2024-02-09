package workspace

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rudderlabs/rudder-cp-sdk/client/internal/base"
	"github.com/rudderlabs/rudder-cp-sdk/model/identity"
)

type getter interface {
	Get(ctx context.Context, endpoint string, headers map[string]string) (int, []byte, error)
}

type Client struct {
	apiCaller getter
	identity  identity.Workspace
}

func New(baseURL string, apiCaller getter, identity identity.Workspace) (*Client, error) {
	c, err := base.New(baseURL, apiCaller)
	if err != nil {
		return nil, err
	}
	return &Client{
		apiCaller: c,
		identity:  identity,
	}, nil
}

func (c *Client) get(ctx context.Context, path string) (int, []byte, error) {
	return c.apiCaller.Get(ctx, path, map[string]string{
		"Content-Type":   "application/json", // @TODO: does it make sense since this is a request and not a response?
		"Authentication": "Basic " + c.identity.WorkspaceToken,
	})
}

func (c *Client) GetWorkspaceConfigs(ctx context.Context) ([]byte, error) {
	sc, data, err := c.get(ctx, "/data-plane/v2/workspaceConfig")
	if err != nil {
		return nil, err
	}

	if sc != http.StatusOK {
		return nil, fmt.Errorf("invalid status code while fetching workspace configs: %d", sc)
	}

	return data, nil
}

func (c *Client) GetUpdatedWorkspaceConfigs(_ context.Context, _ time.Time) ([]byte, error) {
	return nil, errors.New("not implemented")
}
