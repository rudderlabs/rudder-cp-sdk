package cpsdk

import (
	"fmt"
	"net/url"

	"github.com/rudderlabs/rudder-cp-sdk/identity"
	"github.com/rudderlabs/rudder-cp-sdk/tokenfile"
)

type Option func(*ControlPlane) error

var ErrIdentityMutuallyExclusive = fmt.Errorf("cannot set both workspace and namespace identity")

// ErrIdentityTokenWorkspaceIdentity is returned by [New] when [WithIdentityTokenFile] or [WithIdentityToken] is
// combined with a workspace identity.
var ErrIdentityTokenWorkspaceIdentity = fmt.Errorf("identity token is only supported with namespace identity")

// ErrIdentityTokenMutuallyExclusive is returned by [New] when more than one identity token is set, through
// [WithIdentityTokenFile] or [WithIdentityToken].
var ErrIdentityTokenMutuallyExclusive = fmt.Errorf("cannot set both identity token file and identity token")

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

// WithIdentityTokenFile sets the path of a file holding a workload identity token, sent as
// "Authorization: Bearer <token>" instead of the namespace secret on namespace identity requests.
// Set it only once the control plane accepts the token: a refused token gets a 401 with no fallback to the secret.
// [New] reads the file and returns an error if it can't be read or is empty. The file is then read again once the
// token's reuse window has passed, so a rotated token is picked up without restarting; see [tokenfile.Token.Get].
// An empty path sets no token and clears any token set by an earlier option. Combined with a workspace identity, [New] returns [ErrIdentityTokenWorkspaceIdentity],
// and combined with [WithIdentityToken], [ErrIdentityTokenMutuallyExclusive].
func WithIdentityTokenFile(path string) Option {
	if path == "" {
		return WithIdentityToken(nil)
	}
	return WithIdentityToken(tokenfile.New(path))
}

// WithIdentityToken is like [WithIdentityTokenFile] for a [tokenfile.Token] that can be shared with other clients.
// [New] loads it with [tokenfile.Token.Load], so it returns an error if the file can't be read or is empty; the
// caller may have loaded it already. A nil token sets no token and clears any token set by an earlier option.
func WithIdentityToken(t *tokenfile.Token) Option {
	return func(cp *ControlPlane) error {
		if t == nil {
			cp.config.identityToken = nil
			return nil
		}
		if cp.config.identityToken != nil {
			return ErrIdentityTokenMutuallyExclusive
		}
		cp.config.identityToken = t
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
