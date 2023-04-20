package cpsdk

import (
	"fmt"

	"github.com/rudderlabs/rudder-control-plane-sdk/identity"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

type Option func(*SDK) error

var ErrIdentityAlreadySet = fmt.Errorf("cannot set both workspace and namespace identity")

func WithWorkspaceIdentity(workspaceToken string) Option {
	return func(sdk *SDK) error {
		if sdk.namespaceIdentity != nil {
			return ErrIdentityAlreadySet
		}

		sdk.workspaceIdentity = &identity.Workspace{WorkspaceToken: workspaceToken}
		return nil
	}
}

func WithNamespaceIdentity(namespace, secret string) Option {
	return func(sdk *SDK) error {
		if sdk.workspaceIdentity != nil {
			return ErrIdentityAlreadySet
		}

		sdk.namespaceIdentity = &identity.Namespace{Namespace: namespace, Secret: secret}
		return nil
	}
}

func WithBaseUrl(baseUrl string) Option {
	return func(sdk *SDK) error {
		sdk.baseUrl = baseUrl
		return nil
	}
}

func WithAdminCredentials(credentials *identity.AdminCredentials) Option {
	return func(sdk *SDK) error {
		sdk.adminCredentials = credentials
		return nil
	}
}

func WithLogger(log logger.Logger) Option {
	return func(sdk *SDK) error {
		sdk.log = log
		return nil
	}
}
