package poller

import (
	"time"

	"github.com/rudderlabs/rudder-go-kit/logger"
)

type Option func(*Poller)

func WithClient(client Client) Option {
	return func(p *Poller) { p.client = client }
}

func WithLogger(log logger.Logger) Option {
	return func(p *Poller) { p.log = log }
}

func WithPollingInterval(d time.Duration) Option {
	return func(p *Poller) { p.interval = d }
}

func WithPollingBackoffInitialInterval(d time.Duration) Option {
	return func(p *Poller) { p.backoff.initialInterval = d }
}

func WithPollingBackoffMaxInterval(d time.Duration) Option {
	return func(p *Poller) { p.backoff.maxInterval = d }
}

func WithPollingBackoffMultiplier(m float64) Option {
	return func(p *Poller) { p.backoff.multiplier = m }
}

func WithOnResponse(f func(error)) Option {
	return func(p *Poller) { p.onResponse = f }
}
