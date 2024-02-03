package linter

import (
	"fmt"
	"io"
	"testing"
)

func TestAnalyzeFile(t *testing.T) {
	err := analyzeFile("./testdata/structs.go")
	if err != nil {
		t.Errorf("analyzeFile failed with error: %v", err)
	}
}

func BenchmarkMemory(b *testing.B) {
	b.Run("unoptimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			t := TerraformResourceUnoptimized{}
			_, _ = fmt.Fprintf(io.Discard, "%v", t)
		}
	})
	b.Run("optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			t := TerraformResourceOptimized{}
			_, _ = fmt.Fprintf(io.Discard, "%v", t)
		}
	})
}

type TerraformResourceUnoptimized struct {
	Cloud               string // 16 bytes
	Name                string // 16 bytes
	HaveDSL             bool   //  1 byte
	PluginVersion       string // 16 bytes
	IsVersionControlled bool   //  1 byte
	TerraformVersion    string // 16 bytes
	ModuleVersionMajor  int32  //  4 bytes
}

type TerraformResourceOptimized struct {
	Cloud               string // 16 bytes
	Name                string // 16 bytes
	PluginVersion       string // 16 bytes
	TerraformVersion    string // 16 bytes
	ModuleVersionMajor  int32  //  4 bytes
	HaveDSL             bool   //  1 byte
	IsVersionControlled bool   //  1 byte
}
