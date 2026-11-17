package namespace

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/jsonrs"

	"github.com/rudderlabs/rudder-cp-sdk/identity"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/base"
)

type Client struct {
	*base.Client

	Identity *identity.Namespace
	// Secrets controls whether account secrets are embedded in workspace configs responses.
	// If empty, the control plane's default applies.
	Secrets string
	// IdentityTokenFile is the path of a file holding a workload identity token, sent as a Bearer token
	// instead of the namespace secret. If empty, the namespace secret is sent.
	IdentityTokenFile string

	identityToken cachedToken
}

func (c *Client) GetWorkspaceConfigs(ctx context.Context, object any, updatedAfter time.Time) error {
	req, err := c.GetWithAuth(ctx, "/configuration/v2/namespaces/"+c.Identity.Namespace,
		base.WithUpdatedAfter(updatedAfter), base.WithSecrets(c.Secrets))
	if err != nil {
		return err
	}
	return c.SendWorkspaceConfigs(req, object)
}

func (c *Client) GetNamespaceWorkspaces(ctx context.Context) ([]string, error) {
	type response struct {
		Data []string `json:"data"`
	}
	req, err := c.GetWithAuth(ctx, "/configuration/v2/namespaces/"+c.Identity.Namespace+"/workspace-ids")
	if err != nil {
		return nil, fmt.Errorf("creating get request: %w", err)
	}
	resp, err := c.Send(req)
	if err != nil {
		return nil, fmt.Errorf("sending get request: %w", err)
	}
	defer func() { httputil.CloseResponse(resp) }()

	var res response
	if err = jsonrs.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decoding namespace workspaces response: %w", err)
	}
	return res.Data, nil
}

func (c *Client) GetWithAuth(ctx context.Context, path string, queryOpts ...base.QueryOption) (*http.Request, error) {
	req, err := c.Get(ctx, path, queryOpts...)
	if err != nil {
		return nil, err
	}
	if token := c.identityToken.get(c.IdentityTokenFile); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else {
		req.SetBasicAuth(c.Identity.Secret, "")
	}
	return req, nil
}

// identityTokenMaxAge is how long a token read from the file is reused. The kubelet rewrites the file at about 80%
// of the token's lifetime (8 of 10 minutes by default), so a token read under a minute ago is still valid.
const identityTokenMaxAge = time.Minute

// cachedToken reads a token file at most once per identityTokenMaxAge.
type cachedToken struct {
	mu     sync.Mutex
	token  string
	readAt time.Time
	now    func() time.Time // nil means time.Now; set by tests
}

// get returns the token in path. An unset path, or no successful read yet, returns an empty token. A failed or
// empty read keeps the last good token, which the kubelet only replaces with a newer one.
func (t *cachedToken) get(path string) string {
	if path == "" {
		return ""
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	if t.now != nil {
		now = t.now()
	}
	if t.token != "" && now.Sub(t.readAt) < identityTokenMaxAge {
		return t.token
	}
	if data, err := os.ReadFile(path); err == nil {
		if token := strings.TrimSpace(string(data)); token != "" {
			t.token, t.readAt = token, now
		}
	}
	return t.token
}
