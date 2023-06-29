package poller

import (
	"time"

	"github.com/rudderlabs/rudder-go-kit/logger"
)

type Option func(*Poller)

func WithClient(client Client) Option {
	return func(p *Poller) {
		p.client = client
	}
}

func WithPollingInterval(interval time.Duration) Option {
	return func(p *Poller) {
		p.interval = interval
	}
}

func WithLogger(log logger.Logger) Option {
	return func(p *Poller) {
		p.log = log
	}
}
