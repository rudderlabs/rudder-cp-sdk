package poller

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
)

//func TestPollerNew(t *testing.T) {
//	t.Run("should return error if handler is nil", func(t *testing.T) {
//		p, err := poller.New(nil)
//		require.Nil(t, p)
//		require.Error(t, err)
//	})
//
//	t.Run("should return error if client is nil", func(t *testing.T) {
//		p, err := poller.New(func(*modelv2.WorkspaceConfigs) error { return nil })
//		require.Nil(t, p)
//		require.Error(t, err)
//	})
//}
//
//func TestPoller(t *testing.T) {
//	ctrl := gomock.NewController(t)
//
//	t.Run("should poll using client and workspace configs handler", func(t *testing.T) {
//		ctx, cancel := context.WithCancel(context.Background())
//		defer cancel()
//
//		client := mocks.NewMockClient(ctrl)
//		client.EXPECT().GetWorkspaceConfigs(gomock.Any()).Return(mockedResponses[0], nil).Times(1)
//		client.EXPECT().GetUpdatedWorkspaceConfigs(gomock.Any(), mockedResponses[0].UpdatedAt()).Return(mockedResponses[1], nil).Times(1)
//		client.EXPECT().GetUpdatedWorkspaceConfigs(gomock.Any(), mockedResponses[1].UpdatedAt()).Return(mockedResponses[2], nil).Times(1)
//
//		var wg sync.WaitGroup
//		wg.Add(len(mockedResponses))
//		expectedResponseIndex := 0
//
//		startTestPoller(t, ctx, client, func(wcs *modelv2.WorkspaceConfigs) error {
//			require.Equal(t, mockedResponses[expectedResponseIndex], wcs)
//			expectedResponseIndex++
//			if expectedResponseIndex == len(mockedResponses) {
//				cancel()
//			}
//			wg.Done()
//			return nil
//		})
//
//		wg.Wait()
//	})
//
//	t.Run("should skip failed client requests", func(t *testing.T) {
//		ctx, cancel := context.WithCancel(context.Background())
//		defer cancel()
//
//		client := mocks.NewMockClient(ctrl)
//		client.EXPECT().GetWorkspaceConfigs(gomock.Any()).Return(nil, errors.New("first call failed")).Times(1)
//		client.EXPECT().GetWorkspaceConfigs(gomock.Any()).Return(mockedResponses[0], nil).Times(1)
//		client.EXPECT().GetUpdatedWorkspaceConfigs(gomock.Any(), mockedResponses[0].UpdatedAt()).Return(mockedResponses[1], nil).Times(1)
//		client.EXPECT().GetUpdatedWorkspaceConfigs(gomock.Any(), mockedResponses[1].UpdatedAt()).Return(nil, errors.New("fourth call failed")).Times(1)
//		client.EXPECT().GetUpdatedWorkspaceConfigs(gomock.Any(), mockedResponses[1].UpdatedAt()).Return(mockedResponses[2], nil).Times(1)
//
//		var wg sync.WaitGroup
//		wg.Add(len(mockedResponses))
//		expectedResponseIndex := 0
//
//		startTestPoller(t, ctx, client, func(wcs *modelv2.WorkspaceConfigs) error {
//			require.Equal(t, mockedResponses[expectedResponseIndex], wcs)
//			expectedResponseIndex++
//			if expectedResponseIndex == len(mockedResponses) {
//				cancel()
//			}
//			wg.Done()
//			return nil
//		})
//
//		wg.Wait()
//	})
//
//	t.Run("should skip handler failures without updating updatedAt", func(t *testing.T) {
//		ctx, cancel := context.WithCancel(context.Background())
//		defer cancel()
//
//		client := mocks.NewMockClient(ctrl)
//		client.EXPECT().GetWorkspaceConfigs(gomock.Any()).Return(mockedResponses[0], nil).Times(1)
//		// this will be called twice, once for the first failed handler call and once for the second
//		client.EXPECT().GetWorkspaceConfigs(gomock.Any()).Return(mockedResponses[0], nil).Times(1)
//		client.EXPECT().GetUpdatedWorkspaceConfigs(gomock.Any(), mockedResponses[0].UpdatedAt()).Return(mockedResponses[1], nil).Times(1)
//		client.EXPECT().GetUpdatedWorkspaceConfigs(gomock.Any(), mockedResponses[1].UpdatedAt()).Return(mockedResponses[2], nil).Times(1)
//
//		var wg sync.WaitGroup
//		wg.Add(len(mockedResponses))
//		expectedResponseIndex := 0
//		var hasReturnedError bool
//		// start a poller with handler that fails on first attempt and succeeds on second
//		startTestPoller(t, ctx, client, func(wcs *modelv2.WorkspaceConfigs) error {
//			if !hasReturnedError {
//				hasReturnedError = true
//				return errors.New("first call failed")
//			}
//
//			expectedResponseIndex++
//			if expectedResponseIndex == len(mockedResponses) {
//				cancel()
//			}
//			wg.Done()
//			return nil
//		})
//
//		wg.Wait()
//	})
//}

