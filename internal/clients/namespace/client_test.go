package namespace

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCachedToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token")
	now := time.Now()
	tok := cachedToken{now: func() time.Time { return now }}

	require.Empty(t, tok.get(path), "no file yet")
	require.Empty(t, tok.get(""), "no path")

	require.NoError(t, os.WriteFile(path, []byte("token-1\n"), 0o600))
	require.Equal(t, "token-1", tok.get(path), "a missing file is retried on the next call")

	require.NoError(t, os.WriteFile(path, []byte("token-2"), 0o600))
	now = now.Add(identityTokenMaxAge - time.Second)
	require.Equal(t, "token-1", tok.get(path), "reused while under a minute old")

	now = now.Add(time.Second)
	require.Equal(t, "token-2", tok.get(path), "re-read once a minute old")

	require.NoError(t, os.WriteFile(path, []byte(" \n"), 0o600))
	now = now.Add(identityTokenMaxAge)
	require.Equal(t, "token-2", tok.get(path), "an empty file keeps the last good token")

	require.NoError(t, os.Remove(path))
	require.Equal(t, "token-2", tok.get(path), "a failed read keeps the last good token")

	require.NoError(t, os.WriteFile(path, []byte("token-3"), 0o600))
	require.Equal(t, "token-3", tok.get(path), "a failed read leaves the cache expired, so the next call reads again")
}
