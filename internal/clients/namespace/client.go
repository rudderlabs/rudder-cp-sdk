package namespace

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/jsonrs"

	"github.com/rudderlabs/rudder-cp-sdk/identity"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/base"
)

type Client struct {
	*base.Client

	Identity *identity.Namespace
	// Secrets controls whether account secrets are embedded in workspace configs responses.
	// If empty, the control plane's default applies.
	Secrets string
}

func (c *Client) GetWorkspaceConfigs(ctx context.Context, object any, updatedAfter time.Time) error {
	req, err := c.GetWithAuth(ctx, "/configuration/v2/namespaces/"+c.Identity.Namespace,
		base.WithUpdatedAfter(updatedAfter), base.WithSecrets(c.Secrets))
	if err != nil {
		return err
	}
	return c.SendWorkspaceConfigs(req, object)
}

func (c *Client) GetNamespaceWorkspaces(ctx context.Context) ([]string, error) {
	type response struct {
		Data []string `json:"data"`
	}
	req, err := c.GetWithAuth(ctx, "/configuration/v2/namespaces/"+c.Identity.Namespace+"/workspace-ids")
	if err != nil {
		return nil, fmt.Errorf("creating get request: %w", err)
	}
	resp, err := c.Send(req)
	if err != nil {
		return nil, fmt.Errorf("sending get request: %w", err)
	}
	defer func() { httputil.CloseResponse(resp) }()

	var res response
	if err = jsonrs.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decoding namespace workspaces response: %w", err)
	}
	return res.Data, nil
}

func (c *Client) GetWithAuth(ctx context.Context, path string, queryOpts ...base.QueryOption) (*http.Request, error) {
	req, err := c.Get(ctx, path, queryOpts...)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Identity.Secret, "")
	return req, nil
}
