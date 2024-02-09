package poller

import (
	"context"
	"fmt"
	"time"
)

type Poller struct {
	caller   Caller
	handler  Handler
	interval time.Duration
}

type (
	Caller  func(ctx context.Context) (any, error)
	Handler func(any, error)
)

func New(caller Caller, handler Handler, opts ...Option) (*Poller, error) {
	if caller == nil {
		return nil, fmt.Errorf("caller is required")
	}

	if handler == nil {
		return nil, fmt.Errorf("handler is required")
	}

	p := &Poller{
		interval: 1 * time.Second,
		caller:   caller,
		handler:  handler,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p, nil
}

// Start starts the poller goroutine. It will poll for new workspace configs every interval.
// It will stop polling when the context is cancelled.
func (p *Poller) Start(ctx context.Context) {
	go func() {
		for {
			p.handler(p.caller(ctx))

			select {
			case <-ctx.Done():
				return
			case <-time.After(p.interval):
			}
		}
	}()
}

//result := gjson.GetBytes(response, "#.workspaces.@values.#.updatedAt|0")
//for _, v := range result.Array() {
//	if v.Time().After(p.updatedAt) {
//		p.updatedAt = v.Time()
//	}
//}
