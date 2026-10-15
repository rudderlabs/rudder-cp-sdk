package cpsdk

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/testhelper/httptest"

	"github.com/rudderlabs/rudder-cp-sdk/diff"
	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
	"github.com/rudderlabs/rudder-cp-sdk/poller"
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

		require.Equal(t, "gzip", r.Header.Get("Accept-Encoding"))
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer func() {
			require.NoError(t, gz.Close())
		}()
		_, err = gz.Write(responseBody)
		require.NoError(t, err)
	}))
	defer ts.Close()

	cpSDK, err := New(
		WithBaseUrl(ts.URL),
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

func TestGetNamespaceWorkspaces(t *testing.T) {
	t.Run("workspace identity", func(t *testing.T) {
		cpSDK, err := New(
			WithBaseUrl("http://localhost"),
			WithWorkspaceIdentity("token"),
		)
		require.NoError(t, err)

		_, err = cpSDK.GetNamespaceWorkspaces(context.Background())
		require.ErrorIs(t, err, ErrUnsupportedOperation)
	})

	t.Run("namespace identity", func(t *testing.T) {
		var (
			namespace = "test-namespace"
			secret    = "test-secret"
		)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, _, ok := r.BasicAuth()
			require.True(t, ok)
			require.Equal(t, secret, user)

			require.Equal(t, "/configuration/v2/namespaces/"+namespace+"/workspace-ids", r.URL.Path)

			responseBody := []byte(`{"data": ["ws1", "ws2", "ws3"]}`)
			require.Equal(t, "gzip", r.Header.Get("Accept-Encoding"))
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer gz.Close()
			_, err := gz.Write(responseBody)
			require.NoError(t, err)
			require.NoError(t, gz.Close())
			_, err = w.Write(responseBody)
			require.NoError(t, err)
		}))
		defer ts.Close()

		cpSDK, err := New(
			WithBaseUrl(ts.URL),
			WithNamespaceIdentity(namespace, secret),
		)
		require.NoError(t, err)

		workspaceIDs, err := cpSDK.GetNamespaceWorkspaces(context.Background())
		require.NoError(t, err)
		require.Equal(t, []string{"ws1", "ws2", "ws3"}, workspaceIDs)
	})

	t.Run("invalid response code", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()
		cpSDK, err := New(
			WithBaseUrl(ts.URL),
			WithNamespaceIdentity("ns", "secret"),
		)
		require.NoError(t, err)
		_, err = cpSDK.GetNamespaceWorkspaces(context.Background())
		require.Error(t, err)
		var unexpectedStatusErr *UnexpectedStatusCodeError
		require.ErrorAs(t, err, &unexpectedStatusErr)
		require.Equal(t, http.StatusInternalServerError, unexpectedStatusErr.StatusCode)
	})
}

func TestSecrets(t *testing.T) {
	newServerCapturingSecrets := func(t *testing.T, secrets *[]string) *httptest.Server {
		t.Helper()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*secrets = r.URL.Query()["secrets"]
			_, err := w.Write([]byte(`{"workspaces":{}}`))
			require.NoError(t, err)
		}))
		t.Cleanup(ts.Close)
		return ts
	}

	t.Run("namespace identity", func(t *testing.T) {
		for _, tc := range []struct {
			name     string
			options  []Option
			expected []string
		}{
			{name: "not set, control plane default applies", expected: nil},
			{name: "embed", options: []Option{WithSecrets(SecretsEmbed)}, expected: []string{"embed"}},
			{name: "omit", options: []Option{WithSecrets(SecretsOmit)}, expected: []string{"omit"}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				var receivedSecrets []string
				ts := newServerCapturingSecrets(t, &receivedSecrets)

				cpSDK, err := New(append([]Option{
					WithBaseUrl(ts.URL),
					WithNamespaceIdentity("test-namespace", "test-secret"),
				}, tc.options...)...)
				require.NoError(t, err)

				require.NoError(t, cpSDK.GetWorkspaceConfigs(context.Background(), &modelv2.WorkspaceConfigs{}, time.Time{}))
				require.Equal(t, tc.expected, receivedSecrets)
			})
		}
	})

	t.Run("workspace identity", func(t *testing.T) {
		var receivedSecrets []string
		ts := newServerCapturingSecrets(t, &receivedSecrets)

		cpSDK, err := New(
			WithBaseUrl(ts.URL),
			WithWorkspaceIdentity("token"),
			WithSecrets(SecretsEmbed),
		)
		require.NoError(t, err)

		require.NoError(t, cpSDK.GetWorkspaceConfigs(context.Background(), &modelv2.WorkspaceConfigs{}, time.Time{}))
		require.Equal(t, []string{"embed"}, receivedSecrets)
	})

	t.Run("invalid value", func(t *testing.T) {
		_, err := New(
			WithBaseUrl("http://localhost"),
			WithNamespaceIdentity("test-namespace", "test-secret"),
			WithSecrets("invalid"),
		)
		require.ErrorContains(t, err, `invalid secrets option: "invalid"`)
	})
}

