package cpsdk

import (
	"fmt"
	"net/url"

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
