package poller

import (
	"time"
)

type Option func(*Poller)

func WithPollingInterval(interval time.Duration) Option {
	return func(p *Poller) {
		p.interval = interval
	}
}
