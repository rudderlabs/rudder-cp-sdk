package fasthttp

import (
	"context"
	"time"

	"github.com/valyala/fasthttp"
)

type RequestDoer struct {
	client *fasthttp.Client
}

func (a *RequestDoer) Get(_ context.Context, endpoint string, headers map[string]string) (int, []byte, error) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(endpoint)
	req.Header.SetMethod(fasthttp.MethodGet)
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp := fasthttp.AcquireResponse()
	err := a.client.Do(req, resp)
	fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode(), resp.Body(), nil
}

// New @TODO add options
func New() *RequestDoer {
	readTimeout, _ := time.ParseDuration("500ms")
	writeTimeout, _ := time.ParseDuration("500ms")
	maxIdleConnDuration, _ := time.ParseDuration("1h")
	client := &fasthttp.Client{
		ReadTimeout:                   readTimeout,
		WriteTimeout:                  writeTimeout,
		MaxIdleConnDuration:           maxIdleConnDuration,
		NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
		DisablePathNormalizing:        true,
		// increase DNS cache time to an hour instead of default minute
		MaxConnsPerHost: 5000,
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      0, // @TODO add option
			DNSCacheDuration: time.Hour,
		}).Dial,
	}

	return &RequestDoer{
		client: client,
	}
}
