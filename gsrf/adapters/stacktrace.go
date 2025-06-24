package adapters

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kis9a/gsrf"
)

var (
	// Stack trace patterns
	stackMethodPattern = regexp.MustCompile(`^(.+)\.\(\*([^)]+)\)\.(.+)$`)
	stackFuncPattern   = regexp.MustCompile(`^([^[\s]+)\.([^.[]+)$`)
	stackInitPattern   = regexp.MustCompile(`^(.+)\.init\.func\d+$`)
	stackAnonPattern   = regexp.MustCompile(`^(.+)\.func\d+`)
)

// FromStackTrace converts Go runtime stack trace format to GSRF.
func FromStackTrace(trace string) (*gsrf.Symbol, error) {
	// Remove any file:line info (but only if it looks like a file path)
	// Stack traces have format: "pkg.Function /path/to/file.go:123"
	// We need to be careful not to trim spaces inside generics like "Map[K, V]"
	if idx := strings.Index(trace, " /"); idx > 0 {
		trace = trace[:idx]
	} else if idx := strings.LastIndex(trace, " "); idx > 0 {
		// Check if what follows the space looks like a file path
		afterSpace := trace[idx+1:]
		if strings.Contains(afterSpace, ".go:") || strings.Contains(afterSpace, "/") {
			trace = trace[:idx]
		}
	}

	// Check for init functions
	if stackInitPattern.MatchString(trace) {
		parts := strings.Split(trace, ".")
		if len(parts) >= 2 {
			pkg := strings.Join(parts[:len(parts)-2], ".")
			return &gsrf.Symbol{
				PackagePath: pkg,
				Name:        "init",
				IsInit:      true,
				Metadata:    gsrf.Metadata{},
			}, nil
		}
	}

	// Check for anonymous functions
	if matches := stackAnonPattern.FindStringSubmatch(trace); matches != nil {
		// Extract base function
		base := matches[1]
		if idx := strings.LastIndex(base, "."); idx > 0 {
			pkg := base[:idx]
			funcName := base[idx+1:]
			return &gsrf.Symbol{
				PackagePath: pkg,
				Name:        funcName,
				IsAnonymous: true,
				AnonParent:  base,
				Metadata:    gsrf.Metadata{},
			}, nil
		}
	}

	// Check for methods with more robust handling
	if strings.Contains(trace, ".(*") {
		// Find the method pattern more carefully
		startParen := strings.Index(trace, ".(*")
		if startParen > 0 {
			pkg := trace[:startParen]

			// Find the matching closing paren
			remainder := trace[startParen+3:] // Skip ".(*"
			parenDepth := 1
			closeParen := -1
			inBrackets := false

			for i, ch := range remainder {
				if ch == '[' {
					inBrackets = true
				} else if ch == ']' {
					inBrackets = false
				} else if ch == '(' && !inBrackets {
					parenDepth++
				} else if ch == ')' && !inBrackets {
					parenDepth--
					if parenDepth == 0 {
						closeParen = i
						break
					}
				}
			}

			if closeParen > 0 {
				receiverContent := remainder[:closeParen]

				// Parse receiver type and generics
				receiver := receiverContent
				var typeArgs []string

				if idx := strings.Index(receiverContent, "["); idx > 0 {
					receiver = receiverContent[:idx]
					endBracket := strings.LastIndex(receiverContent, "]")
					if endBracket > idx {
						typeArgsStr := receiverContent[idx+1 : endBracket]
						typeArgs = parseTypeParams(typeArgsStr)
					}
				}

				// Method name follows
				afterReceiver := remainder[closeParen+1:]
				if strings.HasPrefix(afterReceiver, ".") {
					method := afterReceiver[1:]

					sym := &gsrf.Symbol{
						PackagePath: pkg,
						Name:        method,
						Receiver: &gsrf.Receiver{
							TypeName:  receiver,
							IsPointer: true,
							TypeArgs:  typeArgs,
						},
						Metadata: gsrf.Metadata{},
					}

					return sym, nil
				}
			}
		}
	}

	// Try function pattern (including generics)
	// First check if there are generics
	if bracketStart := strings.Index(trace, "["); bracketStart > 0 {
		bracketEnd := strings.LastIndex(trace, "]")
		if bracketEnd > bracketStart {
			// Find the last dot before the bracket
			beforeBracket := trace[:bracketStart]
			lastDot := strings.LastIndex(beforeBracket, ".")

			if lastDot > 0 {
				pkg := trace[:lastDot]
				funcName := beforeBracket[lastDot+1:]

				// Extract type params
				typeParamsStr := trace[bracketStart+1 : bracketEnd]
				typeParams := parseTypeParams(typeParamsStr)

				return &gsrf.Symbol{
					PackagePath: pkg,
					Name:        funcName,
					TypeArgs:    typeParams,
					Metadata:    gsrf.Metadata{},
				}, nil
			}
		}
	}

	// No generics - simple function
	// But first check if this looks like a package.function pattern
	lastDot := strings.LastIndex(trace, ".")
	if lastDot > 0 {
		pkg := trace[:lastDot]
		funcName := trace[lastDot+1:]

		// Make sure this is a valid function name (not empty and doesn't contain special chars)
		if funcName != "" && !strings.ContainsAny(funcName, "[]()* \t\n") {
			return &gsrf.Symbol{
				PackagePath: pkg,
				Name:        funcName,
				Metadata:    gsrf.Metadata{},
			}, nil
		}
	}

	return nil, fmt.Errorf("invalid stack trace format: %s", trace)
}

func parseTypeParams(s string) []string {
	// Handle nested generics by tracking bracket depth
	var result []string
	var current strings.Builder
	depth := 0

	for _, ch := range s {
		switch ch {
		case '[':
			depth++
			current.WriteRune(ch)
		case ']':
			depth--
			current.WriteRune(ch)
		case ',':
			if depth == 0 {
				result = append(result, strings.TrimSpace(current.String()))
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		case ' ':
			// Keep spaces in nested types like "Map[K, V]"
			if depth > 0 || current.Len() > 0 {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		result = append(result, strings.TrimSpace(current.String()))
	}

	return result
}

// ToStackTrace converts GSRF to Go runtime stack trace format.
func ToStackTrace(sym *gsrf.Symbol) string {
	var result strings.Builder

	if sym.IsAnonymous {
		// For anonymous functions, AnonParent already contains package.function
		if sym.AnonParent != "" {
			result.WriteString(sym.AnonParent)
			result.WriteString(".func")
			if sym.AnonIndex > 0 {
				result.WriteString(fmt.Sprintf("%d", sym.AnonIndex))
			} else {
				result.WriteByte('1')
			}
		}
		return result.String()
	}

	result.WriteString(sym.PackagePath)
	result.WriteByte('.')

	if sym.IsInit {
		result.WriteString("init.func1")
	} else if sym.Receiver != nil {
		// Stack traces always use pointer notation
		result.WriteString("(*")
		result.WriteString(sym.Receiver.TypeName)
		if len(sym.Receiver.TypeArgs) > 0 {
			result.WriteByte('[')
			result.WriteString(strings.Join(sym.Receiver.TypeArgs, ", "))
			result.WriteByte(']')
		}
		result.WriteString(").")
		result.WriteString(sym.Name)
	} else {
		result.WriteString(sym.Name)
		if len(sym.TypeArgs) > 0 {
			result.WriteByte('[')
			result.WriteString(strings.Join(sym.TypeArgs, ", "))
			result.WriteByte(']')
		}
	}

	return result.String()
}