func TestIncrementalDefinitions(t *testing.T) {
	t.Run("request", func(t *testing.T) {
		for _, tc := range []struct {
			name             string
			enabled          bool
			object           any
			expectedRequests []string // opt-in header of each request
		}{
			{name: "disabled", object: &modelv2.WorkspaceConfigs{}, expectedRequests: []string{""}},
			{name: "supported object", enabled: true, object: &modelv2.WorkspaceConfigs{}, expectedRequests: []string{"true"}},
			{name: "receiver that is not an updateable object", enabled: true, object: &receiver{}, expectedRequests: []string{"true"}},
			{name: "object that is not a receiver", enabled: true, object: &map[string]any{}},
			{name: "non omittable definitions", enabled: true, object: &nonOmittableReceiver{}},
		} {
			for name, identity := range map[string]Option{
				"namespace identity": WithNamespaceIdentity("test-namespace", "test-secret"),
				"workspace identity": WithWorkspaceIdentity("token"),
			} {
				t.Run(tc.name+"/"+name, func(t *testing.T) {
					var requests []string
					ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						requests = append(requests, r.Header.Get("X-Rudder-Incremental-Definitions"))
						_, _ = w.Write([]byte(`{"workspaces":{}}`))
					}))
					defer ts.Close()

					cpSDK, err := New(WithBaseUrl(ts.URL), identity, WithIncrementalDefinitions(tc.enabled))
					require.NoError(t, err)

					err = cpSDK.GetWorkspaceConfigs(context.Background(), tc.object, time.Time{})
					require.Equal(t, tc.expectedRequests, requests)
					if tc.expectedRequests == nil {
						_, permanent := errors.AsType[*PermanenentError](err)
						require.True(t, permanent, "unsupported objects should not be retried")
						return
					}
					require.NoError(t, err)
				})
			}
		}
	})

	t.Run("response", func(t *testing.T) {
		const cursor = "2026-10-14T10:00:00.000Z"
		var (
			confirmed      = map[string]string{"X-Rudder-Incremental-Definitions": "true", "X-Rudder-Config-Updated-At": cursor}
			sentinel       = state(true, time.Unix(1, 0))
			confirmedState = state(true, time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC))
			srcDefs        = modelv2.SourceDefinitions{"webhook": {Name: "webhook"}}
		)
		for _, tc := range []struct {
			name            string
			status          int
			headers         map[string]string
			definitions     string
			expectedErr     string
			expectedState   diff.IncrementalDefinitionsState
			expectedSrcDefs modelv2.SourceDefinitions
		}{
			{name: "old server", definitions: `{"webhook":{"name":"webhook"}}`, expectedSrcDefs: srcDefs},
			{name: "definitions sent", headers: confirmed, definitions: `{"webhook":{"name":"webhook"}}`, expectedState: confirmedState, expectedSrcDefs: srcDefs},
			{name: "definitions unchanged", headers: confirmed, definitions: `null`, expectedState: confirmedState},
			{name: "empty definitions", headers: confirmed, definitions: `{}`, expectedState: confirmedState, expectedSrcDefs: modelv2.SourceDefinitions{}},
			{
				name:          "confirmation without cursor",
				headers:       map[string]string{"X-Rudder-Incremental-Definitions": "true"},
				definitions:   `null`,
				expectedErr:   `invalid X-Rudder-Config-Updated-At header ""`,
				expectedState: sentinel,
			},
			{
				name:          "confirmation with invalid cursor",
				headers:       map[string]string{"X-Rudder-Incremental-Definitions": "true", "X-Rudder-Config-Updated-At": "yesterday"},
				definitions:   `null`,
				expectedErr:   `invalid X-Rudder-Config-Updated-At header "yesterday"`,
				expectedState: sentinel,
			},
			{
				name:            "unexpected confirmation value",
				headers:         map[string]string{"X-Rudder-Incremental-Definitions": "1", "X-Rudder-Config-Updated-At": cursor},
				definitions:     `{"webhook":{"name":"webhook"}}`,
				expectedSrcDefs: srcDefs,
			},
			{
				name:            "cursor without confirmation",
				headers:         map[string]string{"X-Rudder-Config-Updated-At": cursor},
				definitions:     `{"webhook":{"name":"webhook"}}`,
				expectedSrcDefs: srcDefs,
			},
			{
				name:          "unexpected status code",
				status:        http.StatusInternalServerError,
				headers:       confirmed,
				expectedErr:   "unexpected status code: 500",
				expectedState: sentinel,
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					for k, v := range tc.headers {
						w.Header().Set(k, v)
					}
					if tc.status != 0 {
						w.WriteHeader(tc.status)
						return
					}
					_, err := fmt.Fprintf(w, `{"workspaces":{},"sourceDefinitions":%s,"destinationDefinitions":%[1]s}`, tc.definitions)
					assert.NoError(t, err)
				}))
				defer ts.Close()

				cpSDK, err := New(WithBaseUrl(ts.URL), WithNamespaceIdentity("test-namespace", "test-secret"), WithIncrementalDefinitions(true))
				require.NoError(t, err)

				wcs := &modelv2.WorkspaceConfigs{IncrementalDefinitionsState: sentinel}
				err = cpSDK.GetWorkspaceConfigs(context.Background(), wcs, time.Time{})
				require.Equal(t, tc.expectedState, wcs.IncrementalDefinitionsState)
				if tc.expectedErr != "" {
					require.ErrorContains(t, err, tc.expectedErr)
					return
				}
				require.NoError(t, err)
				require.Equal(t, tc.expectedSrcDefs, wcs.SourceDefinitions)
			})
		}
	})
}

