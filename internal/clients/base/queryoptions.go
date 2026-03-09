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
