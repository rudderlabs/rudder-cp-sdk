package base

import (
	"net/url"
	"time"
)

type QueryOption func(q url.Values)

func WithUpdatedAfter(t time.Time) QueryOption {
	return func(q url.Values) {
		if !t.IsZero() {
			q.Add("updatedAfter", t.Format(updatedAfterTimeFormat))
		}
	}
}

// WithSecrets controls whether account secrets are embedded in the response or omitted.
// An empty value leaves the parameter out of the request, letting the control plane apply its default.
func WithSecrets(s string) QueryOption {
	return func(q url.Values) {
		if s != "" {
			q.Add("secrets", s)
		}
	}
}