func TestIncrementalDefinitionsPolling(t *testing.T) {
	ts := func(minute int) string {
		return time.Date(2026, 10, 14, 10, minute, 0, 0, time.UTC).Format(updatedAfterTimeFormat)
	}
	defs := func(name string) modelv2.SourceDefinitions { return modelv2.SourceDefinitions{"webhook": {Name: name}} }

	type result struct {
		updated bool
		srcDefs modelv2.SourceDefinitions
		err     error
	}
	steps := []struct {
		updatedAfter string // expected in the request
		cursor       string // empty for a legacy response, e.g. kill switch on
		workspaces   string
		definitions  string
		expected     result
	}{
		{ // full fetch
			cursor: ts(1), workspaces: fmt.Sprintf(`{"ws-1":{"updatedAt":%q}}`, ts(1)), definitions: `{"webhook":{"name":"v1"}}`,
			expected: result{updated: true, srcDefs: defs("v1")},
		},
		{ // nothing changed
			updatedAfter: ts(1), cursor: ts(2), workspaces: `{"ws-1":null}`, definitions: `null`,
			expected: result{srcDefs: defs("v1")},
		},
		{ // definitions changed
			updatedAfter: ts(2), cursor: ts(3), workspaces: `{"ws-1":null}`, definitions: `{"webhook":{"name":"v2"}}`,
			expected: result{updated: true, srcDefs: defs("v2")},
		},
		{ // kill switch on
			updatedAfter: ts(3), workspaces: `{"ws-1":null}`, definitions: `{"webhook":{"name":"v2"}}`,
			expected: result{srcDefs: defs("v2")},
		},
		{ // kill switch off
			updatedAfter: ts(3), cursor: ts(4), workspaces: `{"ws-1":null}`, definitions: `null`,
			expected: result{srcDefs: defs("v2")},
		},
		{ // definitions changed during the poll, the cursor is earlier than the new workspace
			updatedAfter: ts(4), cursor: ts(5), workspaces: fmt.Sprintf(`{"ws-1":null,"ws-2":{"updatedAt":%q}}`, ts(6)), definitions: `null`,
			expected: result{updated: true, srcDefs: defs("v2")},
		},
		{ // and are delivered by the next one
			updatedAfter: ts(5), cursor: ts(6), workspaces: `{"ws-1":null,"ws-2":null}`, definitions: `{"webhook":{"name":"v3"}}`,
			expected: result{updated: true, srcDefs: defs("v3")},
		},
	}

	var (
		mu       sync.Mutex
		requests []string // opt-in header and updatedAfter of each request
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		requests = append(requests, r.Header.Get("X-Rudder-Incremental-Definitions")+" "+r.URL.Query().Get("updatedAfter"))
		if len(requests) > len(steps) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		s := steps[len(requests)-1]
		if s.cursor != "" {
			w.Header().Set("X-Rudder-Incremental-Definitions", "true")
			w.Header().Set("X-Rudder-Config-Updated-At", s.cursor)
		}
		_, err := fmt.Fprintf(w, `{"workspaces":%s,"sourceDefinitions":%s,"destinationDefinitions":%[2]s}`, s.workspaces, s.definitions)
		assert.NoError(t, err)
	}))
	defer srv.Close()

	cpSDK, err := New(WithBaseUrl(srv.URL), WithNamespaceIdentity("test-namespace", "test-secret"), WithIncrementalDefinitions(true))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var (
		cache   = &modelv2.WorkspaceConfigs{}
		updater diff.Updater[string]
		results []result
	)
	p, err := poller.NewWorkspaceConfigsPoller(
		func(ctx context.Context, obj diff.UpdateableObject[string], updatedAfter time.Time) error {
			return cpSDK.GetWorkspaceConfigs(ctx, obj, updatedAfter)
		},
		func(obj diff.UpdateableObject[string]) (time.Time, bool, error) {
			updatedAt, updated, err := updater.UpdateCache(obj, cache)
			results = append(results, result{updated: updated, srcDefs: cache.SourceDefinitions, err: err})
			if len(results) == len(steps) {
				cancel()
			}
			return updatedAt, updated, err
		},
		func() diff.UpdateableObject[string] { return &modelv2.WorkspaceConfigs{} },
		poller.WithPollingInterval[string](time.Nanosecond),
	)
	require.NoError(t, err)
	p.Run(ctx)

	var expectedRequests []string
	var expectedResults []result
	for _, s := range steps {
		expectedRequests = append(expectedRequests, "true "+s.updatedAfter)
		expectedResults = append(expectedResults, s.expected)
	}
	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, expectedRequests, requests)
	require.Equal(t, expectedResults, results)
}

type receiver struct {
	diff.IncrementalDefinitionsState
	Workspaces map[string]any `json:"workspaces"`
}

// nonOmittableReceiver is a receiver with a non updateable list that is not an [diff.OmittableList].
type nonOmittableReceiver struct {
	diff.IncrementalDefinitionsState
}

func (*nonOmittableReceiver) Updateables() iter.Seq[diff.UpdateableList[string, diff.UpdateableElement]] {
	return func(func(diff.UpdateableList[string, diff.UpdateableElement]) bool) {}
}

func (*nonOmittableReceiver) NonUpdateables() iter.Seq[diff.NonUpdateablesList[string, any]] {
	return func(yield func(diff.NonUpdateablesList[string, any]) bool) {
		yield(struct {
			diff.NonUpdateablesList[string, any]
		}{})
	}
}

func state(confirmed bool, cursor time.Time) (s diff.IncrementalDefinitionsState) {
	s.SetIncrementalDefinitions(confirmed, cursor)
	return s
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
