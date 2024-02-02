package benchmark

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cpsdk "github.com/rudderlabs/rudder-cp-sdk"
)

/*
Original: BenchmarkClient/new-20         	       2	 547824234 ns/op	787311288 B/op	 9892879 allocs/op
FastHTTP: BenchmarkClient/new-20         	       1	1230146368 ns/op	528436016 B/op	 9898799 allocs/op
FastHTTPReq: BenchmarkClient/new-20         	   1	1060077492 ns/op	528428248 B/op	 9898792 allocs/op
*/
func BenchmarkClient(b *testing.B) {
	cwd, err := os.Getwd()
	require.NoError(b, err)

	sample, err := filepath.Abs(filepath.Join(cwd, "testdata", "sample.json"))
	require.NoError(b, err)

	data, err := os.ReadFile(sample)
	require.NoError(b, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(data)
	}))
	defer srv.Close()

	options := []cpsdk.Option{
		cpsdk.WithBaseUrl(srv.URL),
		cpsdk.WithPollingInterval(time.Minute),
		cpsdk.WithNamespaceIdentity("foo", "bar"),
	}

	b.ResetTimer()
	b.Run("new", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sdk, err := cpsdk.New(options...)
			require.NoError(b, err)

			ch := sdk.Subscribe()
			<-ch

			wc, err := sdk.GetWorkspaceConfigs()
			require.NoError(b, err)
			require.Len(b, wc.Workspaces, 21831)
		}
	})
}
