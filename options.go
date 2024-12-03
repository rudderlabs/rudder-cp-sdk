package cpsdk

import (
	"fmt"
	"net/url"
	"time"

	"github.com/rudderlabs/rudder-go-kit/logger"

	"github.com/rudderlabs/rudder-cp-sdk/identity"
)

type Option func(*ControlPlane) error

var ErrIdentityMutuallyExclusive = fmt.Errorf("cannot set both workspace and namespace identity")

func WithWorkspaceIdentity(workspaceToken string) Option {
	return func(cp *ControlPlane) error {
		if cp.namespaceIdentity != nil {
			return ErrIdentityMutuallyExclusive
		}

		cp.workspaceIdentity = &identity.Workspace{WorkspaceToken: workspaceToken}
		return nil
	}
}

func WithNamespaceIdentity(namespace, secret string) Option {
	return func(cp *ControlPlane) error {
		if cp.workspaceIdentity != nil {
			return ErrIdentityMutuallyExclusive
		}

		cp.namespaceIdentity = &identity.Namespace{Namespace: namespace, Secret: secret}
		return nil
	}
}

func WithBaseUrl(baseUrl string) Option {
	return func(cp *ControlPlane) error {
		u, err := url.Parse(baseUrl)
		if err != nil {
			return fmt.Errorf("invalid base url: %w", err)
		}

		cp.baseUrl = u
		cp.baseUrlV2 = u

		return nil
	}
}

func WithAdminCredentials(credentials *identity.AdminCredentials) Option {
	return func(cp *ControlPlane) error {
		cp.adminCredentials = credentials
		return nil
	}
}

func WithLogger(log logger.Logger) Option {
	return func(cp *ControlPlane) error {
		cp.log = log
		return nil
	}
}

// PollerOption is used to configure the poller
type PollerOption func(*pollerConfig)

// WithPoller allows to configure the poller
// When invoked without arguments, it will use the default values set inside this function
func WithPoller(options ...PollerOption) Option {
	return func(cp *ControlPlane) error {
		cp.pollerConfig = &pollerConfig{
			pollingInterval:        1 * time.Second,
			backoffInitialInterval: 1 * time.Second,
			backoffMaxInterval:     1 * time.Minute,
			backoffMultiplier:      1.5,
		}
		for _, option := range options {
			option(cp.pollerConfig)
		}
		return nil
	}
}

type pollerConfig struct {
	pollingInterval        time.Duration
	backoffInitialInterval time.Duration
	backoffMaxInterval     time.Duration
	backoffMultiplier      float64
}

func WithPollingInterval(interval time.Duration) PollerOption {
	return func(pc *pollerConfig) { pc.pollingInterval = interval }
}

func WithPollingBackoffInitialInterval(interval time.Duration) PollerOption {
	return func(pc *pollerConfig) { pc.backoffInitialInterval = interval }
}

func WithPollingBackoffMaxInterval(interval time.Duration) PollerOption {
	return func(pc *pollerConfig) { pc.backoffMaxInterval = interval }
}

func WithPollingBackoffMultiplier(multiplier float64) PollerOption {
	return func(pc *pollerConfig) { pc.backoffMultiplier = multiplier }
}