func TestPollerUpdatedAtParsing(t *testing.T) {
	mockedResponse := []*modelv2.WorkspaceConfigs{
		{
			Workspaces: map[string]*modelv2.WorkspaceConfig{
				"wc-1": {UpdatedAt: time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)},
				"wc-2": {UpdatedAt: time.Date(2009, 11, 21, 20, 34, 58, 651387237, time.UTC)},
				"wc-3": {UpdatedAt: time.Date(2009, 11, 19, 20, 34, 58, 651387237, time.UTC)},
			},
		},
	}
	data, err := json.Marshal(mockedResponse)
	require.NoError(t, err)

	done := make(chan struct{})
	client := fakeClient{data: data}
	handler := func(v []byte) error {
		require.Equal(t, data, v)
		close(done)
		return nil
	}

	p, err := New(handler,
		WithClient(&client),
		WithPollingInterval(1*time.Millisecond),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	p.Start(ctx)
	<-done
	cancel()

	require.Equal(t, time.Date(2009, 11, 21, 20, 34, 58, 651387237, time.UTC), p.updatedAt)
}

type fakeClient struct {
	data []byte
}

func (f *fakeClient) GetWorkspaceConfigs(_ context.Context) ([]byte, error) {
	return f.data, nil
}

func (f *fakeClient) GetUpdatedWorkspaceConfigs(_ context.Context, _ time.Time) ([]byte, error) {
	return f.data, nil
}

//func startTestPoller(t *testing.T, ctx context.Context, client poller.Client, handler poller.WorkspaceConfigHandler) {
//	p, err := poller.New(handler,
//		poller.WithClient(client),
//		poller.WithPollingInterval(1*time.Millisecond),
//	)
//	require.NoError(t, err)
//	p.Start(ctx)
//}
//
//var mockedResponses = []*modelv2.WorkspaceConfigs{
//	{
//		Workspaces: map[string]*modelv2.WorkspaceConfig{
//			"wc-1": {UpdatedAt: time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)},
//			"wc-2": {UpdatedAt: time.Date(2009, 11, 18, 20, 34, 58, 651387237, time.UTC)},
//			"wc-3": {UpdatedAt: time.Date(2009, 11, 19, 20, 34, 58, 651387237, time.UTC)},
//		},
//	},
//	{
//		Workspaces: map[string]*modelv2.WorkspaceConfig{
//			"wc-1": nil,
//			"wc-2": {UpdatedAt: time.Date(2009, 11, 20, 20, 34, 58, 651387237, time.UTC)},
//			"wc-3": nil,
//		},
//	},
//	{
//		Workspaces: map[string]*modelv2.WorkspaceConfig{
//			"wc-1": {UpdatedAt: time.Date(2009, 11, 21, 20, 34, 58, 651387237, time.UTC)},
//			"wc-4": {UpdatedAt: time.Date(2009, 11, 22, 20, 34, 58, 651387237, time.UTC)},
//		},
//	},
//}
