package gsrf

import (
	"fmt"
	"strconv"
	"strings"
)

// Parse parses a GSRF symbol string according to the specification.
func Parse(input string) (*Symbol, error) {
	if input == "" {
		return nil, fmt.Errorf("invalid GSRF symbol: empty string")
	}
	
	// Extract metadata first
	metadata := Metadata{}
	if idx := strings.LastIndex(input, "{"); idx > 0 && strings.HasSuffix(input, "}") {
		metaStr := input[idx+1 : len(input)-1]
		// Only update input if we're not inside a type parameter list
		bracketCount := 0
		for i := 0; i < idx; i++ {
			if input[i] == '[' {
				bracketCount++
			} else if input[i] == ']' {
				bracketCount--
			}
		}
		if bracketCount == 0 {
			input = input[:idx]
			
			// Initialize custom map if needed
			if strings.Contains(metaStr, ":") && !strings.HasPrefix(metaStr, "via:") && 
			   !strings.HasPrefix(metaStr, "alias:") && !strings.HasPrefix(metaStr, "pos:") {
				metadata.Custom = make(map[string]string)
			}
			
			// Parse metadata
			for _, part := range strings.Split(metaStr, ",") {
				if kv := strings.SplitN(part, ":", 2); len(kv) == 2 {
					key := strings.TrimSpace(kv[0])
					value := strings.TrimSpace(kv[1])
					switch key {
					case "via":
						metadata.Via = value
					case "alias":
						metadata.Alias = value
					case "pos":
						metadata.Position = value
					default:
						if metadata.Custom == nil {
							metadata.Custom = make(map[string]string)
						}
						metadata.Custom[key] = value
					}
				}
			}
		}
	}
	
	// Extract context modifier - after metadata extraction
	context := ""
	if idx := strings.LastIndex(input, "@"); idx > 0 {
		// Make sure @ is not inside brackets
		bracketCount := 0
		for i := 0; i < idx; i++ {
			if input[i] == '[' {
				bracketCount++
			} else if input[i] == ']' {
				bracketCount--
			}
		}
		if bracketCount == 0 {
			// Extract context and remove from input
			context = input[idx+1:]
			if context == "" {
				return nil, fmt.Errorf("invalid GSRF symbol: empty context after @")
			}
			input = input[:idx]
		}
	}

	// Handle methods with receivers first
	var packagePath, symbolPart string
	
	// Check for incomplete receiver syntax first
	if strings.Contains(input, ".(") && !strings.Contains(input, ").") {
		return nil, fmt.Errorf("incomplete method receiver")
	}
	
	if strings.Contains(input, ").") {
		// This is a method - find the last ")." to split correctly
		methodSep := strings.LastIndex(input, ").")
		if methodSep == -1 {
			return nil, fmt.Errorf("invalid method format")
		}
		
		// Find the package separator before the receiver
		lastDotBeforeReceiver := strings.LastIndex(input[:methodSep], ".(")
		if lastDotBeforeReceiver == -1 {
			// Try to find a simple dot before the opening parenthesis
			if openParen := strings.Index(input, "("); openParen > 0 {
				lastDotBeforeReceiver = strings.LastIndex(input[:openParen], ".")
			}
			if lastDotBeforeReceiver == -1 {
				return nil, fmt.Errorf("invalid GSRF symbol: no package separator found")
			}
		}
		
		packagePath = input[:lastDotBeforeReceiver]
		symbolPart = input[lastDotBeforeReceiver+1:]
	} else {
		// Not a method - need to handle generics carefully
		// First check if there are brackets
		if idx := strings.Index(input, "["); idx > 0 {
			// Find the last dot before the bracket
			lastDot := strings.LastIndex(input[:idx], ".")
			if lastDot == -1 {
				return nil, fmt.Errorf("invalid GSRF symbol: no package separator found")
			}
			packagePath = input[:lastDot]
			symbolPart = input[lastDot+1:]
		} else {
			// No brackets, simple case
			lastDot := strings.LastIndex(input, ".")
			if lastDot == -1 || lastDot == 0 || lastDot == len(input)-1 {
				return nil, fmt.Errorf("invalid GSRF symbol: no package separator found")
			}
			packagePath = input[:lastDot]
			symbolPart = input[lastDot+1:]
		}
	}
	
	// Validate package and symbol parts
	if packagePath == "" || symbolPart == "" {
		return nil, fmt.Errorf("invalid GSRF symbol: empty package or symbol part")
	}

	sym := &Symbol{
		PackagePath: packagePath,
		Context:     context,
		Metadata:    metadata,
	}

	// Check if it's init function
	if symbolPart == "init" {
		sym.IsInit = true
		sym.Name = "init"
		return sym, nil
	}

	// Check for anonymous function
	if strings.Contains(symbolPart, "·lit") {
		// Extract parent and index
		litIndex := strings.Index(symbolPart, "·lit")
		sym.Name = symbolPart[:litIndex]
		sym.IsAnonymous = true
		sym.AnonParent = packagePath + "." + sym.Name
		
		// Extract index after ·lit
		litPart := "·lit"
		indexStr := symbolPart[litIndex+len(litPart):] // Skip "·lit" properly (5 bytes)
		if indexStr != "" {
			if index, err := strconv.Atoi(indexStr); err == nil {
				sym.AnonIndex = index
			}
		}
	} else if strings.HasPrefix(symbolPart, "(") {
		// Method with receiver
		recvEnd := strings.Index(symbolPart, ")")
		if recvEnd == -1 || !strings.Contains(symbolPart, ").") {
			return nil, fmt.Errorf("invalid method receiver")
		}
		
		recvStr := symbolPart[1:recvEnd]
		isPtr := strings.HasPrefix(recvStr, "*")
		if isPtr {
			recvStr = recvStr[1:]
		}
		
		// Handle generic receivers
		typeName := recvStr
		var typeArgs []string
		if idx := strings.Index(recvStr, "["); idx > 0 {
			typeName = recvStr[:idx]
			if end := strings.LastIndex(recvStr, "]"); end > idx {
				argsStr := recvStr[idx+1 : end]
				typeArgs = parseTypeArgs(argsStr)
			}
		}
		
		sym.Receiver = &Receiver{
			TypeName:  typeName,
			IsPointer: isPtr,
			TypeArgs:  typeArgs,
		}
		
		// Extract method name
		if recvEnd+2 < len(symbolPart) {
			sym.Name = symbolPart[recvEnd+2:]
		} else {
			return nil, fmt.Errorf("invalid GSRF symbol: receiver without method name")
		}
	} else {
		// Simple function or type
		sym.Name = symbolPart
	}

	// Handle type parameters/arguments in name
	if strings.Contains(sym.Name, "[") {
		if idx := strings.Index(sym.Name, "["); idx > 0 {
			baseName := sym.Name[:idx]
			if end := strings.LastIndex(sym.Name, "]"); end > idx {
				argsStr := sym.Name[idx+1 : end]
				// Parse full type args
				sym.TypeArgs = parseTypeArgs(argsStr)
				sym.Name = baseName
			} else {
				// Unclosed bracket
				return nil, fmt.Errorf("invalid GSRF symbol: unclosed type parameter bracket")
			}
		}
	}

	return sym, nil
}

// MustParse parses a GSRF symbol string and panics on error.
func MustParse(input string) *Symbol {
	sym, err := Parse(input)
	if err != nil {
		panic(err)
	}
	return sym
}

// parseTypeArgs splits type arguments by comma, handling nested brackets
func parseTypeArgs(s string) []string {
	if s == "" {
		return []string{}
	}
	
	var args []string
	var current strings.Builder
	depth := 0
	parenDepth := 0
	
	for _, r := range s {
		switch r {
		case '[':
			depth++
			current.WriteRune(r)
		case ']':
			depth--
			current.WriteRune(r)
		case '(':
			parenDepth++
			current.WriteRune(r)
		case ')':
			parenDepth--
			current.WriteRune(r)
		case ',':
			if depth == 0 && parenDepth == 0 {
				if trimmed := strings.TrimSpace(current.String()); trimmed != "" {
					args = append(args, trimmed)
				}
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}
	
	if current.Len() > 0 {
		if trimmed := strings.TrimSpace(current.String()); trimmed != "" {
			args = append(args, trimmed)
		}
	}
	
	return args
}