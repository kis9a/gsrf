// Package gsrf implements the Go Symbol Representation Format specification.
package gsrf

import (
	"fmt"
	"strings"
)

// Symbol represents a parsed GSRF symbol with all features.
type Symbol struct {
	// Core fields
	PackagePath string            // Full package import path
	Name        string            // Function/method/type name
	Receiver    *Receiver         // Method receiver (nil for functions)
	IsInit      bool              // True for init functions
	IsAnonymous bool              // True for anonymous functions
	AnonParent  string            // Parent symbol for anonymous functions
	AnonIndex   int               // Index for anonymous functions (0 = no index)

	// Extended fields (v1.1)
	TypeParams []TypeParam        // Type parameters with constraints
	TypeArgs   []string           // Type arguments (for instantiation)
	Context    string             // Context modifier (@linux, @cgo, etc)
	Metadata   Metadata           // Additional metadata
}

// Receiver represents a method receiver.
type Receiver struct {
	TypeName  string   // Name of the receiver type
	IsPointer bool     // True if pointer receiver
	TypeArgs  []string // Type arguments for generic receivers
}

// TypeParam represents a type parameter with optional constraint.
type TypeParam struct {
	Name       string // Parameter name (e.g., "T")
	Constraint string // Constraint type (empty means "any")
}

// Metadata represents symbol metadata.
type Metadata struct {
	Via      string            // Embedded source (promoted methods)
	Alias    string            // Alias source
	Position string            // Source position (file:line:col)
	Custom   map[string]string // Additional custom metadata
}

// Format returns the formatted GSRF string representation.
func (s *Symbol) Format() string {
	var result strings.Builder
	
	// Package path
	result.WriteString(s.PackagePath)
	result.WriteByte('.')

	// Receiver (for methods)
	if s.Receiver != nil {
		result.WriteByte('(')
		if s.Receiver.IsPointer {
			result.WriteByte('*')
		}
		result.WriteString(s.Receiver.TypeName)
		
		// Generic receiver type args
		if len(s.Receiver.TypeArgs) > 0 {
			result.WriteByte('[')
			result.WriteString(strings.Join(s.Receiver.TypeArgs, ", "))
			result.WriteByte(']')
		}
		
		result.WriteString(").")
	}

	// Function/method name
	if s.IsAnonymous {
		// Anonymous function: use middle dot notation
		result.WriteString(s.Name)
		result.WriteString("·lit")
		if s.AnonIndex > 0 {
			result.WriteString(fmt.Sprintf("%d", s.AnonIndex))
		}
	} else {
		result.WriteString(s.Name)
	}

	// Type parameters or arguments
	if len(s.TypeArgs) > 0 {
		// Type arguments (instantiation) - takes precedence
		result.WriteByte('[')
		result.WriteString(strings.Join(s.TypeArgs, ", "))
		result.WriteByte(']')
	} else if len(s.TypeParams) > 0 {
		// Type parameters (definition)
		result.WriteByte('[')
		for i, tp := range s.TypeParams {
			if i > 0 {
				result.WriteString(", ")
			}
			result.WriteString(tp.Name)
			if tp.Constraint != "" && tp.Constraint != "any" {
				result.WriteByte(' ')
				result.WriteString(tp.Constraint)
			}
		}
		result.WriteByte(']')
	}

	// Context modifier (@linux, @cgo, etc)
	if s.Context != "" {
		result.WriteByte('@')
		result.WriteString(s.Context)
	}

	// Metadata
	if hasMetadata(s.Metadata) {
		var metaParts []string
		
		if s.Metadata.Via != "" {
			metaParts = append(metaParts, "via:"+s.Metadata.Via)
		}
		if s.Metadata.Alias != "" {
			metaParts = append(metaParts, "alias:"+s.Metadata.Alias)
		}
		if s.Metadata.Position != "" {
			metaParts = append(metaParts, "pos:"+s.Metadata.Position)
		}
		for k, v := range s.Metadata.Custom {
			metaParts = append(metaParts, k+":"+v)
		}
		
		if len(metaParts) > 0 {
			result.WriteByte('{')
			result.WriteString(strings.Join(metaParts, ","))
			result.WriteByte('}')
		}
	}

	return result.String()
}

// String implements the Stringer interface.
func (s *Symbol) String() string {
	return s.Format()
}

// hasMetadata checks if the metadata has any values set
func hasMetadata(m Metadata) bool {
	return m.Via != "" || m.Alias != "" || m.Position != "" || len(m.Custom) > 0
}