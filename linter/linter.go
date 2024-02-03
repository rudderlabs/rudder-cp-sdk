package linter

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"unsafe"
)

// fieldInfo holds information about a struct's field.
type fieldInfo struct {
	Name string
	Size int
}

// BySize implements sort.Interface for []fieldInfo based on
// the Size field in descending order.
type BySize []fieldInfo

func (a BySize) Len() int           { return len(a) }
func (a BySize) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySize) Less(i, j int) bool { return a[i].Size > a[j].Size }

// analyzeFile analyzes the Go source file for struct ordering.
func analyzeFile(filename string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	ast.Inspect(node, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			return true
		}

		// Found a struct, now analyze and suggest ordering
		fmt.Printf("Struct: %s\n", ts.Name.Name)
		suggestFieldOrder(st)

		return true
	})

	return nil
}

// suggestFieldOrder suggests an order for struct fields based on their sizes.
func suggestFieldOrder(st *ast.StructType) {
	var fields []fieldInfo

	for _, field := range st.Fields.List {
		// Simplified field size estimation based on type; you might want to expand this
		var size int
		switch fieldType := field.Type.(type) {
		case *ast.Ident:
			size = estimateSize(fieldType.Name)
			// Add more cases as needed to handle other types like arrays, slices, maps, etc.
		}

		for _, name := range field.Names {
			fields = append(fields, fieldInfo{Name: name.Name, Size: size})
		}
	}

	sort.Sort(BySize(fields))

	fmt.Println("Suggested order:")
	for _, field := range fields {
		fmt.Printf("Field: %s (Size: %d bytes)\n", field.Name, field.Size)
	}
}

// estimateSize returns an estimated size for basic types, inspired by what you might expect from unsafe.Sizeof.
func estimateSize(typeName string) int {
	// This mapping simulates the results of unsafe.Sizeof for basic static types.
	switch typeName {
	case "int64", "uint64":
		return int(unsafe.Sizeof(int64(0))) // 8 bytes on 64-bit
	case "int32", "uint32", "float32":
		return int(unsafe.Sizeof(int32(0))) // 4 bytes
	case "int16", "uint16":
		return int(unsafe.Sizeof(int16(0))) // 2 bytes
	case "int8", "uint8", "byte", "bool":
		return int(unsafe.Sizeof(int8(0))) // 1 byte
	case "string":
		// Strings are a bit special since they are a struct containing a pointer and a length.
		// This size would be the size of the struct, not the size of the content it points to.
		return int(unsafe.Sizeof(string(""))) // Typically 16 bytes on 64-bit (pointer + length)
	case "int", "uint", "uintptr":
		return int(unsafe.Sizeof(int(0))) // 8 bytes on 64-bit architectures, 4 bytes on 32-bit
	default:
		// This case handles pointers, slices, maps, channels, functions, and interfaces,
		// which have the size of a pointer on the current architecture.
		return int(unsafe.Sizeof(uintptr(0))) // Typically 8 bytes on 64-bit, 4 bytes on 32-bit
	}
}
