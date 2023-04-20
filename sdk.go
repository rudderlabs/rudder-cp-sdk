package cpsdk

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rudderlabs/rudder-control-plane-sdk/identity"
	"github.com/rudderlabs/rudder-control-plane-sdk/internal/clients/admin"
	"github.com/rudderlabs/rudder-control-plane-sdk/internal/clients/base"
	"github.com/rudderlabs/rudder-control-plane-sdk/internal/clients/namespace"
	"github.com/rudderlabs/rudder-control-plane-sdk/internal/clients/workspace"
	"github.com/rudderlabs/rudder-control-plane-sdk/model"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

var defaultBaseUrl = "https://api.rudderstack.com"

type SDK struct {
	baseUrl           string
	workspaceIdentity *identity.Workspace
	namespaceIdentity *identity.Namespace
	adminCredentials  *identity.AdminCredentials

	httpClient  HTTPClient
	Client      Client
	AdminClient *admin.Client

	log logger.Logger
}

type Client interface {
	GetWorkspaceConfigs(ctx context.Context) (*model.WorkspaceConfigs, error)
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func defaultHttpClient() *http.Client {
	return &http.Client{}
}

func New(options ...Option) (*SDK, error) {
	sdk := &SDK{
		baseUrl: defaultBaseUrl,
		log:     logger.NOP,
	}

	for _, option := range options {
		if err := option(sdk); err != nil {
			return nil, err
		}
	}

	if sdk.httpClient == nil {
		sdk.httpClient = defaultHttpClient()
	}

	baseClient := &base.Client{HTTPClient: sdk.httpClient, BaseURL: sdk.baseUrl}

	// set admin client
	if sdk.adminCredentials != nil {
		sdk.AdminClient = &admin.Client{
			Client:   baseClient,
			Username: sdk.adminCredentials.AdminUsername,
			Password: sdk.adminCredentials.AdminPassword,
		}
	}

	// set client based on identity
	if sdk.workspaceIdentity != nil {
		sdk.Client = &workspace.Client{
			Client:   baseClient,
			Identity: sdk.workspaceIdentity,
		}
	} else if sdk.namespaceIdentity != nil {
		sdk.Client = &namespace.Client{
			Client:   baseClient,
			Identity: sdk.namespaceIdentity,
		}
	} else {
		return nil, fmt.Errorf("workspace or namespace identity must be set")
	}

	return sdk, nil
}
