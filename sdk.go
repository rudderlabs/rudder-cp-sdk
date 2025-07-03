package cpsdk

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rudderlabs/rudder-cp-sdk/identity"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/admin"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/base"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/namespace"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/workspace"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

const (
	defaultBaseUrl   = "https://api.rudderstack.com"
	defaultBaseUrlV2 = "https://dp.api.rudderstack.com"
)

type ControlPlane struct {
	Client      Client
	AdminClient *admin.Client

	baseUrl           *url.URL
	baseUrlV2         *url.URL
	workspaceIdentity *identity.Workspace
	namespaceIdentity *identity.Namespace
	adminCredentials  *identity.AdminCredentials

	httpClient RequestDoer
	log        logger.Logger
}

type Client interface {
	GetWorkspaceConfigs(ctx context.Context, object any, updatedAfter time.Time) error
}

type RequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func New(options ...Option) (*ControlPlane, error) {
	baseUrl, _ := url.Parse(defaultBaseUrl)
	baseUrlV2, _ := url.Parse(defaultBaseUrlV2)
	cp := &ControlPlane{
		baseUrl:   baseUrl,
		baseUrlV2: baseUrlV2,
		log:       logger.NOP,
	}

	for _, option := range options {
		if err := option(cp); err != nil {
			return nil, err
		}
	}

	if err := cp.setupClients(); err != nil {
		return nil, err
	}

	return cp, nil
}

func (cp *ControlPlane) setupClients() error {
	if cp.httpClient == nil {
		cp.httpClient = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				DisableKeepAlives:   false,
			},
		}
	}

	baseClient := &base.Client{HTTPClient: cp.httpClient, BaseURL: cp.baseUrl}
	baseClientV2 := &base.Client{HTTPClient: cp.httpClient, BaseURL: cp.baseUrlV2}

	// set admin client
	if cp.adminCredentials != nil {
		cp.AdminClient = &admin.Client{
			Client:   baseClient,
			Username: cp.adminCredentials.AdminUsername,
			Password: cp.adminCredentials.AdminPassword,
		}
	}

	// set client based on identity
	if cp.workspaceIdentity != nil {
		cp.Client = &workspace.Client{
			Client:   baseClient,
			Identity: cp.workspaceIdentity,
		}
	} else if cp.namespaceIdentity != nil {
		cp.Client = &namespace.Client{
			Client:   baseClientV2,
			Identity: cp.namespaceIdentity,
		}
	} else {
		return fmt.Errorf("workspace or namespace identity must be set")
	}

	return nil
}

// GetWorkspaceConfigs returns the latest workspace configs.
func (cp *ControlPlane) GetWorkspaceConfigs(ctx context.Context, object any, updatedAfter time.Time) error {
	return cp.Client.GetWorkspaceConfigs(ctx, object, updatedAfter)
}
