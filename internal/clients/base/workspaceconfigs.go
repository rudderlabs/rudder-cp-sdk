package base

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/jsonrs"

	"github.com/rudderlabs/rudder-cp-sdk/diff"
)

const (
	incrementalDefinitionsHeader = "X-Rudder-Incremental-Definitions"
	configUpdatedAtHeader        = "X-Rudder-Config-Updated-At"
)

// SendWorkspaceConfigs sends a workspace configs request and decodes the response into object.
// If object is a [diff.IncrementalDefinitionsReceiver], it is also told whether the control plane confirmed
// incremental definitions and the cursor to use for the next request.
func (c *Client) SendWorkspaceConfigs(req *http.Request, object any) error {
	if c.IncrementalDefinitions {
		if !supportsIncrementalDefinitions(object) {
			return &PermanenentError{Err: fmt.Errorf("incremental definitions are not supported by %T", object)}
		}
		req.Header.Set(incrementalDefinitionsHeader, "true")
	}

	res, err := c.Send(req)
	if err != nil {
		return err
	}
	defer func() { httputil.CloseResponse(res) }()

	confirmed := c.IncrementalDefinitions && res.Header.Get(incrementalDefinitionsHeader) == "true"
	var cursor time.Time
	if confirmed {
		raw := res.Header.Get(configUpdatedAtHeader)
		if cursor, err = time.Parse(time.RFC3339Nano, raw); err != nil {
			return fmt.Errorf("invalid %s header %q: %w", configUpdatedAtHeader, raw, err)
		}
	}

	if err = jsonrs.NewDecoder(res.Body).Decode(object); err != nil {
		return fmt.Errorf("failed to decode workspace configs: %w", err)
	}

	if r, ok := object.(diff.IncrementalDefinitionsReceiver); ok {
		r.SetIncrementalDefinitions(confirmed, cursor)
	}
	return nil
}

// supportsIncrementalDefinitions reports whether object is a [diff.IncrementalDefinitionsReceiver] and,
// if it is a [diff.UpdateableObject], all of its non updateable lists are [diff.OmittableList]s.
func supportsIncrementalDefinitions(object any) bool {
	if _, ok := object.(diff.IncrementalDefinitionsReceiver); !ok {
		return false
	}
	if uo, ok := object.(diff.UpdateableObject[string]); ok {
		for l := range uo.NonUpdateables() {
			if _, ok := l.(diff.OmittableList); !ok {
				return false
			}
		}
	}
	return true
}
