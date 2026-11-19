package cpsdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rudderlabs/rudder-go-kit/httputil"

	"github.com/rudderlabs/rudder-cp-sdk/identity"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/base"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/namespace"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/workspace"
	"github.com/rudderlabs/rudder-cp-sdk/tokenfile"
)

const (
	defaultWorkspaceIdentityBaseURL = "https://api.rudderstack.com"
	defaultNamespaceIdentityBaseURL = "https://dp.api.rudderstack.com"
)

// Secrets determines how account secrets are returned by the workspace configs endpoints.
type Secrets string

const (
	// SecretsEmbed returns account secrets in plain form, embedded in the workspace configs response.
	SecretsEmbed Secrets = "embed"
	// SecretsOmit leaves account secrets out of the workspace configs response. This is the control plane's default.
	SecretsOmit Secrets = "omit"
)

type ControlPlane struct {
	Client
	config struct {
		baseUrl                *url.URL
		workspaceIdentity      *identity.Workspace
		namespaceIdentity      *identity.Namespace
		httpClient             RequestDoer
		secrets                Secrets
		incrementalDefinitions bool
		identityTokenFile      string
		identityToken          *tokenfile.Token
	}
}

type (
	Client interface {
		// GetWorkspaceConfigs decodes the workspace configs in the provided object that were updated after the specified time.
		GetWorkspaceConfigs(ctx context.Context, object any, updatedAfter time.Time) error

		// GetNamespaceWorkspaces returns the list of workspace IDs under the namespace.
		// This is only applicable for Namespace Identity and will return an error for Workspace Identity.
		GetNamespaceWorkspaces(ctx context.Context) (workspaceIDs []string, err error)
	}
)

type RequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func New(options ...Option) (*ControlPlane, error) {
	cp := &ControlPlane{}
	for _, option := range options {
		if err := option(cp); err != nil {
			return nil, err
		}
	}
	if (cp.config.identityTokenFile != "" || cp.config.identityToken != nil) && cp.config.workspaceIdentity != nil {
		return nil, ErrIdentityTokenWorkspaceIdentity
	}
	if cp.config.identityTokenFile != "" {
		if cp.config.identityToken != nil {
			return nil, ErrIdentityTokenMutuallyExclusive
		}
		cp.config.identityToken = tokenfile.New(cp.config.identityTokenFile)
		if err := cp.config.identityToken.Load(); err != nil {
			return nil, fmt.Errorf("loading identity token: %w", err)
		}
	}
	if cp.config.httpClient == nil {
		client := httputil.DefaultHttpClient()
		client.CheckRedirect = keepAuthorizationOnOrigin
		cp.config.httpClient = client
	}
	// set client based on identity
	if cp.config.workspaceIdentity != nil {
		baseUrl := cp.config.baseUrl
		if baseUrl == nil {
			baseUrl, _ = url.Parse(defaultWorkspaceIdentityBaseURL)
		}
		cp.Client = &workspace.Client{
			Client:   &base.Client{HTTPClient: cp.config.httpClient, BaseURL: baseUrl, IncrementalDefinitions: cp.config.incrementalDefinitions},
			Identity: cp.config.workspaceIdentity,
			Secrets:  string(cp.config.secrets),
		}
	} else if cp.config.namespaceIdentity != nil {
		baseUrl := cp.config.baseUrl
		if baseUrl == nil {
			baseUrl, _ = url.Parse(defaultNamespaceIdentityBaseURL)
		}
		cp.Client = &namespace.Client{
			Client:        &base.Client{HTTPClient: cp.config.httpClient, BaseURL: baseUrl, IncrementalDefinitions: cp.config.incrementalDefinitions},
			Identity:      cp.config.namespaceIdentity,
			Secrets:       string(cp.config.secrets),
			IdentityToken: cp.config.identityToken,
		}
	} else {
		return nil, fmt.Errorf("workspace or namespace identity must be set")
	}
	return cp, nil
}

// keepAuthorizationOnOrigin follows up to 10 redirects like net/http's default policy and removes the Authorization
// header when a redirect leaves the scheme, host or port of the original request, where net/http only removes it for
// a different host.
func keepAuthorizationOnOrigin(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	to, from := req.URL, via[0].URL
	if !strings.EqualFold(to.Scheme, from.Scheme) || !strings.EqualFold(to.Hostname(), from.Hostname()) || effectivePort(to) != effectivePort(from) {
		req.Header.Del("Authorization")
	}
	return nil
}

// effectivePort returns the URL's port, or the default port of its scheme when it has none.
func effectivePort(u *url.URL) string {
	if port := u.Port(); port != "" {
		return port
	}
	switch strings.ToLower(u.Scheme) {
	case "http":
		return "80"
	case "https":
		return "443"
	}
	return ""
}
