// Package adapters provides converters between GSRF and other formats.
package adapters

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kis9a/gsrf"
)

var (
	// SSA patterns
	ssaInitPattern     = regexp.MustCompile(`^(.+)\.init#(\d+)$`)
	ssaFuncPattern     = regexp.MustCompile(`^(.+)\.([^.]+)$`)
	ssaMethodPattern   = regexp.MustCompile(`^(.+)\.\((\*?)([^)]+)\)\.([^.]+)$`)
	ssaAnonPattern     = regexp.MustCompile(`^(.+)\.([^.]+)\$\d+$`)
	ssaLocationPattern = regexp.MustCompile(`^(.+)@([^:]+):(\d+):(\d+)$`)
)

// FromSSA converts SSA format to GSRF.
func FromSSA(ssa string) (*gsrf.Symbol, error) {
	// Remove location info if present
	location := ""
	if matches := ssaLocationPattern.FindStringSubmatch(ssa); matches != nil {
		ssa = matches[1]
		location = fmt.Sprintf("%s:%s:%s", matches[2], matches[3], matches[4])
	}

	// Try init pattern
	if matches := ssaInitPattern.FindStringSubmatch(ssa); matches != nil {
		sym := &gsrf.Symbol{
			PackagePath: matches[1],
			Name:        "init",
			IsInit:      true,
		}
		if location != "" {
			sym.Metadata = gsrf.Metadata{
				Position: location,
			}
		}
		return sym, nil
	}

	// Try method pattern
	if matches := ssaMethodPattern.FindStringSubmatch(ssa); matches != nil {
		sym := &gsrf.Symbol{
			PackagePath: matches[1],
			Name:        matches[4],
			Receiver: &gsrf.Receiver{
				TypeName:  matches[3],
				IsPointer: matches[2] == "*",
			},
		}
		if location != "" {
			sym.Metadata = gsrf.Metadata{
				Position: location,
			}
		}
		return sym, nil
	}

	// Try anonymous function pattern
	if matches := ssaAnonPattern.FindStringSubmatch(ssa); matches != nil {
		parts := strings.Split(matches[0], "$")
		if len(parts) == 2 {
			index, _ := strconv.Atoi(parts[1])
			baseParts := strings.LastIndex(parts[0], ".")
			if baseParts >= 0 {
				sym := &gsrf.Symbol{
					PackagePath: parts[0][:baseParts],
					Name:        parts[0][baseParts+1:], // Extract parent function name
					IsAnonymous: true,
					AnonParent:  parts[0],
					AnonIndex:   index,
				}
				if location != "" {
					sym.Metadata = gsrf.Metadata{
						Position: location,
					}
				}
				return sym, nil
			}
		}
	}

	// Try function pattern
	if matches := ssaFuncPattern.FindStringSubmatch(ssa); matches != nil {
		sym := &gsrf.Symbol{
			PackagePath: matches[1],
			Name:        matches[2],
		}
		if location != "" {
			sym.Metadata = gsrf.Metadata{
				Position: location,
			}
		}
		return sym, nil
	}

	return nil, fmt.Errorf("invalid SSA format: %s", ssa)
}

// ToSSA converts GSRF to SSA format.
func ToSSA(sym *gsrf.Symbol) string {
	var result strings.Builder

	result.WriteString(sym.PackagePath)
	result.WriteByte('.')

	if sym.IsInit {
		result.WriteString("init#1")
	} else if sym.Receiver != nil {
		result.WriteByte('(')
		if sym.Receiver.IsPointer {
			result.WriteByte('*')
		}
		result.WriteString(sym.Receiver.TypeName)
		result.WriteByte(')')
		result.WriteByte('.')
		result.WriteString(sym.Name)
	} else if sym.IsAnonymous {
		// Convert to SSA anonymous format
		// Use Name field which contains the parent function name
		result.WriteString(sym.Name)
		result.WriteByte('$')
		if sym.AnonIndex > 0 {
			result.WriteString(strconv.Itoa(sym.AnonIndex))
		} else {
			result.WriteByte('1')
		}
	} else {
		result.WriteString(sym.Name)
	}

	// Add location metadata if available
	if sym.Metadata.Position != "" {
		result.WriteByte('@')
		result.WriteString(sym.Metadata.Position)
	}

	return result.String()
}