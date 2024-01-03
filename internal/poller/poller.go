//go:generate mockgen -source=poller.go -destination=mocks/poller.go -package=mocks
package poller

import (
	"context"
	"fmt"
	"time"

	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

// Poller periodically polls for new workspace configs and runs a handler on them.
type Poller struct {
	client    Client
	interval  time.Duration
	handler   WorkspaceConfigHandler
	updatedAt time.Time
	log       logger.Logger
}

type WorkspaceConfigHandler func(*modelv2.WorkspaceConfigs) error

type Client interface {
	GetWorkspaceConfigs(ctx context.Context) (*modelv2.WorkspaceConfigs, error)
	GetUpdatedWorkspaceConfigs(ctx context.Context, updatedAt time.Time) (*modelv2.WorkspaceConfigs, error)
}

func New(handler WorkspaceConfigHandler, opts ...Option) (*Poller, error) {
	p := &Poller{
		interval: 1 * time.Second,
		handler:  handler,
		log:      logger.NOP,
	}

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

// Start starts the poller goroutine. It will poll for new workspace configs every interval.
// It will stop polling when the context is cancelled.
func (p *Poller) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(p.interval):
				if err := p.poll(ctx); err != nil {
					p.log.Errorf("failed to poll for workspace configs: %v", err)
				}
			}
		}
	}()
}

func (p *Poller) poll(ctx context.Context) error {
	var response *modelv2.WorkspaceConfigs
	if p.updatedAt.IsZero() {
		p.log.Debug("polling for workspace configs")
		wcs, err := p.client.GetWorkspaceConfigs(ctx)
		if err != nil {
			return fmt.Errorf("failed to get workspace configs: %w", err)
		}

		response = wcs
	} else {
		p.log.Debugf("polling for workspace configs updated after %v", p.updatedAt)
		wcs, err := p.client.GetUpdatedWorkspaceConfigs(ctx, p.updatedAt)
		if err != nil {
			return fmt.Errorf("failed to get updated workspace configs: %w", err)
		}

		response = wcs
	}

	if err := p.handler(response); err != nil {
		return fmt.Errorf("failed to handle workspace configs: %w", err)
	}

	// only update updatedAt if we managed to handle the response
	// so that we don't miss any updates in case of an error
	p.updatedAt = response.UpdatedAt()

	return nil
}
