package namespace

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rudderlabs/rudder-cp-sdk/model/identity"
)

type getter interface {
	Get(ctx context.Context, endpoint string, headers map[string]string) (int, []byte, error)
}

type Client struct {
	apiCaller getter
	identity  identity.Namespace
}

func New(apiCaller getter, identity identity.Namespace) (*Client, error) {
	if apiCaller == nil {
		return nil, errors.New("getter is required")
	}

	return &Client{
		apiCaller: apiCaller,
		identity:  identity,
	}, nil
}

func (c *Client) get(ctx context.Context, path string) (int, []byte, error) {
	return c.apiCaller.Get(ctx, path, map[string]string{
		"Content-Type":   "application/json", // @TODO: does it make sense since this is a request and not a response?
		"Authentication": "Basic " + c.identity.Secret,
	})
}

func (c *Client) GetWorkspaceConfigs(ctx context.Context) ([]byte, error) {
	sc, data, err := c.get(ctx, "/data-plane/v2/namespaces/"+c.identity.Namespace+"/config")
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
