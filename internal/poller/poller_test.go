package poller_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-cp-sdk/internal/poller"
	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
)

func TestPollerNew(t *testing.T) {
	t.Run("should return error if handler is nil", func(t *testing.T) {
		p, err := poller.New(nil)
		require.Nil(t, p)
		require.Error(t, err)
	})

	t.Run("should return error if client is nil", func(t *testing.T) {
		p, err := poller.New(func(*modelv2.WorkspaceConfigs) error { return nil })
		require.Nil(t, p)
		require.Error(t, err)
	})
}

func TestPoller(t *testing.T) {
	t.Run("should poll using client and workspace configs handler", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client := &mockClient{calls: []clientCall{
			{
				dataToBeReturned:  mockedResponses[0],
				expectedUpdatedAt: time.Time{},
			},
			{
				dataToBeReturned:  mockedResponses[1],
				expectedUpdatedAt: mockedResponses[0].UpdatedAt(),
			},
			{
				dataToBeReturned:  mockedResponses[2],
				expectedUpdatedAt: mockedResponses[1].UpdatedAt(),
			},
		}}

		var wg sync.WaitGroup
		wg.Add(len(mockedResponses))
		expectedResponseIndex := 0

		startTestPoller(t, ctx, client, func(wcs *modelv2.WorkspaceConfigs) error {
			require.Equal(t, mockedResponses[expectedResponseIndex], wcs)
			expectedResponseIndex++
			if expectedResponseIndex == len(mockedResponses) {
				cancel()
			}
			wg.Done()
			return nil
		})

		wg.Wait()
	})

	t.Run("should skip failed client requests", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client := &mockClient{calls: []clientCall{
			{
				errToBeReturned: errors.New("first call failed"),
			},
			{
				dataToBeReturned:  mockedResponses[0],
				expectedUpdatedAt: time.Time{},
			},
			{
				dataToBeReturned:  mockedResponses[1],
				expectedUpdatedAt: mockedResponses[0].UpdatedAt(),
			},
			{
				dataToBeReturned:  mockedResponses[2],
				expectedUpdatedAt: mockedResponses[1].UpdatedAt(),
			},
		}}

		var wg sync.WaitGroup
		wg.Add(len(mockedResponses))
		expectedResponseIndex := 0

		startTestPoller(t, ctx, client, func(wcs *modelv2.WorkspaceConfigs) error {
			require.Equal(t, mockedResponses[expectedResponseIndex], wcs)
			expectedResponseIndex++
			if expectedResponseIndex == len(mockedResponses) {
				cancel()
			}
			wg.Done()
			return nil
		})

		wg.Wait()
	})

	t.Run("should skip handler failures without updating updatedAt", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client := &mockClient{calls: []clientCall{
			{ // this will be called twice, once for the first failed handler call and once for the second
				dataToBeReturned:  mockedResponses[0],
				expectedUpdatedAt: time.Time{},
			},
			{
				dataToBeReturned:  mockedResponses[0],
				expectedUpdatedAt: time.Time{},
			},
			{
				dataToBeReturned:  mockedResponses[1],
				expectedUpdatedAt: mockedResponses[0].UpdatedAt(),
			},
			{
				dataToBeReturned:  mockedResponses[2],
				expectedUpdatedAt: mockedResponses[1].UpdatedAt(),
			},
		}}

		var wg sync.WaitGroup
		wg.Add(len(mockedResponses))
		expectedResponseIndex := 0
		var hasReturnedError bool
		// start a poller with handler that fails on first attempt and succeeds on second
		startTestPoller(t, ctx, client, func(wcs *modelv2.WorkspaceConfigs) error {
			if !hasReturnedError {
				hasReturnedError = true
				return errors.New("first call failed")
			}

			expectedResponseIndex++
			if expectedResponseIndex == len(mockedResponses) {
				cancel()
			}
			wg.Done()
			return nil
		})

		wg.Wait()
	})
}

func startTestPoller(t *testing.T, ctx context.Context, client poller.Client, handler poller.WorkspaceConfigHandler) {
	t.Helper()

	p, err := poller.New(handler,
		poller.WithClient(client),
		poller.WithPollingInterval(time.Nanosecond),
		poller.WithPollingBackoffInitialInterval(time.Nanosecond),
		poller.WithPollingBackoffMaxInterval(time.Nanosecond),
		poller.WithPollingBackoffMultiplier(1),
	)
	require.NoError(t, err)

	done := make(chan struct{})
	t.Cleanup(func() { <-done })
	go func() {
		p.Run(ctx)
		close(done)
	}()
}

type mockClient struct {
	calls    []clientCall
	nextCall int
}

type clientCall struct {
	dataToBeReturned  *modelv2.WorkspaceConfigs
	errToBeReturned   error
	expectedUpdatedAt time.Time
}

func (m *mockClient) GetWorkspaceConfigs(ctx context.Context, object any, updatedAfter time.Time) error {
	if m.nextCall >= len(m.calls) {
		return errors.New("no more calls")
	}

	call := m.calls[m.nextCall]
	m.nextCall++

	if call.expectedUpdatedAt.Nanosecond() != updatedAfter.Nanosecond() {
		return errors.New("unexpected updatedAt")
	}

	if call.errToBeReturned != nil {
		return call.errToBeReturned
	}

	*object.(*modelv2.WorkspaceConfigs) = *call.dataToBeReturned

	return ctx.Err()
}

var mockedResponses = []*modelv2.WorkspaceConfigs{
	{
		Workspaces: map[string]*modelv2.WorkspaceConfig{
			"wc-1": {UpdatedAt: time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)},
			"wc-2": {UpdatedAt: time.Date(2009, 11, 18, 20, 34, 58, 651387237, time.UTC)},
			"wc-3": {UpdatedAt: time.Date(2009, 11, 19, 20, 34, 58, 651387237, time.UTC)},
		},
	},
	{
		Workspaces: map[string]*modelv2.WorkspaceConfig{
			"wc-1": nil,
			"wc-2": {UpdatedAt: time.Date(2009, 11, 20, 20, 34, 58, 651387237, time.UTC)},
			"wc-3": nil,
		},
	},
	{
		Workspaces: map[string]*modelv2.WorkspaceConfig{
			"wc-1": {UpdatedAt: time.Date(2009, 11, 21, 20, 34, 58, 651387237, time.UTC)},
			"wc-4": {UpdatedAt: time.Date(2009, 11, 22, 20, 34, 58, 651387237, time.UTC)},
		},
	},
}
