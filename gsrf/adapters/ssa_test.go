package adapters

import (
	"testing"

	"github.com/kis9a/gsrf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromSSA(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *gsrf.Symbol
		wantErr  bool
	}{
		{
			name:  "simple function",
			input: "fmt.Println",
			expected: &gsrf.Symbol{
				PackagePath: "fmt",
				Name:        "Println",
				Metadata:    gsrf.Metadata{},
			},
		},
		{
			name:  "init function",
			input: "pkg.init#1",
			expected: &gsrf.Symbol{
				PackagePath: "pkg",
				Name:        "init",
				IsInit:      true,
				Metadata:    gsrf.Metadata{},
			},
		},
		{
			name:  "init function #2",
			input: "github.com/user/repo.init#2",
			expected: &gsrf.Symbol{
				PackagePath: "github.com/user/repo",
				Name:        "init",
				IsInit:      true,
				Metadata:    gsrf.Metadata{},
			},
		},
		{
			name:  "value receiver method",
			input: "net/http.(HandlerFunc).ServeHTTP",
			expected: &gsrf.Symbol{
				PackagePath: "net/http",
				Name:        "ServeHTTP",
				Receiver: &gsrf.Receiver{
					TypeName:  "HandlerFunc",
					IsPointer: false,
				},
				Metadata: gsrf.Metadata{},
			},
		},
		{
			name:  "pointer receiver method",
			input: "pkg.(*Server).Start",
			expected: &gsrf.Symbol{
				PackagePath: "pkg",
				Name:        "Start",
				Receiver: &gsrf.Receiver{
					TypeName:  "Server",
					IsPointer: true,
				},
				Metadata: gsrf.Metadata{},
			},
		},
		{
			name:  "anonymous function",
			input: "main.main$1",
			expected: &gsrf.Symbol{
				PackagePath: "main",
				Name:        "main",
				IsAnonymous: true,
				AnonParent:  "main.main",
				AnonIndex:   1,
				Metadata:    gsrf.Metadata{},
			},
		},
		{
			name:  "function with location",
			input: "pkg.Function@file.go:12:1",
			expected: &gsrf.Symbol{
				PackagePath: "pkg",
				Name:        "Function",
				Metadata: gsrf.Metadata{
					Position: "file.go:12:1",
				},
			},
		},
		{
			name:  "method with location",
			input: "pkg.(*Type).Method@/path/to/file.go:100:5",
			expected: &gsrf.Symbol{
				PackagePath: "pkg",
				Name:        "Method",
				Receiver: &gsrf.Receiver{
					TypeName:  "Type",
					IsPointer: true,
				},
				Metadata: gsrf.Metadata{
					Position: "/path/to/file.go:100:5",
				},
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
			got, err := FromSSA(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestToSSA(t *testing.T) {
	tests := []struct {
		name     string
		symbol   *gsrf.Symbol
		expected string
	}{
		{
			name: "simple function",
			symbol: &gsrf.Symbol{
				PackagePath: "fmt",
				Name:        "Println",
			},
			expected: "fmt.Println",
		},
		{
			name: "init function",
			symbol: &gsrf.Symbol{
				PackagePath: "pkg",
				Name:        "init",
				IsInit:      true,
			},
			expected: "pkg.init#1",
		},
		{
			name: "value receiver method",
			symbol: &gsrf.Symbol{
				PackagePath: "net/http",
				Name:        "ServeHTTP",
				Receiver: &gsrf.Receiver{
					TypeName:  "HandlerFunc",
					IsPointer: false,
				},
			},
			expected: "net/http.(HandlerFunc).ServeHTTP",
		},
		{
			name: "pointer receiver method",
			symbol: &gsrf.Symbol{
				PackagePath: "pkg",
				Name:        "Start",
				Receiver: &gsrf.Receiver{
					TypeName:  "Server",
					IsPointer: true,
				},
			},
			expected: "pkg.(*Server).Start",
		},
		{
			name: "anonymous function",
			symbol: &gsrf.Symbol{
				PackagePath: "main",
				Name:        "main",
				IsAnonymous: true,
				AnonIndex:   2,
			},
			expected: "main.main$2",
		},
		{
			name: "anonymous without index",
			symbol: &gsrf.Symbol{
				PackagePath: "main",
				Name:        "main",
				IsAnonymous: true,
			},
			expected: "main.main$1",
		},
		{
			name: "function with metadata",
			symbol: &gsrf.Symbol{
				PackagePath: "pkg",
				Name:        "Function",
				Metadata: gsrf.Metadata{
					Position: "file.go:12:1",
				},
			},
			expected: "pkg.Function@file.go:12:1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToSSA(tt.symbol)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestSSARoundTrip(t *testing.T) {
	// Test that we can convert from SSA to GSRF and back
	inputs := []string{
		"fmt.Println",
		"pkg.init#1",
		"net/http.(*Server).Serve",
		"main.main$1",
		"pkg.Function@file.go:10:5",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			sym, err := FromSSA(input)
			require.NoError(t, err)

			output := ToSSA(sym)
			assert.Equal(t, input, output)
		})
	}
}
