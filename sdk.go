package cpsdk

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rudderlabs/rudder-cp-sdk/client/namespace"
	"github.com/rudderlabs/rudder-cp-sdk/client/workspace"
	whttp "github.com/rudderlabs/rudder-cp-sdk/client/wrapper/http"
	"github.com/rudderlabs/rudder-cp-sdk/internal/poller"
	"github.com/rudderlabs/rudder-cp-sdk/model/identity"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

const defaultBaseUrl = "https://api.rudderstack.com"

type RequestDoer interface {
	Get(_ context.Context, endpoint string, headers map[string]string) (int, []byte, error)
}

type Client interface {
	GetWorkspaceConfigs(ctx context.Context) ([]byte, error)
	GetUpdatedWorkspaceConfigs(ctx context.Context, updatedAfter time.Time) ([]byte, error)
}

type ControlPlane struct {
	baseUrl           string
	workspaceIdentity *identity.Workspace
	namespaceIdentity *identity.Namespace
	adminCredentials  *identity.AdminCredentials

	client      Client
	requestDoer RequestDoer

	pollingInterval time.Duration
	poller          *poller.Poller
	pollerStop      func()

	log logger.Logger
}

func New(options ...Option) (*ControlPlane, error) {
	cp := &ControlPlane{
		log:        logger.NOP,
		pollerStop: func() {},
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
	if cp.requestDoer == nil {
		cp.requestDoer = whttp.New(http.DefaultClient)
	}

	baseURL := cp.baseUrl
	if baseURL == "" {
		baseURL = defaultBaseUrl
	}

	// set client based on identity
	var err error
	if cp.workspaceIdentity != nil {
		cp.client, err = workspace.New(baseURL, cp.requestDoer, *cp.workspaceIdentity)
	} else if cp.namespaceIdentity != nil {
		cp.client, err = namespace.New(baseURL, cp.requestDoer, *cp.namespaceIdentity)
	} else {
		return fmt.Errorf("workspace or namespace identity must be set")
	}
	return err
}

func (cp *ControlPlane) setupPoller() error {
	if cp.pollingInterval == 0 {
		return nil
	}

	caller := func(ctx context.Context) ([]byte, error) {
		return cp.client.GetWorkspaceConfigs(ctx)
	}

	handler := func(data []byte, err error) {
		// @TODO send to subscribers here
		// @TODO if cache is enabled then also do the parsing
		// @TODO pass favourite parser via options
	}

	p, err := poller.New(caller, handler,
		poller.WithPollingInterval(cp.pollingInterval),
	)
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
	cp.pollerStop()
}

// GetWorkspaceConfigs returns the latest workspace configs.
// If polling is enabled, this will return the cached configs, if they have been retrieved at least once.
// Otherwise, it will make a request to the control plane to get the latest configs.
func (cp *ControlPlane) GetWorkspaceConfigs(ctx context.Context) ([]byte, error) {
	// @TODO add GetWorkspaceConfigs and GetRawWorkspaceConfigs
	// @TODO the non-raw version uses the cache, the other one does not
	return cp.client.GetWorkspaceConfigs(ctx)
}

// Subscribe @TODO SubscribeRaw?
// @TODO rename to SubscribeToWorkspaceConfigs
func (cp *ControlPlane) Subscribe() chan []byte {
	// @TODO: someone subscribes and we don't have a poller, create one
	// return cp.configsCache.Subscribe(
	return nil // @TODO implement
}
