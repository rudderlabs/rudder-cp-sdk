package poller

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-cp-sdk/diff"
	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

func TestPollerNew(t *testing.T) {
	getter := func(ctx context.Context, l diff.UpdateableList[string, diff.UpdateableElement], updatedAfter time.Time) error {
		return nil
	}
	handler := func(list diff.UpdateableList[string, diff.UpdateableElement]) (time.Time, error) {
		return time.Time{}, nil
	}
	constructor := func() diff.UpdateableList[string, diff.UpdateableElement] {
		return nil
	}

	t.Run("should return error if getter is nil", func(t *testing.T) {
		p, err := NewWorkspaceConfigsPoller[string, diff.UpdateableElement](nil, handler, constructor)
		require.Nil(t, p)
		require.Error(t, err)
	})

	t.Run("should return error if handler is nil", func(t *testing.T) {
		p, err := NewWorkspaceConfigsPoller[string, diff.UpdateableElement](getter, nil, constructor)
		require.Nil(t, p)
		require.Error(t, err)
	})

	t.Run("should return error if constructor is nil", func(t *testing.T) {
		p, err := NewWorkspaceConfigsPoller[string, diff.UpdateableElement](getter, handler, nil)
		require.Nil(t, p)
		require.Error(t, err)
	})
}

func TestPoller(t *testing.T) {
	t.Run("should poll using getter and workspace configs handler", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		client := &mockClient{calls: []clientCall{
			{
				dataToBeReturned:  mockedResponses[0],
				expectedUpdatedAt: time.Time{},
			},
			{
				dataToBeReturned:  mockedResponses[1],
				expectedUpdatedAt: time.Date(2009, 11, 19, 20, 34, 58, 651387237, time.UTC),
			},
			{
				dataToBeReturned:  mockedResponses[2],
				expectedUpdatedAt: time.Date(2009, 11, 20, 20, 34, 58, 651387237, time.UTC),
			},
		}}

		var wg sync.WaitGroup
		wg.Add(len(mockedResponses))
		expectedResponseIndex := 0

		getLatestUpdatedAt := getLatestUpdatedAt()
		runTestPoller(t, ctx, client, func(list diff.UpdateableList[string, *modelv2.WorkspaceConfig]) (time.Time, error) {
			defer wg.Done()

			require.Equalf(t, mockedResponses[expectedResponseIndex], list, "Response index: %d", expectedResponseIndex)

			expectedResponseIndex++
			if expectedResponseIndex == len(mockedResponses) {
				cancel()
			}

			return getLatestUpdatedAt(list), nil
		})

		wg.Wait()
	})

	t.Run("should skip failed getter requests", func(t *testing.T) {
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
				expectedUpdatedAt: time.Date(2009, 11, 19, 20, 34, 58, 651387237, time.UTC),
			},
			{
				dataToBeReturned:  mockedResponses[2],
				expectedUpdatedAt: time.Date(2009, 11, 20, 20, 34, 58, 651387237, time.UTC),
			},
		}}

		var wg sync.WaitGroup
		wg.Add(len(mockedResponses))
		expectedResponseIndex := 0

		getLatestUpdatedAt := getLatestUpdatedAt()
		runTestPoller(t, ctx, client, func(list diff.UpdateableList[string, *modelv2.WorkspaceConfig]) (time.Time, error) {
			defer wg.Done()

			require.Equalf(t, mockedResponses[expectedResponseIndex], list, "Response index: %d", expectedResponseIndex)

			expectedResponseIndex++
			if expectedResponseIndex == len(mockedResponses) {
				cancel()
			}

			return getLatestUpdatedAt(list), nil
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
				expectedUpdatedAt: time.Date(2009, 11, 19, 20, 34, 58, 651387237, time.UTC),
			},
			{
				dataToBeReturned:  mockedResponses[2],
				expectedUpdatedAt: time.Date(2009, 11, 20, 20, 34, 58, 651387237, time.UTC),
			},
		}}

		var wg sync.WaitGroup
		wg.Add(len(mockedResponses))
		expectedResponseIndex := 0
		var hasReturnedError bool
		// start a poller with handler that fails on first attempt and succeeds on second
		getLatestUpdatedAt := getLatestUpdatedAt()
		runTestPoller(t, ctx, client, func(list diff.UpdateableList[string, *modelv2.WorkspaceConfig]) (time.Time, error) {
			if !hasReturnedError {
				hasReturnedError = true
				return time.Time{}, errors.New("first call failed")
			}

			expectedResponseIndex++
			if expectedResponseIndex == len(mockedResponses) {
				cancel()
			}

			wg.Done()

			return getLatestUpdatedAt(list), nil
		})

		wg.Wait()
	})
}

