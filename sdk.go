package cpsdk

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rudderlabs/rudder-cp-sdk/identity"
	"github.com/rudderlabs/rudder-cp-sdk/internal/cache"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/admin"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/base"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/namespace"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/workspace"
	"github.com/rudderlabs/rudder-cp-sdk/internal/poller"
	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
	"github.com/rudderlabs/rudder-cp-sdk/notifications"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

const defaultBaseUrl = "https://api.rudderstack.com"

type ControlPlane struct {
	baseUrl           *url.URL
	workspaceIdentity *identity.Workspace
	namespaceIdentity *identity.Namespace
	adminCredentials  *identity.AdminCredentials

	httpClient RequestDoer

	Client      Client
	AdminClient *admin.Client

	log logger.Logger

	configsCache *cache.WorkspaceConfigCache

	pollingInterval time.Duration
	poller          *poller.Poller
	pollerStop      context.CancelFunc
}

type Client interface {
	GetWorkspaceConfigs(ctx context.Context) (*modelv2.WorkspaceConfigs, error)
	GetUpdatedWorkspaceConfigs(ctx context.Context, updatedAfter time.Time) (*modelv2.WorkspaceConfigs, error)
}

type RequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func New(options ...Option) (*ControlPlane, error) {
	url, _ := url.Parse(defaultBaseUrl)
	cp := &ControlPlane{
		baseUrl:         url,
		log:             logger.NOP,
		pollingInterval: 1 * time.Second,
		configsCache:    &cache.WorkspaceConfigCache{},
	}

	for _, option := range options {
		if err := option(cp); err != nil {
			return nil, err
		}
	}

	if err := cp.setupClients(); err != nil {
		return nil, err
	}

	if err := cp.setupPoller(); err != nil {
		return nil, err
	}

	return cp, nil
}

func (cp *ControlPlane) setupClients() error {
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
		return fmt.Errorf("workspace or namespace identity must be set")
	}

	return nil
}

func (cp *ControlPlane) setupPoller() error {
	if cp.pollingInterval == 0 {
		return nil
	}

	handle := func(wc *modelv2.WorkspaceConfigs) error {
		cp.configsCache.Set(wc)
		return nil
	}

	p, err := poller.New(handle,
		poller.WithClient(cp.Client),
		poller.WithPollingInterval(cp.pollingInterval),
		poller.WithLogger(cp.log))
	if err != nil {
		return err
	}

	cp.poller = p
	ctx, cancel := context.WithCancel(context.Background())
	cp.pollerStop = cancel
	cp.poller.Start(ctx)

	return nil
}

// Close stops any background processes such as polling for workspace configs.
func (cp *ControlPlane) Close() {
	if cp.poller != nil {
		cp.pollerStop()
	}
}

// GetWorkspaceConfigs returns the latest workspace configs.
// If polling is enabled, this will return the cached configs, if they have been retrieved at least once.
// Otherwise, it will make a request to the control plane to get the latest configs.
func (cp *ControlPlane) GetWorkspaceConfigs() (*modelv2.WorkspaceConfigs, error) {
	if cp.poller != nil {
		return cp.configsCache.Get(), nil
	} else {
		return cp.Client.GetWorkspaceConfigs(context.Background())
	}
}

type Subscriber interface {
	Notifications() chan notifications.WorkspaceConfigNotification
}

func (cp *ControlPlane) Subscribe() chan notifications.WorkspaceConfigNotification {
	return cp.configsCache.Subscribe()
}
