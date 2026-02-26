package poller

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/logger"

	"github.com/rudderlabs/rudder-cp-sdk/diff"
	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
)

func TestPollerNew(t *testing.T) {
	getter := func(_ context.Context, _ diff.UpdateableObject[string], _ time.Time) error {
		return nil
	}
	handler := func(_ diff.UpdateableObject[string]) (time.Time, bool, error) {
		return time.Time{}, false, nil
	}
	constructor := func() diff.UpdateableObject[string] {
		return nil
	}

	t.Run("should return error if getter is nil", func(t *testing.T) {
		p, err := NewWorkspaceConfigsPoller[string](nil, handler, constructor)
		require.Nil(t, p)
		require.Error(t, err)
	})

	t.Run("should return error if handler is nil", func(t *testing.T) {
		p, err := NewWorkspaceConfigsPoller[string](getter, nil, constructor)
		require.Nil(t, p)
		require.Error(t, err)
	})

	t.Run("should return error if constructor is nil", func(t *testing.T) {
		p, err := NewWorkspaceConfigsPoller[string](getter, handler, nil)
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
		runTestPoller(t, ctx, client, func(obj diff.UpdateableObject[string]) (time.Time, bool, error) {
			defer wg.Done()

			require.Equalf(t, mockedResponses[expectedResponseIndex], obj, "Response index: %d", expectedResponseIndex)

			expectedResponseIndex++
			if expectedResponseIndex == len(mockedResponses) {
				cancel()
			}

			return getLatestUpdatedAt(obj), true, nil
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
		runTestPoller(t, ctx, client, func(obj diff.UpdateableObject[string]) (time.Time, bool, error) {
			defer wg.Done()

			require.Equalf(t, mockedResponses[expectedResponseIndex], obj, "Response index: %d", expectedResponseIndex)

			expectedResponseIndex++
			if expectedResponseIndex == len(mockedResponses) {
				cancel()
			}

			return getLatestUpdatedAt(obj), true, nil
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
		runTestPoller(t, ctx, client, func(obj diff.UpdateableObject[string]) (time.Time, bool, error) {
			if !hasReturnedError {
				hasReturnedError = true
				return time.Time{}, false, errors.New("first call failed")
			}

			expectedResponseIndex++
			if expectedResponseIndex == len(mockedResponses) {
				cancel()
			}

			wg.Done()

			return getLatestUpdatedAt(obj), true, nil
		})

		wg.Wait()
	})
}

func runTestPoller(
	t *testing.T,
	ctx context.Context,
	client *mockClient,
	handler func(object diff.UpdateableObject[string]) (time.Time, bool, error),
) {
	t.Helper()

	poll, err := setupPoller(
		func(ctx context.Context, object any, updatedAfter time.Time) error {
			return client.GetWorkspaceConfigs(ctx, object, updatedAfter)
		},
		func(obj diff.UpdateableObject[string]) (time.Time, bool, error) {
			return handler(obj)
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

func setupPoller(
	getter func(ctx context.Context, object any, updatedAfter time.Time) error,
	handler WorkspaceConfigsHandler[string],
	log logger.Logger,
) (func(context.Context), error) {
	p, err := newWorkspaceConfigsPoller[string](
		func(ctx context.Context, l diff.UpdateableObject[string], updatedAfter time.Time) error {
			return getter(ctx, l, updatedAfter)
		},
		handler,
		func() diff.UpdateableObject[string] {
			return &modelv2.WorkspaceConfigs{}
		},
		log,
	)
	if err != nil {
		return nil, fmt.Errorf("error setting up poller: %v", err)
	}

	return p.Run, nil
}

func newWorkspaceConfigsPoller[K comparable](
	getter WorkspaceConfigsGetter[K],
	handler WorkspaceConfigsHandler[K],
	constructor func() diff.UpdateableObject[K],
	log logger.Logger,
) (*WorkspaceConfigsPoller[K], error) {
	return NewWorkspaceConfigsPoller(getter, handler, constructor,
		WithLogger[K](log.Child("poller")),
		WithPollingInterval[K](time.Nanosecond),
		WithPollingBackoffInitialInterval[K](time.Nanosecond),
		WithPollingBackoffMaxInterval[K](time.Nanosecond),
		WithPollingBackoffMultiplier[K](1),
	)
}

func getLatestUpdatedAt() func(list diff.UpdateableObject[string]) time.Time {
	var latestUpdatedAt time.Time
	return func(obj diff.UpdateableObject[string]) time.Time {
		for uo := range obj.Updateables() {
			for _, wc := range uo.List() {
				if wc.IsNil() || wc.GetUpdatedAt().IsZero() {
					continue
				}
				if wc.GetUpdatedAt().After(latestUpdatedAt) {
					latestUpdatedAt = wc.GetUpdatedAt()
				}
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

	type T = *modelv2.WorkspaceConfigs
	*object.(T) = *call.dataToBeReturned.(T)

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
