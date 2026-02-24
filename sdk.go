package cpsdk

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rudderlabs/rudder-go-kit/httputil"

	"github.com/rudderlabs/rudder-cp-sdk/identity"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/base"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/namespace"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/workspace"
)

const (
	defaultWorkspaceIdentityBaseURL = "https://api.rudderstack.com"
	defaultNamespaceIdentityBaseURL = "https://dp.api.rudderstack.com"
)

type ControlPlane struct {
	Client
	config struct {
		baseUrl           *url.URL
		workspaceIdentity *identity.Workspace
		namespaceIdentity *identity.Namespace
		httpClient        RequestDoer
	}
}

type Client interface {
	GetWorkspaceConfigs(ctx context.Context, object any, updatedAfter time.Time) error
}

type RequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func New(options ...Option) (*ControlPlane, error) {
	cp := &ControlPlane{}
	for _, option := range options {
		if err := option(cp); err != nil {
			return nil, err
		}
	}
	if cp.config.httpClient == nil {
		cp.config.httpClient = httputil.DefaultHttpClient()
	}
	// set client based on identity
	if cp.config.workspaceIdentity != nil {
		baseUrl := cp.config.baseUrl
		if baseUrl == nil {
			baseUrl, _ = url.Parse(defaultWorkspaceIdentityBaseURL)
		}
		cp.Client = &workspace.Client{
			Client:   &base.Client{HTTPClient: cp.config.httpClient, BaseURL: baseUrl},
			Identity: cp.config.workspaceIdentity,
		}
	} else if cp.config.namespaceIdentity != nil {
		baseUrl := cp.config.baseUrl
		if baseUrl == nil {
			baseUrl, _ = url.Parse(defaultNamespaceIdentityBaseURL)
		}
		cp.Client = &namespace.Client{
			Client:   &base.Client{HTTPClient: cp.config.httpClient, BaseURL: baseUrl},
			Identity: cp.config.namespaceIdentity,
		}
	} else {
		return nil, fmt.Errorf("workspace or namespace identity must be set")
	}
	return cp, nil
}