func runTestPoller[K string](
	t *testing.T,
	ctx context.Context,
	client *mockClient,
	handler func(diff.UpdateableList[K, *modelv2.WorkspaceConfig]) (time.Time, error),
) {
	t.Helper()

	poll, err := setupPoller(
		func(ctx context.Context, object any, updatedAfter time.Time) error {
			return client.GetWorkspaceConfigs(ctx, object, updatedAfter)
		},
		func(list diff.UpdateableList[K, *modelv2.WorkspaceConfig]) (time.Time, error) {
			return handler(list)
		},
		logger.NOP,
	)
	require.NoError(t, err)

	done := make(chan struct{})
	t.Cleanup(func() { <-done })
	go func() {
		poll(ctx)
		close(done)
	}()
}

func setupPoller[K string](
	getter func(ctx context.Context, object any, updatedAfter time.Time) error,
	handler WorkspaceConfigsHandler[K, *modelv2.WorkspaceConfig],
	log logger.Logger,
) (func(context.Context), error) {
	p, err := newWorkspaceConfigsPoller[K, *modelv2.WorkspaceConfig](
		func(ctx context.Context, l diff.UpdateableList[K, *modelv2.WorkspaceConfig], updatedAfter time.Time) error {
			return getter(ctx, l, updatedAfter)
		},
		handler,
		func() diff.UpdateableList[K, *modelv2.WorkspaceConfig] {
			return &modelv2.WorkspaceConfigs[K, *modelv2.WorkspaceConfig]{}
		},
		log,
	)
	if err != nil {
		return nil, fmt.Errorf("error setting up poller: %v", err)
	}

	return p.Run, nil
}

func newWorkspaceConfigsPoller[K comparable, T diff.UpdateableElement](
	getter WorkspaceConfigsGetter[K, T],
	handler WorkspaceConfigsHandler[K, T],
	constructor func() diff.UpdateableList[K, T],
	log logger.Logger,
) (*WorkspaceConfigsPoller[K, T], error) {
	return NewWorkspaceConfigsPoller(getter, handler, constructor,
		WithLogger[K, T](log.Child("poller")),
		WithPollingInterval[K, T](time.Nanosecond),
		WithPollingBackoffInitialInterval[K, T](time.Nanosecond),
		WithPollingBackoffMaxInterval[K, T](time.Nanosecond),
		WithPollingBackoffMultiplier[K, T](1),
	)
}

func getLatestUpdatedAt() func(list diff.UpdateableList[string, *modelv2.WorkspaceConfig]) time.Time {
	var latestUpdatedAt time.Time
	return func(list diff.UpdateableList[string, *modelv2.WorkspaceConfig]) time.Time {
		for _, wc := range list.List() {
			if wc.IsNil() || wc.GetUpdatedAt().IsZero() {
				continue
			}
			if wc.GetUpdatedAt().After(latestUpdatedAt) {
				latestUpdatedAt = wc.GetUpdatedAt()
			}
		}
		return latestUpdatedAt
	}
}

type mockClient struct {
	calls    []clientCall
	nextCall int
}

type clientCall struct {
	dataToBeReturned  any
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

	type T = *modelv2.WorkspaceConfigs[string, *modelv2.WorkspaceConfig]
	*object.(T) = *call.dataToBeReturned.(T)

	return ctx.Err()
}

var mockedResponses = []*modelv2.WorkspaceConfigs[string, *modelv2.WorkspaceConfig]{
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
