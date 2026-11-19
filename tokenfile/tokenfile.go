// Package tokenfile reads a token from a file that is rewritten when the token rotates, such as a projected
// service account token, and caches it between reads.
package tokenfile

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// maxAge is how long a token read from the file is reused. The kubelet rewrites the file at about 80% of the token's
// lifetime (8 of 10 minutes by default), so a token read under a minute ago is still valid.
const maxAge = time.Minute

// Token is a token read from a file. It is safe for concurrent use.
type Token struct {
	path string

	mu     sync.Mutex
	token  string
	readAt time.Time
}

// New returns a Token read from the file at path. It does not read the file; see [Token.Load].
func New(path string) *Token {
	return &Token{path: path}
}

// Load reads the file. It returns an error if the file can't be read or holds only whitespace.
func (t *Token) Load() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.load()
}

func (t *Token) load() error {
	data, err := os.ReadFile(t.path)
	if err != nil {
		return err
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return fmt.Errorf("token file %q is empty", t.path)
	}
	t.token, t.readAt = token, time.Now()
	return nil
}

// Get returns the token, reading the file again once the token is a minute old. A failed or empty read keeps the
// last token, which is retried on the next call, so Get never returns an empty token after a successful [Token.Load].
func (t *Token) Get() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if time.Since(t.readAt) >= maxAge {
		if err := t.load(); err != nil {
			return t.token // the last token; readAt is unchanged, so the next call reads the file again
		}
	}
	return t.token
}
