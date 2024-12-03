package poller

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
	"github.com/rudderlabs/rudder-go-kit/logger"
	obskit "github.com/rudderlabs/rudder-observability-kit/go/labels"
)

// Poller periodically polls for new workspace configs and runs a handler on them.
type Poller struct {
	client    Client
	interval  time.Duration
	handler   WorkspaceConfigHandler
	updatedAt time.Time
	backoff   struct {
		initialInterval time.Duration
		maxInterval     time.Duration
		multiplier      float64
	}
	log logger.Logger
}

type WorkspaceConfigHandler func(*modelv2.WorkspaceConfigs) error

type Client interface {
	GetWorkspaceConfigs(ctx context.Context, object any, updatedAfter time.Time) error
}

func New(handler WorkspaceConfigHandler, opts ...Option) (*Poller, error) {
	p := &Poller{
		interval: 1 * time.Second,
		handler:  handler,
		log:      logger.NOP,
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

	if p.client == nil {
		return nil, fmt.Errorf("client is required")
	}

	return p, nil
}

// Run starts polling for new workspace configs every interval.
// It will stop polling when the context is cancelled.
func (p *Poller) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(p.interval):
			err := p.poll(ctx)
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

/*
TODO we should detect inconsistencies, like if a workspace that we are not aware of is returned with null
as if it wasn't updated since the last call but we never received it. in that case we should log an error
and trigger a full update.
*/
func (p *Poller) poll(ctx context.Context) error {
	p.log.Debugn("polling for workspace configs", logger.NewTimeField("updatedAt", p.updatedAt))

	var wcs modelv2.WorkspaceConfigs
	err := p.client.GetWorkspaceConfigs(ctx, &wcs, p.updatedAt)
	if err != nil {
		return fmt.Errorf("failed to get updated workspace configs: %w", err)
	}

	if err := p.handler(&wcs); err != nil {
		return fmt.Errorf("failed to handle workspace configs: %w", err)
	}

	// only update updatedAt if we managed to handle the response
	// so that we don't miss any updates in case of an error
	if !wcs.UpdatedAt().IsZero() {
		// There is a case where all workspaces have not been updated since the last request.
		// In that case updatedAt will be zero.
		p.updatedAt = wcs.UpdatedAt()
	}

	return nil
}

func (p *Poller) nextBackOff() func() time.Duration {
	return backoff.NewExponentialBackOff(
		backoff.WithInitialInterval(p.backoff.initialInterval),
		backoff.WithMaxInterval(p.backoff.maxInterval),
		backoff.WithMultiplier(p.backoff.multiplier),
	).NextBackOff
}
