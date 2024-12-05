package poller

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/rudderlabs/rudder-cp-sdk/diff"
	"github.com/rudderlabs/rudder-go-kit/logger"
	obskit "github.com/rudderlabs/rudder-observability-kit/go/labels"
)

type WorkspaceConfigsHandler[K comparable, T diff.UpdateableElement] func(list diff.UpdateableList[K, T]) (time.Time, error)

type WorkspaceConfigsGetter[K comparable, T diff.UpdateableElement] func(ctx context.Context, l diff.UpdateableList[K, T], updatedAfter time.Time) error

// WorkspaceConfigsPoller periodically polls for new workspace configs and runs a handler on them.
type WorkspaceConfigsPoller[K comparable, T diff.UpdateableElement] struct {
	getter      WorkspaceConfigsGetter[K, T]
	handler     WorkspaceConfigsHandler[K, T]
	constructor func() diff.UpdateableList[K, T]
	interval    time.Duration
	updatedAt   time.Time
	onResponse  func(error)
	backoff     struct {
		initialInterval time.Duration
		maxInterval     time.Duration
		multiplier      float64
	}
	log logger.Logger
}

func NewWorkspaceConfigsPoller[K comparable, T diff.UpdateableElement](
	getter WorkspaceConfigsGetter[K, T],
	handler WorkspaceConfigsHandler[K, T],
	constructor func() diff.UpdateableList[K, T],
	opts ...Option[K, T],
) (*WorkspaceConfigsPoller[K, T], error) {
	p := &WorkspaceConfigsPoller[K, T]{
		getter:      getter,
		handler:     handler,
		constructor: constructor,
		interval:    1 * time.Second,
		log:         logger.NOP,
	}
	p.backoff.initialInterval = 1 * time.Second
	p.backoff.maxInterval = 1 * time.Minute
	p.backoff.multiplier = 1.5

	for _, opt := range opts {
		opt(p)
	}

	if p.handler == nil {
		return nil, fmt.Errorf("handler is required")
	}

	if p.constructor == nil {
		return nil, fmt.Errorf("constructor is required")
	}

	if p.getter == nil {
		return nil, fmt.Errorf("getter is required")
	}

	return p, nil
}

// Run starts polling for new workspace configs every interval.
// It will stop polling when the context is cancelled.
func (p *WorkspaceConfigsPoller[K, T]) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(p.interval):
			err := p.poll(ctx)
			if p.onResponse != nil {
				p.onResponse(err)
			}
			if err == nil {
				continue
			}

			p.log.Errorn("failed to poll workspace configs", obskit.Error(err))

			nextBackOff := p.nextBackOff()
		retryLoop:
			for delay := nextBackOff(); delay != backoff.Stop; delay = nextBackOff() {
				select {
				case <-ctx.Done():
					return
				case <-time.After(delay):
					err = p.poll(ctx)
					if p.onResponse != nil {
						p.onResponse(err)
					}
					if err != nil {
						p.log.Warnn("failed to poll workspace configs after delay",
							logger.NewDurationField("delay", delay),
							obskit.Error(err),
						)
					} else {
						break retryLoop
					}
				}
			}
			if err != nil {
				p.log.Errorn("failed to poll workspace configs after backoff",
					logger.NewDurationField("backoff", p.backoff.maxInterval),
					obskit.Error(err),
				)
			}
		}
	}
}

func (p *WorkspaceConfigsPoller[K, T]) poll(ctx context.Context) error {
	p.log.Debugn("polling for workspace configs", logger.NewTimeField("updatedAt", p.updatedAt))

	response := p.constructor()
	err := p.getter(ctx, response, p.updatedAt)
	if err != nil {
		return fmt.Errorf("failed to get updated workspace configs: %w", err)
	}

	updatedAt, err := p.handler(response)
	if err != nil {
		return fmt.Errorf("failed to handle workspace configs: %w", err)
	}

	if !updatedAt.IsZero() {
		p.updatedAt = updatedAt
	}

	return nil
}

func (p *WorkspaceConfigsPoller[K, T]) nextBackOff() func() time.Duration {
	return backoff.NewExponentialBackOff(
		backoff.WithInitialInterval(p.backoff.initialInterval),
		backoff.WithMaxInterval(p.backoff.maxInterval),
		backoff.WithMultiplier(p.backoff.multiplier),
	).NextBackOff
}
