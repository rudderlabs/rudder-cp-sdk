package base

import (
	"context"
	"fmt"
	"net/url"
)

type getter interface {
	Get(ctx context.Context, endpoint string, headers map[string]string) (int, []byte, error)
}

type Client struct {
	baseURL   string
	apiCaller getter
}

func New(baseURL string, apiCaller getter) (*Client, error) {
	if err := isValidURL(baseURL); err != nil {
		return nil, fmt.Errorf("invalid baseURL: %w", err)
	}
	if apiCaller == nil {
		return nil, fmt.Errorf("getter is required")
	}
	return &Client{
		baseURL:   baseURL,
		apiCaller: apiCaller,
	}, nil
}

func (c *Client) Get(ctx context.Context, path string, headers map[string]string) (int, []byte, error) {
	endpoint, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return 0, nil, err
	}
	return c.apiCaller.Get(ctx, endpoint, headers)
}

func isValidURL(s string) error {
	if s == "" {
		return fmt.Errorf("url is empty")
	}

	parsedURL, err := url.ParseRequestURI(s)
	if err != nil {
		return err
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("invalid scheme: %s or host: %s", parsedURL.Scheme, parsedURL.Host)
	}

	return nil
}
