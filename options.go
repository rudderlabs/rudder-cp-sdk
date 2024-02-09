package cpsdk

import (
	"fmt"

	"github.com/rudderlabs/rudder-cp-sdk/model/identity"
	"github.com/rudderlabs/rudder-go-kit/logger"
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
		cp.baseUrl = baseUrl
		return nil
	}
}

func WithLogger(log logger.Logger) Option {
	return func(cp *ControlPlane) error {
		cp.log = log
		return nil
	}
}

func WithRequestDoer(requestDoer RequestDoer) Option {
	return func(cp *ControlPlane) error {
		cp.requestDoer = requestDoer
		return nil
	}
}

// WithPollingInterval sets the interval at which the SDK polls for new configs.
// If not set, the SDK will poll every 1 second.
// If set to 0, the SDK will not poll for new configs.
//func WithPollingInterval(interval time.Duration) Option {
//	return func(cp *ControlPlane) error {
//		cp.pollingInterval = interval
//		return nil
//	}
//}
