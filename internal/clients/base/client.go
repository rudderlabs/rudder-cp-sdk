package base

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

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

func (c *Client) Get(ctx context.Context, path string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.Url(path), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Client) Send(req *http.Request) ([]byte, error) {
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
