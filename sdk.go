package cpsdk

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rudderlabs/rudder-control-plane-sdk/identity"
	"github.com/rudderlabs/rudder-control-plane-sdk/internal/clients/admin"
	"github.com/rudderlabs/rudder-control-plane-sdk/internal/clients/base"
	"github.com/rudderlabs/rudder-control-plane-sdk/internal/clients/namespace"
	"github.com/rudderlabs/rudder-control-plane-sdk/internal/clients/workspace"
	"github.com/rudderlabs/rudder-control-plane-sdk/modelv2"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

const defaultBaseUrl = "https://api.rudderstack.com"

type ControlPlane struct {
	baseUrl           *url.URL
	workspaceIdentity *identity.Workspace
	namespaceIdentity *identity.Namespace
	adminCredentials  *identity.AdminCredentials

	httpClient  RequestDoer

	Client      Client
	AdminClient *admin.Client

	log logger.Logger
}

type Client interface {
	GetWorkspaceConfigs(ctx context.Context) (*modelv2.WorkspaceConfigs, error)
}

type RequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func New(options ...Option) (*ControlPlane, error) {
	url, _ := url.Parse(defaultBaseUrl)
	cp := &ControlPlane{
		baseUrl: url,
		log:     logger.NOP,
	}

	for _, option := range options {
		if err := option(cp); err != nil {
			return nil, err
		}
	}

	if cp.httpClient == nil {
		cp.httpClient = http.DefaultClient
	}

	baseClient := &base.Client{HTTPClient: cp.httpClient, BaseURL: cp.baseUrl}

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
			Client:   baseClient,
			Identity: cp.namespaceIdentity,
		}
	} else {
		return nil, fmt.Errorf("workspace or namespace identity must be set")
	}

	return cp, nil
}

func (cp *ControlPlane) Close() {
	// TODO: do any cleanup here
}
