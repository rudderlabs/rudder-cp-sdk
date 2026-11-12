package cpsdk

import (
	"fmt"
	"net/url"

	"github.com/rudderlabs/rudder-cp-sdk/identity"
)

type Option func(*ControlPlane) error

var ErrIdentityMutuallyExclusive = fmt.Errorf("cannot set both workspace and namespace identity")

func WithWorkspaceIdentity(workspaceToken string) Option {
	return func(cp *ControlPlane) error {
		if cp.config.namespaceIdentity != nil {
			return ErrIdentityMutuallyExclusive
		}
		cp.config.workspaceIdentity = &identity.Workspace{WorkspaceToken: workspaceToken}
		return nil
	}
}

func WithNamespaceIdentity(namespace, secret string) Option {
	return func(cp *ControlPlane) error {
		if cp.config.workspaceIdentity != nil {
			return ErrIdentityMutuallyExclusive
		}

		cp.config.namespaceIdentity = &identity.Namespace{Namespace: namespace, Secret: secret}
		return nil
	}
}

func WithBaseUrl(baseUrl string) Option {
	return func(cp *ControlPlane) error {
		u, err := url.Parse(baseUrl)
		if err != nil {
			return fmt.Errorf("invalid base url: %w", err)
		}
		cp.config.baseUrl = u
		return nil
	}
}

// WithSecrets controls whether account secrets are embedded in workspace configs responses.
// If not set, the control plane's default applies, which is [SecretsOmit].
func WithSecrets(s Secrets) Option {
	return func(cp *ControlPlane) error {
		switch s {
		case SecretsEmbed, SecretsOmit:
			cp.config.secrets = s
			return nil
		default:
			return fmt.Errorf("invalid secrets option: %q", s)
		}
	}
}

// WithIdentityTokenFile sets the path of a file holding a workload identity token.
// The file is read on every namespace identity request, so a rotated token is picked up without restarting,
// and its trimmed content is sent in the X-Dataplane-Identity header alongside the existing credentials.
// If the file is missing, unreadable or empty, the request is sent without the header.
// It has no effect with a workspace identity.
// The default HTTP client drops the header on redirects to a different scheme or host;
// a RequestDoer set with [WithRequestDoer] applies its own redirect policy.
func WithIdentityTokenFile(path string) Option {
	return func(cp *ControlPlane) error {
		cp.config.identityTokenFile = path
		return nil
	}
}

func WithRequestDoer(reqDoer RequestDoer) Option {
	return func(cp *ControlPlane) error {
		cp.config.httpClient = reqDoer
		return nil
	}
}

// WithIncrementalDefinitions enables requesting incremental definitions, i.e. definitions omitted when unchanged.
// It only applies to objects passed to GetWorkspaceConfigs that implement [diff.IncrementalDefinitionsReceiver]
// and, if they are [diff.UpdateableObject]s, whose non updateable lists all implement [diff.OmittableList].
// Otherwise GetWorkspaceConfigs returns a permanent error without sending the request.
// Only the namespace endpoint confirms incremental definitions, elsewhere responses stay legacy.
func WithIncrementalDefinitions(enabled bool) Option {
	return func(cp *ControlPlane) error {
		cp.config.incrementalDefinitions = enabled
		return nil
	}
}
