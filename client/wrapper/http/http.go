package fasthttp

import (
	"context"
	"io"
	"net/http"
)

type RequestDoer struct {
	client *http.Client
}

func (a *RequestDoer) Get(ctx context.Context, endpoint string, headers map[string]string) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, nil, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return 0, nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, body, nil
}

// New @TODO add options
func New(client *http.Client) *RequestDoer {
	return &RequestDoer{
		client: client,
	}
}
