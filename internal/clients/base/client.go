package base

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/valyala/fasthttp"
)

type Client struct {
	HTTPClient *fasthttp.Client
	BaseURL    *url.URL
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (c *Client) Url(path string) string {
	return c.BaseURL.JoinPath(path).String()
}

func (c *Client) Get(ctx context.Context, path string) (*fasthttp.Request, error) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(c.Url(path))
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *Client) Send(req *fasthttp.Request) ([]byte, error) {
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := c.HTTPClient.Do(req, resp)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	return resp.Body(), nil

	//res, err := c.HTTPClient.Do(req)
	//if err != nil {
	//	return nil, err
	//}
	//defer res.Body.Close()
	//
	//if res.StatusCode != http.StatusOK {
	//	return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	//}
	//
	//data, err := io.ReadAll(res.Body)
	//if err != nil {
	//	return nil, err
	//}
	//
	//return data, nil
}
