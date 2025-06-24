package gsrf

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		want       *Symbol
		wantErr    bool
	}{
		// Basic tests
		{
			name:  "simple function",
			input: "github.com/user/repo.Function",
			want: &Symbol{
				PackagePath: "github.com/user/repo",
				Name:        "Function",
				Metadata:    Metadata{},
			},
		},
		{
			name:  "method with pointer receiver",
			input: "github.com/user/repo.(*Type).Method",
			want: &Symbol{
				PackagePath: "github.com/user/repo",
				Name:        "Method",
				Receiver: &Receiver{
					TypeName:  "Type",
					IsPointer: true,
				},
				Metadata: Metadata{},
			},
		},
		{
			name:  "method with value receiver",
			input: "net/http.(HandlerFunc).ServeHTTP",
			want: &Symbol{
				PackagePath: "net/http",
				Name:        "ServeHTTP",
				Receiver: &Receiver{
					TypeName:  "HandlerFunc",
					IsPointer: false,
				},
				Metadata: Metadata{},
			},
		},
		{
			name:  "init function",
			input: "database/sql.init",
			want: &Symbol{
				PackagePath: "database/sql",
				Name:        "init",
				IsInit:      true,
				Metadata:    Metadata{},
			},
		},
		{
			name:  "anonymous function",
			input: "main.main·lit",
			want: &Symbol{
				PackagePath: "main",
				Name:        "main",
				IsAnonymous: true,
				AnonParent:  "main.main",
				AnonIndex:   0,
				Metadata:    Metadata{},
			},
		},
		{
			name:  "numbered anonymous function",
			input: "main.(*Server).Start·lit2",
			want: &Symbol{
				PackagePath: "main",
				Name:        "(*Server).Start",
				IsAnonymous: true,
				AnonParent:  "main.(*Server).Start",
				AnonIndex:   2,
				Metadata:    Metadata{},
			},
		},
		{
			name:  "generic function with type args",
			input: "github.com/user/repo.Map[string, int]",
			want: &Symbol{
				PackagePath: "github.com/user/repo",
				Name:        "Map",
				TypeArgs:    []string{"string", "int"},
				Metadata:    Metadata{},
			},
		},
		{
			name:  "generic receiver",
			input: "github.com/user/repo.(*List[T]).Add",
			want: &Symbol{
				PackagePath: "github.com/user/repo",
				Name:        "Add",
				Receiver: &Receiver{
					TypeName:  "List",
					IsPointer: true,
					TypeArgs:  []string{"T"},
				},
				Metadata: Metadata{},
			},
		},
		{
			name:  "with build context",
			input: "net.(*netFD).connect@linux",
			want: &Symbol{
				PackagePath: "net",
				Name:        "connect",
				Receiver: &Receiver{
					TypeName:  "netFD",
					IsPointer: true,
				},
				Context:  "linux",
				Metadata: Metadata{},
			},
		},
		{
			name:  "with metadata",
			input: "io.(*Buffer).Write{via:Writer}",
			want: &Symbol{
				PackagePath: "io",
				Name:        "Write",
				Receiver: &Receiver{
					TypeName:  "Buffer",
					IsPointer: true,
				},
				Metadata: Metadata{
					Via: "Writer",
				},
			},
		},
		{
			name:  "complex example",
			input: "pkg.(*Controller[T]).Handle@linux{via:Base[T],pos:file.go:10:1}",
			want: &Symbol{
				PackagePath: "pkg",
				Name:        "Handle",
				Receiver: &Receiver{
					TypeName:  "Controller",
					IsPointer: true,
					TypeArgs:  []string{"T"},
				},
				Context: "linux",
				Metadata: Metadata{
					Via:      "Base[T]",
					Position: "file.go:10:1",
				},
			},
		},

		// Error cases
		{
			name:    "missing package separator",
			input:   "invalidformat",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "incomplete receiver",
			input:   "pkg.(*Type",
			wantErr: true,
		},
		{
			name:    "empty context modifier",
			input:   "pkg.Function@",
			wantErr: true,
		},
		{
			name:    "unclosed type bracket",
			input:   "pkg.Function[T",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				// Print detailed comparison for debugging
				if got.PackagePath != tt.want.PackagePath {
					t.Errorf("PackagePath = %q, want %q", got.PackagePath, tt.want.PackagePath)
				}
				if got.Name != tt.want.Name {
					t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
				}
				if got.IsInit != tt.want.IsInit {
					t.Errorf("IsInit = %v, want %v", got.IsInit, tt.want.IsInit)
				}
				if got.IsAnonymous != tt.want.IsAnonymous {
					t.Errorf("IsAnonymous = %v, want %v", got.IsAnonymous, tt.want.IsAnonymous)
				}
				if got.AnonParent != tt.want.AnonParent {
					t.Errorf("AnonParent = %q, want %q", got.AnonParent, tt.want.AnonParent)
				}
				if got.AnonIndex != tt.want.AnonIndex {
					t.Errorf("AnonIndex = %d, want %d", got.AnonIndex, tt.want.AnonIndex)
				}
				if got.Context != tt.want.Context {
					t.Errorf("Context = %q, want %q", got.Context, tt.want.Context)
				}
				if !reflect.DeepEqual(got.Receiver, tt.want.Receiver) && (got.Receiver != nil || tt.want.Receiver != nil) {
					t.Errorf("Receiver = %+v, want %+v", got.Receiver, tt.want.Receiver)
				}
				if !reflect.DeepEqual(got.TypeArgs, tt.want.TypeArgs) {
					t.Errorf("TypeArgs = %v, want %v", got.TypeArgs, tt.want.TypeArgs)
				}
				if got.Metadata.Via != tt.want.Metadata.Via {
					t.Errorf("Metadata.Via = %q, want %q", got.Metadata.Via, tt.want.Metadata.Via)
				}
				if got.Metadata.Position != tt.want.Metadata.Position {
					t.Errorf("Metadata.Position = %q, want %q", got.Metadata.Position, tt.want.Metadata.Position)
				}
				t.Errorf("Parse() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		sym := MustParse("fmt.Println")
		if sym.PackagePath != "fmt" || sym.Name != "Println" {
			t.Errorf("MustParse() = %+v", sym)
		}
	})

	t.Run("invalid input panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustParse() did not panic")
			}
		}()
		MustParse("invalid")
	})
}

func TestParseTypeArgs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "simple types",
			input: "string, int",
			want:  []string{"string", "int"},
		},
		{
			name:  "nested generics",
			input: "Map[string, int], []byte",
			want:  []string{"Map[string, int]", "[]byte"},
		},
		{
			name:  "complex nested",
			input: "chan Map[K, V], *Slice[T]",
			want:  []string{"chan Map[K, V]", "*Slice[T]"},
		},
		{
			name:  "single type",
			input: "T",
			want:  []string{"T"},
		},
		{
			name:  "with constraints",
			input: "T comparable, U any",
			want:  []string{"T comparable", "U any"},
		},
		{
			name:  "empty string",
			input: "",
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTypeArgs(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTypeArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}