package profiler

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/rudderlabs/rudder-go-kit/logger"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func MemProfile(log logger.Logger) {
	f, _ := os.Create(fmt.Sprintf("heap-%d.pprof", r.Int()))
	runtime.GC()
	_ = pprof.WriteHeapProfile(f)
	_ = f.Close()

	log.Infof("heap profile written to %s", f.Name())
}
