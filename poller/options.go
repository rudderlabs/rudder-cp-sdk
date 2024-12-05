package poller

import (
	"time"

	"github.com/rudderlabs/rudder-cp-sdk/diff"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

type Option[K comparable, T diff.UpdateableElement] func(*WorkspaceConfigsPoller[K, T])

func WithLogger[K comparable, T diff.UpdateableElement](log logger.Logger) Option[K, T] {
	return func(p *WorkspaceConfigsPoller[K, T]) { p.log = log }
}

func WithPollingInterval[K comparable, T diff.UpdateableElement](d time.Duration) Option[K, T] {
	return func(p *WorkspaceConfigsPoller[K, T]) { p.interval = d }
}

func WithPollingBackoffInitialInterval[K comparable, T diff.UpdateableElement](d time.Duration) Option[K, T] {
	return func(p *WorkspaceConfigsPoller[K, T]) { p.backoff.initialInterval = d }
}

func WithPollingBackoffMaxInterval[K comparable, T diff.UpdateableElement](d time.Duration) Option[K, T] {
	return func(p *WorkspaceConfigsPoller[K, T]) { p.backoff.maxInterval = d }
}

func WithPollingBackoffMultiplier[K comparable, T diff.UpdateableElement](m float64) Option[K, T] {
	return func(p *WorkspaceConfigsPoller[K, T]) { p.backoff.multiplier = m }
}

func WithOnResponse[K comparable, T diff.UpdateableElement](f func(error)) Option[K, T] {
	return func(p *WorkspaceConfigsPoller[K, T]) { p.onResponse = f }
}
