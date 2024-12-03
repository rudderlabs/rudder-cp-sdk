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

const (
	defaultBaseUrl   = "https://api.rudderstack.com"
	defaultBaseUrlV2 = "https://dp.api.rudderstack.com"
)

/*
TODO think of an abstracted way to handle diffs from updatedAfter even though we might not know
the object type. we could get functions that will tell us whether a workspace exists or not or a function to
set it/update it.
*/

type ControlPlane struct {
	baseUrl           *url.URL
	baseUrlV2         *url.URL
	workspaceIdentity *identity.Workspace
	namespaceIdentity *identity.Namespace
	adminCredentials  *identity.AdminCredentials

	httpClient RequestDoer

	Client      Client
	AdminClient *admin.Client

	log logger.Logger

	configsCache *cache.WorkspaceConfigCache

	poller       *poller.Poller
	pollerConfig *pollerConfig
	pollerStop   context.CancelFunc
	pollerDone   chan struct{}
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
		baseUrl:      baseUrl,
		baseUrlV2:    baseUrlV2,
		log:          logger.NOP,
		configsCache: &cache.WorkspaceConfigCache{},
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

func (cp *ControlPlane) setupPoller() error {
	if cp.pollerConfig == nil {
		return nil
	}

	handle := func(wc *modelv2.WorkspaceConfigs) error {
		cp.configsCache.Set(wc)
		return nil
	}

	p, err := poller.New(handle,
		poller.WithClient(cp.Client),
		poller.WithPollingInterval(cp.pollerConfig.pollingInterval),
		poller.WithPollingBackoffInitialInterval(cp.pollerConfig.backoffInitialInterval),
		poller.WithPollingBackoffMaxInterval(cp.pollerConfig.backoffMaxInterval),
		poller.WithPollingBackoffMultiplier(cp.pollerConfig.backoffMultiplier),
		poller.WithLogger(cp.log),
	)
	if err != nil {
		return err
	}

	cp.poller = p
	ctx, cancel := context.WithCancel(context.Background())
	cp.pollerStop = cancel
	cp.pollerDone = make(chan struct{})
	go func() {
		cp.poller.Run(ctx)
		close(cp.pollerDone)
	}()

	return nil
}

// Close stops any background processes such as polling for workspace configs.
func (cp *ControlPlane) Close(ctx context.Context) error {
	if cp.pollerDone == nil {
		return nil
	}

	cp.pollerStop()

	select {
	case <-cp.pollerDone:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetWorkspaceConfigs returns the latest workspace configs.
// If polling is enabled, this will return the cached configs, if they have been retrieved at least once.
// Otherwise, it will make a request to the control plane to get the latest configs.
func (cp *ControlPlane) GetWorkspaceConfigs(ctx context.Context, object any, updatedAfter time.Time) error {
	if cp.poller != nil {
		object = cp.configsCache.Get() //nolint:ineffassign,wastedassign
		return nil
	} else {
		return cp.Client.GetWorkspaceConfigs(ctx, object, updatedAfter)
	}
}

type Subscriber interface {
	Notifications() chan notifications.WorkspaceConfigNotification
}

func (cp *ControlPlane) Subscribe() chan notifications.WorkspaceConfigNotification {
	return cp.configsCache.Subscribe()
}
