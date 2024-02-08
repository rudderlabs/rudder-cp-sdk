package namespace

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/rudderlabs/rudder-cp-sdk/identity"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/base"
	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
	"github.com/rudderlabs/rudder-cp-sdk/modelv2/parser"
)

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

func (c *Client) GetWorkspaceConfigs(ctx context.Context) (*modelv2.WorkspaceConfigs, error) {
	req, err := c.Get(ctx, "/data-plane/v2/namespaces/"+c.Identity.Namespace+"/config")
	if err != nil {
		return nil, err
	}

	data, err := c.Send(req)
	if err != nil {
		return nil, err
	}

	wcs, err := parser.Parse(data)
	if err != nil {
		return nil, err
	}

	return wcs, nil
}

func (c *Client) GetUpdatedWorkspaceConfigs(ctx context.Context, updatedAfter time.Time) (*modelv2.WorkspaceConfigs, error) {
	u := url.URL{}
	u.Path = "/data-plane/v2/namespaces/" + c.Identity.Namespace + "/config"

	values := u.Query()
	values.Add("updatedAfter", updatedAfter.Format(base.UpdatedAfterTimeFormat))
	u.RawQuery = values.Encode()

	req, err := c.Get(ctx, u.String())
	if err != nil {
		return nil, err
	}

	data, err := c.Send(req)
	if err != nil {
		return nil, err
	}

	wcs, err := parser.Parse(data)
	if err != nil {
		return nil, err
	}

	return wcs, nil
}
