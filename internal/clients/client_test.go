package clients

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/rudderlabs/rudder-cp-sdk/identity"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/base"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/namespace"
	"github.com/rudderlabs/rudder-cp-sdk/internal/clients/workspace"
	"github.com/rudderlabs/rudder-go-kit/testhelper/rand"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	t.Run("workspace client", func(t *testing.T) {
		ctx := context.Background()

		tpl, err := template.ParseFiles("./testdata/workspaceConfigTemplate.json")
		require.NoError(t, err)

		writeKey := rand.String(27)
		workspaceID := rand.String(27)
		sourceID := rand.String(27)
		destinationID := rand.String(27)
		webhookURL := "http://localhost:8080/v1/batch"
		workspaceToken := rand.String(27)

		workspaceConfig := bytes.NewBuffer(nil)
		require.NoError(t, tpl.Execute(workspaceConfig, map[string]any{
			"webhookUrl":    webhookURL,
			"writeKey":      writeKey,
			"workspaceId":   workspaceID,
			"sourceID":      sourceID,
			"destinationID": destinationID,
			"updatedAt":     time.Time{}.Format(time.RFC3339),
		}))
		backendConfigSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, _, ok := r.BasicAuth()
			require.True(t, ok)
			require.Equalf(t, workspaceToken, u, "Expected HTTP basic authentication to be %q, got %q instead", workspaceToken, u)

			switch r.URL.String() {
			case "/data-plane/v2/workspaceConfig":
				n, err := w.Write(workspaceConfig.Bytes())
				require.NoError(t, err)
				require.Equal(t, workspaceConfig.Len(), n)
			default:
				require.FailNowf(t, "BackendConfig", "Unexpected %s to BackendConfig, not found: %+v", r.Method, r.URL)
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer backendConfigSrv.Close()

		parsedURL, _ := url.Parse(backendConfigSrv.URL)
		baseClient := &base.Client{HTTPClient: http.DefaultClient, BaseURL: parsedURL}
		wc := &workspace.Client{
			Client: baseClient,
			Identity: &identity.Workspace{
				WorkspaceToken: workspaceToken,
			},
		}

		configs, err := wc.GetWorkspaceConfigs(ctx)
		require.NoError(t, err)

		require.NotEmpty(t, configs)
		require.Len(t, configs.Workspaces, 1)
		require.Contains(t, configs.Workspaces, workspaceID)
		require.Len(t, configs.Workspaces[workspaceID].Sources, 1)
		require.Contains(t, configs.Workspaces[workspaceID].Sources, sourceID)
		require.Equal(t, configs.Workspaces[workspaceID].Sources[sourceID].WriteKey, writeKey)
		require.Len(t, configs.Workspaces[workspaceID].Destinations, 1)
		require.Contains(t, configs.Workspaces[workspaceID].Destinations, destinationID)
		require.Equal(t, configs.Workspaces[workspaceID].Destinations[destinationID].DefinitionName, "WEBHOOK")
		require.Len(t, configs.SourceDefinitions, 1)
		require.Len(t, configs.DestinationDefinitions, 1)
		require.Empty(t, configs.UpdatedAt())

		configs, err = wc.GetUpdatedWorkspaceConfigs(ctx, time.Now())
		require.Error(t, err)
		require.Empty(t, configs)
	})
	t.Run("namespace client", func(t *testing.T) {
		t.Run("without updatedAt", func(t *testing.T) {
			ctx := context.Background()

			tpl, err := template.ParseFiles("./testdata/workspaceConfigTemplate.json")
			require.NoError(t, err)

			writeKey := rand.String(27)
			testNamespace := "some-test-namespace"
			workspaceID := rand.String(27)
			sourceID := rand.String(27)
			destinationID := rand.String(27)
			webhookURL := "http://localhost:8080/v1/batch"
			hostedSecret := rand.String(27)

			workspaceConfig := bytes.NewBuffer(nil)
			require.NoError(t, tpl.Execute(workspaceConfig, map[string]any{
				"webhookUrl":    webhookURL,
				"writeKey":      writeKey,
				"workspaceId":   workspaceID,
				"sourceID":      sourceID,
				"destinationID": destinationID,
				"updatedAt":     time.Time{}.Format(time.RFC3339),
			}))

			backendConfigSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				u, _, ok := r.BasicAuth()
				require.True(t, ok)
				require.Equalf(t, hostedSecret, u, "Expected HTTP basic authentication to be %q, got %q instead", hostedSecret, u)

				switch r.URL.String() {
				case "/data-plane/v2/namespaces/" + testNamespace + "/config":
					n, err := w.Write(workspaceConfig.Bytes())
					require.NoError(t, err)
					require.Equal(t, workspaceConfig.Len(), n)
				default:
					require.FailNowf(t, "BackendConfig", "Unexpected %s to BackendConfig, not found: %+v", r.Method, r.URL)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer backendConfigSrv.Close()

			parsedURL, _ := url.Parse(backendConfigSrv.URL)
			baseClient := &base.Client{HTTPClient: http.DefaultClient, BaseURL: parsedURL}
			wc := &namespace.Client{
				Client: baseClient,
				Identity: &identity.Namespace{
					Namespace: testNamespace,
					Secret:    hostedSecret,
				},
			}

			configs, err := wc.GetWorkspaceConfigs(ctx)
			require.NoError(t, err)

			require.NotEmpty(t, configs)
			require.Len(t, configs.Workspaces, 1)
			require.Contains(t, configs.Workspaces, workspaceID)
			require.Len(t, configs.Workspaces[workspaceID].Sources, 1)
			require.Contains(t, configs.Workspaces[workspaceID].Sources, sourceID)
			require.Equal(t, configs.Workspaces[workspaceID].Sources[sourceID].WriteKey, writeKey)
			require.Len(t, configs.Workspaces[workspaceID].Destinations, 1)
			require.Contains(t, configs.Workspaces[workspaceID].Destinations, destinationID)
			require.Equal(t, configs.Workspaces[workspaceID].Destinations[destinationID].DefinitionName, "WEBHOOK")
			require.Len(t, configs.SourceDefinitions, 1)
			require.Len(t, configs.DestinationDefinitions, 1)
			require.Empty(t, configs.UpdatedAt())
		})
		t.Run("with updatedAt", func(t *testing.T) {
			ctx := context.Background()

			tpl, err := template.ParseFiles("./testdata/workspaceConfigTemplate.json")
			require.NoError(t, err)

			writeKey := rand.String(27)
			testNamespace := "some-test-namespace"
			workspaceID := rand.String(27)
			sourceID := rand.String(27)
			destinationID := rand.String(27)
			webhookURL := "http://localhost:8080/v1/batch"
			hostedSecret := rand.String(27)

			updatedAt := time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC)
			updatedWorkspaceConfig := bytes.NewBuffer(nil)
			require.NoError(t, tpl.Execute(updatedWorkspaceConfig, map[string]any{
				"webhookUrl":    webhookURL,
				"writeKey":      writeKey,
				"workspaceId":   workspaceID,
				"sourceID":      sourceID,
				"destinationID": destinationID,
				"updatedAt":     updatedAt.Format(time.RFC3339),
			}))

			namespaceURL := "/data-plane/v2/namespaces/" + testNamespace + "/config"

			backendConfigSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				u, _, ok := r.BasicAuth()
				require.True(t, ok)
				require.Equalf(t, hostedSecret, u, "Expected HTTP basic authentication to be %q, got %q instead", hostedSecret, u)

				switch {
				case strings.Contains(r.URL.String(), namespaceURL) && strings.Contains(r.URL.String(), "updatedAfter"):
					n, err := w.Write(updatedWorkspaceConfig.Bytes())
					require.NoError(t, err)
					require.Equal(t, updatedWorkspaceConfig.Len(), n)
				default:
					require.FailNowf(t, "BackendConfig", "Unexpected %s to BackendConfig, not found: %+v", r.Method, r.URL)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer backendConfigSrv.Close()

			parsedURL, _ := url.Parse(backendConfigSrv.URL)
			baseClient := &base.Client{HTTPClient: http.DefaultClient, BaseURL: parsedURL}
			wc := &namespace.Client{
				Client: baseClient,
				Identity: &identity.Namespace{
					Namespace: testNamespace,
					Secret:    hostedSecret,
				},
			}

			configs, err := wc.GetUpdatedWorkspaceConfigs(ctx, updatedAt)
			require.NoError(t, err)
			require.NotEmpty(t, configs)
			require.Len(t, configs.Workspaces, 1)
			require.Contains(t, configs.Workspaces, workspaceID)
			require.Len(t, configs.Workspaces[workspaceID].Sources, 1)
			require.Contains(t, configs.Workspaces[workspaceID].Sources, sourceID)
			require.Equal(t, configs.Workspaces[workspaceID].Sources[sourceID].WriteKey, writeKey)
			require.Len(t, configs.Workspaces[workspaceID].Destinations, 1)
			require.Contains(t, configs.Workspaces[workspaceID].Destinations, destinationID)
			require.Equal(t, configs.Workspaces[workspaceID].Destinations[destinationID].DefinitionName, "WEBHOOK")
			require.Len(t, configs.SourceDefinitions, 1)
			require.Len(t, configs.DestinationDefinitions, 1)
			require.EqualValues(t, updatedAt.UTC().UTC(), configs.UpdatedAt().UTC())
		})
	})
}
