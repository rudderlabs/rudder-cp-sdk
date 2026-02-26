package poller

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v5"

	"github.com/rudderlabs/rudder-go-kit/logger"
	obskit "github.com/rudderlabs/rudder-observability-kit/go/labels"

	"github.com/rudderlabs/rudder-cp-sdk/diff"
)

type WorkspaceConfigsGetter[K comparable] func(ctx context.Context, l diff.UpdateableObject[K], updatedAfter time.Time) error

type WorkspaceConfigsHandler[K comparable] func(obj diff.UpdateableObject[K]) (time.Time, bool, error)

// WorkspaceConfigsPoller periodically polls for new workspace configs and runs a handler on them.
type WorkspaceConfigsPoller[K comparable] struct {
	getter      WorkspaceConfigsGetter[K]
	handler     WorkspaceConfigsHandler[K]
	constructor func() diff.UpdateableObject[K]
	interval    time.Duration
	updatedAt   time.Time
	onResponse  func(context.Context, bool, error)
	backoff     struct {
		initialInterval time.Duration
		maxInterval     time.Duration
		maxElapsedTime  time.Duration
		maxRetries      uint64
		multiplier      float64
	}
	log logger.Logger
}

func NewWorkspaceConfigsPoller[K comparable](
	getter WorkspaceConfigsGetter[K],
	handler WorkspaceConfigsHandler[K],
	constructor func() diff.UpdateableObject[K],
	opts ...Option[K],
) (*WorkspaceConfigsPoller[K], error) {
	p := &WorkspaceConfigsPoller[K]{
		getter:      getter,
		handler:     handler,
		constructor: constructor,
		interval:    1 * time.Second,
		log:         logger.NOP,
	}
	p.backoff.initialInterval = 1 * time.Second
	p.backoff.maxInterval = 1 * time.Minute
	p.backoff.maxElapsedTime = 5 * time.Minute
	p.backoff.maxRetries = 15
	p.backoff.multiplier = 1.5

	for _, opt := range opts {
		opt(p)
	}

	if p.getter == nil {
		return nil, fmt.Errorf("getter is required")
	}

	if p.handler == nil {
		return nil, fmt.Errorf("handler is required")
	}

	if p.constructor == nil {
		return nil, fmt.Errorf("constructor is required")
	}

	return p, nil
}

// Run starts polling for new workspace configs every interval.
// It will stop polling when the context is cancelled.
func (p *WorkspaceConfigsPoller[K]) Run(ctx context.Context) {
	for {
		_, err := backoff.Retry(ctx,
			func() (*struct{}, error) {
				updated, err := p.poll(ctx)
				if p.onResponse != nil {
					p.onResponse(ctx, updated, err)
				}
				return nil, err
			},
			backoff.WithBackOff(&backoff.ExponentialBackOff{
				InitialInterval:     p.backoff.initialInterval,
				RandomizationFactor: backoff.DefaultRandomizationFactor,
				Multiplier:          p.backoff.multiplier,
				MaxInterval:         p.backoff.maxInterval,
			}),
			backoff.WithMaxTries(uint(p.backoff.maxRetries)+1),
			backoff.WithMaxElapsedTime(p.backoff.maxElapsedTime),
			backoff.WithNotify(func(err error, d time.Duration) {
				p.log.Warnn("retrying workspace config poll after backoff delay",
					logger.NewDurationField("delay", d),
					obskit.Error(err),
				)
			}),
		)
		if err != nil {
			p.log.Errorn("failed to poll workspace configs after backoff",
				logger.NewDurationField("backoff", p.backoff.maxInterval),
				obskit.Error(err),
			)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(p.interval):

		}
	}
}

func (p *WorkspaceConfigsPoller[K]) poll(ctx context.Context) (bool, error) {
	p.log.Debugn("polling for workspace configs", logger.NewTimeField("updatedAt", p.updatedAt))

	response := p.constructor()
	err := p.getter(ctx, response, p.updatedAt)
	if err != nil {
		return false, fmt.Errorf("failed to get updated workspace configs: %w", err)
	}

	updatedAt, updated, err := p.handler(response)
	if err != nil {
		return false, fmt.Errorf("failed to handle workspace configs: %w", err)
	}

	if !updatedAt.IsZero() {
		p.updatedAt = updatedAt
	}

	return updated, nil
}
