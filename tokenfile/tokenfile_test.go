package tokenfile

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		tok := New(filepath.Join(t.TempDir(), "missing"))
		require.ErrorIs(t, tok.Load(), fs.ErrNotExist)
		require.Empty(t, tok.Get(), "nothing was read")
	})

	t.Run("empty file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "token")
		require.NoError(t, os.WriteFile(path, []byte(" \n"), 0o600))
		tok := New(path)
		require.Error(t, tok.Load())
		require.Empty(t, tok.Get(), "nothing was read")
	})

	t.Run("trims whitespace", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "token")
		require.NoError(t, os.WriteFile(path, []byte(" token-1\n"), 0o600))
		tok := New(path)
		require.NoError(t, tok.Load())
		require.Equal(t, "token-1", tok.Get())
	})
}

func TestGet(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "token")
		write := func(content string) {
			t.Helper()
			require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
		}
		write("token-1")
		tok := New(path)
		require.NoError(t, tok.Load())

		write("token-2")
		time.Sleep(maxAge - time.Nanosecond)
		require.Equal(t, "token-1", tok.Get(), "reused while under a minute old")

		time.Sleep(time.Nanosecond)
		require.Equal(t, "token-2", tok.Get(), "read again once a minute old")

		write(" \n")
		time.Sleep(maxAge)
		require.Equal(t, "token-2", tok.Get(), "an empty file keeps the last token")

		require.NoError(t, os.Remove(path))
		require.Equal(t, "token-2", tok.Get(), "a failed read keeps the last token")

		write("token-3")
		require.Equal(t, "token-3", tok.Get(), "a failed read is retried on the next call")
	})
}

func TestGetConcurrent(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "token")
		require.NoError(t, os.WriteFile(path, []byte("token-0"), 0o600))
		tok := New(path)
		require.NoError(t, tok.Load())

		var wg sync.WaitGroup
		wg.Go(func() {
			// rotate, empty and remove the file while the readers below run
			for i := range 300 {
				switch i % 3 {
				case 0:
					assert.NoError(t, os.WriteFile(path, []byte("token-"+strconv.Itoa(i)), 0o600))
				case 1:
					assert.NoError(t, os.WriteFile(path, nil, 0o600))
				case 2:
					assert.NoError(t, os.Remove(path))
				}
				time.Sleep(10 * time.Second)
			}
		})
		for range 10 {
			wg.Go(func() {
				for range 300 {
					assert.NotEmpty(t, tok.Get())
					time.Sleep(10 * time.Second)
				}
			})
		}
		wg.Wait()
	})
}
