package cpsdk

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/rudderlabs/rudder-cp-sdk/diff"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/rudderlabs/rudder-go-kit/testhelper/httptest"
)

const updatedAfterTimeFormat = "2006-01-02T15:04:05.000Z"

func TestIncrementalUpdates(t *testing.T) {
	var (
		ctx                  = context.Background()
		namespace            = "free-us-1"
		secret               = "service-secret"
		requestNumber        int
		receivedUpdatedAfter []time.Time
	)

	responseBodyFromFile, err := os.ReadFile("./testdata/sample_namespace.json")
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() { requestNumber++ }()

		user, _, ok := r.BasicAuth()
		require.True(t, ok)
		require.Equal(t, secret, user)

		var (
			err              error
			updatedAfterTime time.Time
			responseBody     []byte
		)
		for k, v := range r.URL.Query() {
			if k != "updatedAfter" {
				continue
			}

			updatedAfterTime, err = time.Parse(updatedAfterTimeFormat, v[0])
			require.NoError(t, err)

			receivedUpdatedAfter = append(receivedUpdatedAfter, updatedAfterTime)
		}

		switch requestNumber {
		case 0: // 1st request, return file content as is
			responseBody = responseBodyFromFile
		case 1: // 2nd request, return new workspace, no updates for the other 2
			responseBody = fmt.Appendf(nil, `{
				"workspaces": {
					"dummy":{"updatedAt":%q,"libraries":[{"versionId":"foo"},{"versionId":"bar"}]},
					"2hCBi02C8xYS8Rsy1m9bJjTlKy6":null,
					"2bVMV2JiAJe42OXZrzyvJI75v0N":null
				}
			}`, updatedAfterTime.Add(time.Minute).Format(updatedAfterTimeFormat))
		case 2: // 3rd request, return updated dummy workspace, no updates for the other 2
			responseBody = fmt.Appendf(nil, `{
				"workspaces": {
					"dummy":{"updatedAt":%q,"libraries":[{"versionId":"baz"}]},
					"2hCBi02C8xYS8Rsy1m9bJjTlKy6":null,
					"2bVMV2JiAJe42OXZrzyvJI75v0N":null
				}
			}`, updatedAfterTime.Add(time.Minute).Format(updatedAfterTimeFormat))
		case 3, 4: // 4th and 5th request, delete the dummy workspace
			responseBody = []byte(`{
				"workspaces": {
					"2hCBi02C8xYS8Rsy1m9bJjTlKy6":null,
					"2bVMV2JiAJe42OXZrzyvJI75v0N":null
				}
			}`)
		case 5: // new workspace, but its update time is before the last request, so no updates
			responseBody = []byte(`{
				"workspaces": {
					"someWorkspaceID": null
				}
			}`)
		default:
			responseBody = responseBodyFromFile
		}

		_, _ = w.Write(responseBody)
	}))
	defer ts.Close()

	cpSDK, err := New(
		WithBaseUrl(ts.URL),
		WithLogger(logger.NOP),
		WithNamespaceIdentity(namespace, secret),
	)
	require.NoError(t, err)

	getLatestUpdatedAt := getLatestUpdatedAt() // this is to cache the latestUpdatedAt

	// send the request the first time
	wcs := &modelv2.WorkspaceConfigs{}
	err = cpSDK.GetWorkspaceConfigs(ctx, wcs, time.Time{})
	require.NoError(t, err)
	require.Len(t, wcs.Workspaces, 2)
	require.Contains(t, wcs.Workspaces, "2hCBi02C8xYS8Rsy1m9bJjTlKy6")
	require.NotNil(t, wcs.Workspaces["2hCBi02C8xYS8Rsy1m9bJjTlKy6"])
	require.Contains(t, wcs.Workspaces, "2bVMV2JiAJe42OXZrzyvJI75v0N")
	require.NotNil(t, wcs.Workspaces["2bVMV2JiAJe42OXZrzyvJI75v0N"])
	require.Empty(t, receivedUpdatedAfter, "The first request should not have updatedAfter in the query params")
	latestUpdatedAt, updatedAt := getLatestUpdatedAt(wcs)
	require.Equal(t, "2024-11-27T20:13:30.647Z", latestUpdatedAt.Format(updatedAfterTimeFormat))
	require.Equal(t, "2024-11-27T20:13:30.647Z", updatedAt.Format(updatedAfterTimeFormat))

	// send the request again, should receive the new dummy workspace and no updates for the other 2 workspaces
	wcs = &modelv2.WorkspaceConfigs{} // reset the workspace configs
	err = cpSDK.GetWorkspaceConfigs(ctx, wcs, latestUpdatedAt)
	require.NoError(t, err)
	require.Len(t, wcs.Workspaces, 3)
	require.Contains(t, wcs.Workspaces, "2hCBi02C8xYS8Rsy1m9bJjTlKy6")
	require.Nil(t, wcs.Workspaces["2hCBi02C8xYS8Rsy1m9bJjTlKy6"], "The workspace should have not been updated")
	require.Contains(t, wcs.Workspaces, "2bVMV2JiAJe42OXZrzyvJI75v0N")
	require.Nil(t, wcs.Workspaces["2bVMV2JiAJe42OXZrzyvJI75v0N"], "The workspace should have not been updated")
	require.Contains(t, wcs.Workspaces, "dummy")
	require.NotNil(t, wcs.Workspaces["dummy"])
	require.Len(t, receivedUpdatedAfter, 1)
	latestUpdatedAt, updatedAt = getLatestUpdatedAt(wcs)
	require.Equal(t, "2024-11-27T20:14:30.647Z", latestUpdatedAt.Format(updatedAfterTimeFormat))
	require.Equal(t, "2024-11-27T20:14:30.647Z", updatedAt.Format(updatedAfterTimeFormat))
	expectedUpdatedAfter, err := time.Parse(updatedAfterTimeFormat, "2024-11-27T20:13:30.647Z")
	require.NoError(t, err)
	require.Equal(t, receivedUpdatedAfter[0], expectedUpdatedAfter, updatedAfterTimeFormat)

	// send the request again, should receive the updated dummy workspace
	wcs = &modelv2.WorkspaceConfigs{} // reset the workspace configs
	err = cpSDK.GetWorkspaceConfigs(ctx, wcs, latestUpdatedAt)
	require.NoError(t, err)
	require.Len(t, wcs.Workspaces, 3)
	require.Contains(t, wcs.Workspaces, "2hCBi02C8xYS8Rsy1m9bJjTlKy6")
	require.Nil(t, wcs.Workspaces["2hCBi02C8xYS8Rsy1m9bJjTlKy6"], "The workspace should have not been updated")
	require.Contains(t, wcs.Workspaces, "2bVMV2JiAJe42OXZrzyvJI75v0N")
	require.Nil(t, wcs.Workspaces["2bVMV2JiAJe42OXZrzyvJI75v0N"], "The workspace should have not been updated")
	require.Contains(t, wcs.Workspaces, "dummy")
	require.NotNil(t, wcs.Workspaces["dummy"])
	require.Len(t, receivedUpdatedAfter, 2)
	latestUpdatedAt, updatedAt = getLatestUpdatedAt(wcs)
	require.Equal(t, "2024-11-27T20:15:30.647Z", latestUpdatedAt.Format(updatedAfterTimeFormat))
	require.Equal(t, "2024-11-27T20:15:30.647Z", updatedAt.Format(updatedAfterTimeFormat))
	expectedUpdatedAfter, err = time.Parse(updatedAfterTimeFormat, "2024-11-27T20:14:30.647Z")
	require.NoError(t, err)
	require.Equal(t, receivedUpdatedAfter[1], expectedUpdatedAfter, updatedAfterTimeFormat)

	// send the request again, should not receive dummy since it was deleted
	wcs = &modelv2.WorkspaceConfigs{} // reset the workspace configs
	err = cpSDK.GetWorkspaceConfigs(ctx, wcs, latestUpdatedAt)
	require.NoError(t, err)
	latestUpdatedAt, updatedAt = getLatestUpdatedAt(wcs)
	require.Truef(t, updatedAt.IsZero(), "%+v", wcs)
	require.Equal(t, "2024-11-27T20:15:30.647Z", latestUpdatedAt.Format(updatedAfterTimeFormat))
	require.Len(t, wcs.Workspaces, 2)
	require.Contains(t, wcs.Workspaces, "2hCBi02C8xYS8Rsy1m9bJjTlKy6")
	require.Nil(t, wcs.Workspaces["2hCBi02C8xYS8Rsy1m9bJjTlKy6"], "The workspace should have not been updated")
	require.Contains(t, wcs.Workspaces, "2bVMV2JiAJe42OXZrzyvJI75v0N")
	require.Nil(t, wcs.Workspaces["2bVMV2JiAJe42OXZrzyvJI75v0N"], "The workspace should have not been updated")
	require.Len(t, receivedUpdatedAfter, 3)
	expectedUpdatedAfter, err = time.Parse(updatedAfterTimeFormat, "2024-11-27T20:15:30.647Z")
	require.NoError(t, err)
	require.Equal(t, receivedUpdatedAfter[2], expectedUpdatedAfter, updatedAfterTimeFormat)

	// send the request again, the updatedAfter should be the same as the last request since no updates
	wcs = &modelv2.WorkspaceConfigs{} // reset the workspace configs
	err = cpSDK.GetWorkspaceConfigs(ctx, wcs, latestUpdatedAt)
	require.NoError(t, err)
	latestUpdatedAt, updatedAt = getLatestUpdatedAt(wcs)
	require.Truef(t, updatedAt.IsZero(), "%+v", wcs)
	require.Equal(t, "2024-11-27T20:15:30.647Z", latestUpdatedAt.Format(updatedAfterTimeFormat))
	require.Len(t, wcs.Workspaces, 2)
	require.Contains(t, wcs.Workspaces, "2hCBi02C8xYS8Rsy1m9bJjTlKy6")
	require.Nil(t, wcs.Workspaces["2hCBi02C8xYS8Rsy1m9bJjTlKy6"], "The workspace should have not been updated")
	require.Contains(t, wcs.Workspaces, "2bVMV2JiAJe42OXZrzyvJI75v0N")
	require.Nil(t, wcs.Workspaces["2bVMV2JiAJe42OXZrzyvJI75v0N"], "The workspace should have not been updated")
	require.Len(t, receivedUpdatedAfter, 4)
	expectedUpdatedAfter, err = time.Parse(updatedAfterTimeFormat, "2024-11-27T20:15:30.647Z")
	require.NoError(t, err)
	require.Equal(t, receivedUpdatedAfter[3], expectedUpdatedAfter, updatedAfterTimeFormat)

	// last request, ideally the application should detect that there is an inconsistency and trigger a full update
	// although that behaviour is not tested here
	wcs = &modelv2.WorkspaceConfigs{} // reset the workspace configs
	err = cpSDK.GetWorkspaceConfigs(ctx, wcs, latestUpdatedAt)
	require.NoError(t, err)
	latestUpdatedAt, updatedAt = getLatestUpdatedAt(wcs)
	require.Truef(t, updatedAt.IsZero(), "%+v", wcs)
	require.Equal(t, "2024-11-27T20:15:30.647Z", latestUpdatedAt.Format(updatedAfterTimeFormat))
	require.Len(t, wcs.Workspaces, 1)
	require.Contains(t, wcs.Workspaces, "someWorkspaceID")
	require.Nil(t, wcs.Workspaces["someWorkspaceID"], "The workspace should have not been updated")
	require.Len(t, receivedUpdatedAfter, 5)
	expectedUpdatedAfter, err = time.Parse(updatedAfterTimeFormat, "2024-11-27T20:15:30.647Z")
	require.NoError(t, err)
	require.Equal(t, receivedUpdatedAfter[4], expectedUpdatedAfter, updatedAfterTimeFormat)
}

func getLatestUpdatedAt() func(list diff.UpdateableObject[string]) (time.Time, time.Time) {
	var latestUpdatedAt time.Time
	return func(obj diff.UpdateableObject[string]) (time.Time, time.Time) {
		var localUpdatedAt time.Time
		for uo := range obj.Updateables() {
			for _, wc := range uo.List() {
				if wc.IsNil() || wc.GetUpdatedAt().IsZero() {
					continue
				}
				if wc.GetUpdatedAt().After(latestUpdatedAt) {
					latestUpdatedAt = wc.GetUpdatedAt()
				}
				if wc.GetUpdatedAt().After(localUpdatedAt) {
					localUpdatedAt = wc.GetUpdatedAt()
				}
			}
		}
		return latestUpdatedAt, localUpdatedAt
	}
}
