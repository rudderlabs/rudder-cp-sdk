//go:generate mockgen -source=poller.go -destination=mocks/poller.go -package=mocks
package poller

import (
	"context"
	"fmt"
	"time"

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

type WorkspaceConfigHandler func([]byte) error

type Client interface {
	GetWorkspaceConfigs(ctx context.Context) ([]byte, error)
	GetUpdatedWorkspaceConfigs(ctx context.Context, updatedAt time.Time) ([]byte, error)
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
	return nil
}
