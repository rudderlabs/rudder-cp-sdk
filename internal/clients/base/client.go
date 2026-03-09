package base

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/rudderlabs/rudder-go-kit/httputil"
)

const updatedAfterTimeFormat = "2006-01-02T15:04:05.000Z"

type Client struct {
	HTTPClient HTTPClient
	BaseURL    *url.URL
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (c *Client) Url(path string) string {
	return c.BaseURL.JoinPath(path).String()
}

func (c *Client) Get(ctx context.Context, path string, queryOpts ...QueryOption) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.Url(path), http.NoBody)
	if err != nil {
		return nil, &PermanenentError{Err: fmt.Errorf("creating request: %w", err)}
	}
	if len(queryOpts) > 0 {
		q := req.URL.Query()
		for _, opt := range queryOpts {
			opt(q)
		}
		req.URL.RawQuery = q.Encode()
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// Send the request and return the response body if the status code is 200 OK, otherwise return an error.
func (c *Client) Send(req *http.Request) (io.ReadCloser, error) {
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		defer func() { httputil.CloseResponse(res) }()
		return nil, NewUnexpectedStatusCodeError(res)
	}
	return res.Body, nil
}
