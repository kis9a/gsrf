package adapters

import (
	"testing"

	"github.com/kis9a/gsrf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromStackTrace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *gsrf.Symbol
		wantErr  bool
	}{
		{
			name:  "simple function",
			input: "main.main",
			expected: &gsrf.Symbol{
				PackagePath: "main",
				Name:        "main",
				Metadata:    gsrf.Metadata{},
			},
		},
		{
			name:  "package function",
			input: "github.com/user/repo.Function",
			expected: &gsrf.Symbol{
				PackagePath: "github.com/user/repo",
				Name:        "Function",
				Metadata:    gsrf.Metadata{},
			},
		},
		{
			name:  "pointer receiver method",
			input: "net/http.(*Server).Serve",
			expected: &gsrf.Symbol{
				PackagePath: "net/http",
				Name:        "Serve",
				Receiver: &gsrf.Receiver{
					TypeName:  "Server",
					IsPointer: true,
				},
				Metadata: gsrf.Metadata{},
			},
		},
		{
			name:  "init function",
			input: "pkg.init.func1",
			expected: &gsrf.Symbol{
				PackagePath: "pkg",
				Name:        "init",
				IsInit:      true,
				Metadata:    gsrf.Metadata{},
			},
		},
		{
			name:  "anonymous function",
			input: "main.main.func1",
			expected: &gsrf.Symbol{
				PackagePath: "main",
				Name:        "main",
				IsAnonymous: true,
				AnonParent:  "main.main",
				Metadata:    gsrf.Metadata{},
			},
		},
		{
			name:  "function with generics",
			input: "pkg.Map[int, string]",
			expected: &gsrf.Symbol{
				PackagePath: "pkg",
				Name:        "Map",
				TypeArgs:    []string{"int", "string"},
				Metadata:    gsrf.Metadata{},
			},
		},
		{
			name:  "method with generic receiver",
			input: "pkg.(*List[T]).Add",
			expected: &gsrf.Symbol{
				PackagePath: "pkg",
				Name:        "Add",
				Receiver: &gsrf.Receiver{
					TypeName:  "List",
					IsPointer: true,
					TypeArgs:  []string{"T"},
				},
				Metadata: gsrf.Metadata{},
			},
		},
		{
			name:  "complex generic receiver",
			input: "pkg.(*Map[K, V]).Get",
			expected: &gsrf.Symbol{
				PackagePath: "pkg",
				Name:        "Get",
				Receiver: &gsrf.Receiver{
					TypeName:  "Map",
					IsPointer: true,
					TypeArgs:  []string{"K", "V"},
				},
				Metadata: gsrf.Metadata{},
			},
		},
		{
			name:  "function with file info",
			input: "main.main /path/to/file.go:123",
			expected: &gsrf.Symbol{
				PackagePath: "main",
				Name:        "main",
				Metadata:    gsrf.Metadata{},
			},
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromStackTrace(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestToStackTrace(t *testing.T) {
	tests := []struct {
		name     string
		symbol   *gsrf.Symbol
		expected string
	}{
		{
			name: "simple function",
			symbol: &gsrf.Symbol{
				PackagePath:  "main",
				Name: "main",
			},
			expected: "main.main",
		},
		{
			name: "package function",
			symbol: &gsrf.Symbol{
				PackagePath:  "github.com/user/repo",
				Name: "Function",
			},
			expected: "github.com/user/repo.Function",
		},
		{
			name: "pointer receiver method",
			symbol: &gsrf.Symbol{
				PackagePath:  "net/http",
				Name: "Serve",
				Receiver: &gsrf.Receiver{
					TypeName:      "Server",
					IsPointer: true,
				},
			},
			expected: "net/http.(*Server).Serve",
		},
		{
			name: "value receiver method (converts to pointer)",
			symbol: &gsrf.Symbol{
				PackagePath:  "pkg",
				Name: "Method",
				Receiver: &gsrf.Receiver{
					TypeName:      "Type",
					IsPointer: false,
				},
			},
			expected: "pkg.(*Type).Method",
		},
		{
			name: "init function",
			symbol: &gsrf.Symbol{
				PackagePath: "pkg",
				IsInit:  true,
			},
			expected: "pkg.init.func1",
		},
		{
			name: "anonymous function",
			symbol: &gsrf.Symbol{
				PackagePath:     "main",
				IsAnonymous: true,
				AnonParent:  "main.main",
				AnonIndex:   2,
			},
			expected: "main.main.func2",
		},
		{
			name: "function with generics",
			symbol: &gsrf.Symbol{
				PackagePath:    "pkg",
				Name:   "Map",
				TypeArgs: []string{"int", "string"},
			},
			expected: "pkg.Map[int, string]",
		},
		{
			name: "method with generic receiver",
			symbol: &gsrf.Symbol{
				PackagePath:  "pkg",
				Name: "Add",
				Receiver: &gsrf.Receiver{
					TypeName:      "List",
					IsPointer: true,
					TypeArgs:  []string{"T"},
				},
			},
			expected: "pkg.(*List[T]).Add",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToStackTrace(tt.symbol)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestStackTraceRoundTrip(t *testing.T) {
	// Test conversions that should round-trip successfully
	inputs := []string{
		"main.main",
		"pkg.Function",
		"net/http.(*Server).Serve",
		"pkg.init.func1",
		"pkg.Map[int, string]",
		"pkg.(*List[T]).Add",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			sym, err := FromStackTrace(input)
			require.NoError(t, err)

			output := ToStackTrace(sym)
			assert.Equal(t, input, output)
		})
	}
}
