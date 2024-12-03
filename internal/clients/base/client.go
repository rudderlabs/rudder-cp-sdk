package base

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

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

func (c *Client) Url(path string, updatedAfter time.Time) string {
	v := c.BaseURL.JoinPath(path)

	if !updatedAfter.IsZero() {
		queryValues := v.Query()
		queryValues.Add("updatedAfter", updatedAfter.Format(updatedAfterTimeFormat))
		v.RawQuery = queryValues.Encode()
	}

	return v.String()
}

func (c *Client) Get(ctx context.Context, path string, updatedAfter time.Time) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.Url(path, updatedAfter), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	return req, nil
}

func (c *Client) Send(req *http.Request) (io.ReadCloser, error) {
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		defer func() { httputil.CloseResponse(res) }()
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	if res.Header.Get("Content-Encoding") == "gzip" {
		gzipRdr, err := gzip.NewReader(res.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		return &gzipReadCloser{
			gzipReader: gzipRdr,
			resBody:    res.Body,
		}, nil
	}

	return res.Body, nil
}

// gzipReadCloser is a ReadCloser that closes both the gzip reader and the response body.
// unfortunately the gzip.Reader does not close the underlying reader when it is closed.
type gzipReadCloser struct {
	gzipReader io.ReadCloser
	resBody    io.ReadCloser
}

func (g *gzipReadCloser) Read(p []byte) (n int, err error) {
	return g.gzipReader.Read(p)
}

func (g *gzipReadCloser) Close() error {
	defer func() { _ = g.resBody.Close() }()
	return g.gzipReader.Close()
}
