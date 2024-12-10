package poller

import (
	"time"

	"github.com/rudderlabs/rudder-go-kit/logger"
)

type Option[K comparable] func(*WorkspaceConfigsPoller[K])

func WithLogger[K comparable](log logger.Logger) Option[K] {
	return func(p *WorkspaceConfigsPoller[K]) { p.log = log }
}

func WithPollingInterval[K comparable](d time.Duration) Option[K] {
	return func(p *WorkspaceConfigsPoller[K]) { p.interval = d }
}

func WithPollingBackoffInitialInterval[K comparable](d time.Duration) Option[K] {
	return func(p *WorkspaceConfigsPoller[K]) { p.backoff.initialInterval = d }
}

func WithPollingBackoffMaxInterval[K comparable](d time.Duration) Option[K] {
	return func(p *WorkspaceConfigsPoller[K]) { p.backoff.maxInterval = d }
}

func WithPollingMaxElapsedTime[K comparable](d time.Duration) Option[K] {
	return func(p *WorkspaceConfigsPoller[K]) { p.backoff.maxElapsedTime = d }
}

func WithPollingMaxRetries[K comparable](n uint64) Option[K] {
	return func(p *WorkspaceConfigsPoller[K]) { p.backoff.maxRetries = n }
}

func WithPollingBackoffMultiplier[K comparable](m float64) Option[K] {
	return func(p *WorkspaceConfigsPoller[K]) { p.backoff.multiplier = m }
}

func WithOnResponse[K comparable](f func(bool, error)) Option[K] {
	return func(p *WorkspaceConfigsPoller[K]) { p.onResponse = f }
}
