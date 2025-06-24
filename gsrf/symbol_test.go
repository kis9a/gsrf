package gsrf

import (
	"testing"
)

func TestSymbol_Format(t *testing.T) {
	tests := []struct {
		name     string
		symbol   *Symbol
		expected string
	}{
		// Basic formatting
		{
			name: "simple function",
			symbol: &Symbol{
				PackagePath: "fmt",
				Name:        "Println",
			},
			expected: "fmt.Println",
		},
		{
			name: "method",
			symbol: &Symbol{
				PackagePath: "net/http",
				Name:        "ServeHTTP",
				Receiver: &Receiver{
					TypeName:  "Server",
					IsPointer: true,
				},
			},
			expected: "net/http.(*Server).ServeHTTP",
		},
		{
			name: "anonymous function",
			symbol: &Symbol{
				PackagePath: "main",
				Name:        "main",
				IsAnonymous: true,
			},
			expected: "main.main·lit",
		},
		{
			name: "numbered anonymous function",
			symbol: &Symbol{
				PackagePath: "main",
				Name:        "handler",
				IsAnonymous: true,
				AnonIndex:   3,
			},
			expected: "main.handler·lit3",
		},
		{
			name: "generic function",
			symbol: &Symbol{
				PackagePath: "slices",
				Name:        "Sort",
				TypeArgs:    []string{"int"},
			},
			expected: "slices.Sort[int]",
		},
		{
			name: "generic with constraints",
			symbol: &Symbol{
				PackagePath: "pkg",
				Name:        "Process",
				TypeParams: []TypeParam{
					{Name: "T", Constraint: "comparable"},
					{Name: "U", Constraint: "any"},
				},
			},
			expected: "pkg.Process[T comparable, U]",
		},
		{
			name: "generic receiver",
			symbol: &Symbol{
				PackagePath: "container",
				Name:        "Push",
				Receiver: &Receiver{
					TypeName:  "Stack",
					IsPointer: true,
					TypeArgs:  []string{"*Node"},
				},
			},
			expected: "container.(*Stack[*Node]).Push",
		},
		{
			name: "with context",
			symbol: &Symbol{
				PackagePath: "syscall",
				Name:        "Open",
				Context:     "darwin",
			},
			expected: "syscall.Open@darwin",
		},
		{
			name: "with metadata",
			symbol: &Symbol{
				PackagePath: "io",
				Name:        "Write",
				Receiver: &Receiver{
					TypeName:  "MultiWriter",
					IsPointer: true,
				},
				Metadata: Metadata{
					Via:      "Writer",
					Position: "multi.go:25:1",
				},
			},
			expected: "io.(*MultiWriter).Write{via:Writer,pos:multi.go:25:1}",
		},
		{
			name: "complex with all features",
			symbol: &Symbol{
				PackagePath: "pkg",
				Name:        "Handle",
				Receiver: &Receiver{
					TypeName:  "Server",
					IsPointer: true,
					TypeArgs:  []string{"T", "U"},
				},
				TypeArgs: []string{"string", "int"},
				Context:  "linux",
				Metadata: Metadata{
					Via:      "BaseServer[T, U]",
					Position: "server.go:100:5",
					Custom: map[string]string{
						"deprecated": "true",
					},
				},
			},
			expected: "pkg.(*Server[T, U]).Handle[string, int]@linux{via:BaseServer[T, U],pos:server.go:100:5,deprecated:true}",
		},
		{
			name: "init function",
			symbol: &Symbol{
				PackagePath: "database/sql",
				Name:        "init",
				IsInit:      true,
			},
			expected: "database/sql.init",
		},
		{
			name: "type params without constraint",
			symbol: &Symbol{
				PackagePath: "pkg",
				Name:        "Map",
				TypeParams: []TypeParam{
					{Name: "K"},
					{Name: "V"},
				},
			},
			expected: "pkg.Map[K, V]",
		},
		{
			name: "empty metadata fields",
			symbol: &Symbol{
				PackagePath: "pkg",
				Name:        "Func",
				Metadata:    Metadata{}, // All empty
			},
			expected: "pkg.Func",
		},
		{
			name: "only custom metadata",
			symbol: &Symbol{
				PackagePath: "pkg",
				Name:        "Func",
				Metadata: Metadata{
					Custom: map[string]string{
						"test": "value",
					},
				},
			},
			expected: "pkg.Func{test:value}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.symbol.Format()
			if got != tt.expected {
				t.Errorf("Format() = %v, want %v", got, tt.expected)
			}

			// Test String() method as well
			gotString := tt.symbol.String()
			if gotString != tt.expected {
				t.Errorf("String() = %v, want %v", gotString, tt.expected)
			}
		})
	}
}