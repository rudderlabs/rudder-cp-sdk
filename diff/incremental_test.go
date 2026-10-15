package diff_test

import (
	"errors"
	"iter"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-cp-sdk/diff"
	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
)

func TestUpdateCacheIncrementalDefinitions(t *testing.T) {
	var (
		cursor     = time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC)
		ws         = &modelv2.WorkspaceConfig{UpdatedAt: cursor.Add(-time.Hour)}
		srcDefs    = modelv2.SourceDefinitions{"webhook": {Name: "webhook"}}
		newSrcDefs = modelv2.SourceDefinitions{"webhook": {Name: "webhook v2"}}
		dstDefs    = modelv2.DestinationDefinitions{"webhook": {Name: "webhook"}}
	)
	legacy := func(src modelv2.SourceDefinitions, dst modelv2.DestinationDefinitions) *modelv2.WorkspaceConfigs {
		return &modelv2.WorkspaceConfigs{
			Workspaces:             modelv2.Workspaces{"ws-1": ws},
			SourceDefinitions:      src,
			DestinationDefinitions: dst,
		}
	}
	incremental := func(src modelv2.SourceDefinitions, dst modelv2.DestinationDefinitions) *modelv2.WorkspaceConfigs {
		wcs := &modelv2.WorkspaceConfigs{
			Workspaces:             modelv2.Workspaces{"ws-1": nil},
			SourceDefinitions:      src,
			DestinationDefinitions: dst,
		}
		wcs.SetIncrementalDefinitions(true, cursor)
		return wcs
	}

	for _, tc := range []struct {
		name          string
		cache         *modelv2.WorkspaceConfigs // empty if nil
		response      diff.UpdateableObject[string]
		expectedErr   string
		fullFetch     bool
		expectUpdated bool
		expectedCache *modelv2.WorkspaceConfigs
	}{
		{
			name:          "omitted definitions keep the cached copy",
			cache:         legacy(srcDefs, dstDefs),
			response:      incremental(nil, nil),
			expectedCache: legacy(srcDefs, dstDefs),
		},
		{
			name:          "omitted definitions keep an empty cached copy",
			cache:         legacy(modelv2.SourceDefinitions{}, modelv2.DestinationDefinitions{}),
			response:      incremental(nil, nil),
			expectedCache: legacy(modelv2.SourceDefinitions{}, modelv2.DestinationDefinitions{}),
		},
		{
			name:          "sent definitions replace the cached copy",
			cache:         legacy(srcDefs, dstDefs),
			response:      incremental(newSrcDefs, nil),
			expectUpdated: true,
			expectedCache: legacy(newSrcDefs, dstDefs),
		},
		{
			name:          "omitted definitions nil in cache",
			response:      incremental(nil, nil),
			expectedErr:   "was omitted but was nil in cache",
			fullFetch:     true,
			expectedCache: &modelv2.WorkspaceConfigs{},
		},
		{
			name:  "sent definitions with a workspace not in cache",
			cache: legacy(srcDefs, dstDefs),
			response: func() *modelv2.WorkspaceConfigs {
				wcs := incremental(newSrcDefs, nil)
				wcs.Workspaces = modelv2.Workspaces{"ws-2": nil}
				return wcs
			}(),
			expectedErr:   "was not present in cache",
			fullFetch:     true,
			expectedCache: legacy(srcDefs, dstDefs),
		},
		{
			name:          "non omittable definitions",
			cache:         legacy(srcDefs, dstDefs),
			response:      &nonOmittableWorkspaceConfigs{*incremental(newSrcDefs, nil)},
			expectedErr:   "does not implement OmittableList",
			expectedCache: legacy(srcDefs, dstDefs),
		},
		{
			name:          "legacy response with a workspace not in cache",
			response:      &modelv2.WorkspaceConfigs{Workspaces: modelv2.Workspaces{"ws-1": nil}},
			expectedErr:   "was not present in cache",
			fullFetch:     true,
			expectedCache: &modelv2.WorkspaceConfigs{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cache := tc.cache
			if cache == nil {
				cache = &modelv2.WorkspaceConfigs{}
			}

			var updater diff.Updater[string]
			updatedAt, updated, err := updater.UpdateCache(tc.response, cache)
			require.Equal(t, tc.expectedCache, cache)
			if tc.expectedErr != "" {
				require.ErrorContains(t, err, tc.expectedErr)
				require.Equal(t, tc.fullFetch, errors.Is(err, diff.ErrFullFetchRequired))
				return
			}
			require.NoError(t, err)
			require.Equal(t, cursor, updatedAt, "the server cursor should be returned")
			require.Equal(t, tc.expectUpdated, updated)
		})
	}
}

// nonOmittableWorkspaceConfigs hides Omitted from the definitions lists.
type nonOmittableWorkspaceConfigs struct{ modelv2.WorkspaceConfigs }

func (wcs *nonOmittableWorkspaceConfigs) NonUpdateables() iter.Seq[diff.NonUpdateablesList[string, any]] {
	return func(yield func(diff.NonUpdateablesList[string, any]) bool) {
		yield(struct {
			diff.NonUpdateablesList[string, any]
		}{&wcs.SourceDefinitions})
	}
}
