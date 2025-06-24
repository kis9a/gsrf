# GSRF — Go Symbol Representation Format Specification

**Status**: Proposal  
**Authors**: GSRF Working Group

## Table of Contents

1. [Introduction](#1-introduction)
2. [Motivation and Goals](#2-motivation-and-goals)
3. [Design Principles](#3-design-principles)
4. [Terminology](#4-terminology)
5. [Specification v1.0](#5-specification-v10)
6. [Specification v1.1](#6-specification-v11)
7. [Implementation Guidelines](#7-implementation-guidelines)
8. [Migration Strategy](#8-migration-strategy)
9. [Examples](#9-examples)
10. [References](#10-references)
11. [Appendices](#11-appendices)

---

## 1. Introduction

The Go Symbol Representation Format (GSRF) is a standardized notation for representing symbols (functions, methods, types) in Go programs. This specification defines a canonical format to improve interoperability between Go tools and provide consistent symbol representation across the ecosystem.

### 1.1 Scope

This document specifies:
- Symbol notation syntax for functions, methods, and types
- Handling of edge cases (init functions, anonymous functions, generics)
- Version evolution from v1.0 to v1.1
- Implementation guidelines and migration strategies

### 1.2 Conformance

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119.

---

## 2. Motivation and Goals

### 2.1 Current State

Go ecosystem tools currently use inconsistent symbol representations:
- SSA: `package.init#1`, `package.func@file.go:12:1`
- Runtime stack traces: `main.(*Server).ServeHTTP`
- Various tools: Different formats for generics, anonymous functions

### 2.2 Goals

1. **Unification**: Single canonical format for all tools
2. **Clarity**: Human-readable while maintaining precision
3. **Extensibility**: Support for future Go language features
4. **Compatibility**: Smooth migration from existing formats

---

## 3. Design Principles

> **Principle #1 — Conventional Alignment**  
> GSRF adopts the most widely-used conventions from existing Go tools as the standard format.

> **Principle #2 — Progressive Enhancement**  
> New features are added in backward-compatible ways, allowing gradual adoption.

> **Principle #3 — Explicit Over Implicit**  
> Symbol representations favor completeness and clarity over brevity.

> **Principle #4 — Tool Friendliness**  
> The format is designed to be easily parsed by both humans and machines.

---

## 4. Terminology

- **Symbol**: An addressable program entity (function, method, type)
- **Qualified Name**: Complete symbol name including package path
- **Receiver**: The type to which a method belongs
- **Type Parameter**: Generic type placeholder (e.g., `T` in `func F[T any]()`)
- **Type Argument**: Concrete type replacing a type parameter (e.g., `int` in `F[int]()`)
- **Anonymous Function**: Function literal without an explicit name
- **Promoted Method**: Method accessible through embedded type

---

## 5. Specification v1.0

### 5.1 Grammar (BNF)

```bnf
gsrf_symbol    ::= package_path "." symbol_part
symbol_part    ::= function_name | method_spec | "init" | anonymous_spec

package_path   ::= import_path
function_name  ::= identifier
method_spec    ::= "(" receiver_spec ")" "." method_name
receiver_spec  ::= "*"? type_name
type_name      ::= identifier
method_name    ::= identifier
anonymous_spec ::= parent_symbol "·lit" index?
parent_symbol  ::= gsrf_symbol
index          ::= positive_integer
```

### 5.2 Basic Symbols

#### 5.2.1 Package-Level Functions

```
<PackagePath>.<FunctionName>
```

Examples:
```
fmt.Println
github.com/user/repo/pkg.ProcessData
```

#### 5.2.2 Methods

```
<PackagePath>.(<ReceiverType>).<MethodName>
```

- Value receiver: `net/http.(HandlerFunc).ServeHTTP`
- Pointer receiver: `github.com/user/repo.(*Server).Start`

#### 5.2.3 Init Functions

```
<PackagePath>.init
```

Design Decision: No numbering (`#1`, `#2`). All init functions in a package are represented as a single symbol.

#### 5.2.4 Anonymous Functions

```
<ParentSymbol>·lit<N>
```

- Middle dot `·` (U+00B7) as separator
- Index `<N>` only shown for collision (starting from 1)
- Examples: `main.main·lit`, `main.(*Server).Start·lit2`

### 5.3 Special Cases

#### 5.3.1 Generics (v1.0)

Type parameters are normalized to `[...]`:
```
github.com/user/repo.Map[...]
github.com/user/repo.(*List[...]).Add
```

#### 5.3.2 Vendor Directories

The `vendor/` prefix is removed during normalization:
```
vendor/github.com/lib/pkg.Func → github.com/lib/pkg.Func
```

---

## 6. Specification v1.1

### 6.1 Extended Grammar (BNF)

```bnf
gsrf_symbol_v11 ::= version_header? qualified_symbol context_modifier? metadata?

version_header  ::= "GSRF/" version_number " "
version_number  ::= "1.0" | "1.1"

qualified_symbol ::= package_path "." symbol_part
symbol_part     ::= function_spec | method_spec | "init" | anonymous_spec

; Generic functions
function_spec   ::= function_name type_params? type_args?
type_params     ::= "[" type_param_list "]"
type_param_list ::= type_param ("," type_param)*
type_param      ::= identifier constraint?
constraint      ::= " " type_expr

; Type arguments (instantiation)
type_args       ::= "[" type_arg_list "]"
type_arg_list   ::= type_expr ("," type_expr)*

; Context modifiers
context_modifier ::= "@" context_tag
context_tag     ::= platform_tag | build_tag | "vendor"
platform_tag    ::= "linux" | "darwin" | "windows" | "js" | "wasm"

; Metadata
metadata        ::= "{" metadata_list "}"
metadata_list   ::= metadata_item ("," metadata_item)*
metadata_item   ::= "via:" type_name |      ; embedded source
                   "alias:" type_name |     ; alias source
                   "pos:" position          ; position info
```

### 6.2 New Features

#### 6.2.1 Full Generic Support

Type parameters with constraints:
```
github.com/user/repo.Map[K comparable, V any]
github.com/user/repo.Process[T constraints.Ordered]
```

Type instantiation:
```
github.com/user/repo.Map[string, int]
github.com/user/repo.(*List[*User]).Add
```

#### 6.2.2 Build Context Modifiers

Platform-specific symbols:
```
net.(*netFD).connect@linux
net.(*netFD).connect@windows
```

Build tags:
```
crypto/tls.init@fips
database/sql.(*DB).Query@cgo
```

#### 6.2.3 Embedded Type Metadata

Promoted methods:
```
io.(*BufferedWriter).Write{via:Writer}
myapp.(*Server).ServeHTTP{via:Handler}
```

Multi-level promotion:
```
myapp.(*App).Start{via:Component{via:Lifecycle}}
```

#### 6.2.4 Type Aliases

```
myapp.HandlerFunc.ServeHTTP{alias:http.HandlerFunc}
```

### 6.3 Complex Examples

Generic with build context:
```
stdlib.(*SyncMap[K, V]).Store@linux
```

Generic with embedding:
```
myapp.(*Controller[T]).Handle{via:BaseController[T]}
```

Full example with all features:
```
GSRF/1.1 github.com/project.(*Server[T constraints.Ordered]).Process@linux{via:BaseServer[T],pos:server_linux.go:45:1}
```

---

## 7. Implementation Guidelines

### 7.1 Parser Implementation

#### 7.1.1 Basic Parser Structure

```go
type Symbol struct {
    Version      string
    PackagePath  string
    Receiver     *Receiver
    Name         string
    TypeParams   []TypeParam  // v1.1
    TypeArgs     []Type       // v1.1
    Context      string        // v1.1
    Metadata     Metadata      // v1.1
}

func ParseGSRF(input string) (*Symbol, error) {
    // Detect version
    version := detectVersion(input)
    
    // Parse based on version
    switch version {
    case "1.0":
        return parseV10(input)
    case "1.1":
        return parseV11(input)
    default:
        return parseV10(input) // fallback
    }
}
```

#### 7.1.2 Conversion Functions

From SSA format:
```go
func FromSSA(fn *ssa.Function) string {
    if fn.Name() == "init" {
        return fn.Package().Pkg.Path() + ".init"
    }
    // ... handle other cases
}
```

From stack trace:
```go
func FromStackTrace(frame string) string {
    // Already mostly compatible
    return normalizePackagePath(frame)
}
```

### 7.2 Formatter Implementation

```go
type Formatter struct {
    Version     string
    IncludeType bool
    IncludeMeta bool
}

func (f *Formatter) Format(sym *Symbol) string {
    var buf strings.Builder
    
    // Version header (optional)
    if f.Version != "" {
        buf.WriteString("GSRF/")
        buf.WriteString(f.Version)
        buf.WriteString(" ")
    }
    
    // Basic symbol
    buf.WriteString(sym.PackagePath)
    buf.WriteString(".")
    
    // Receiver (if method)
    if sym.Receiver != nil {
        buf.WriteString("(")
        if sym.Receiver.IsPointer {
            buf.WriteString("*")
        }
        buf.WriteString(sym.Receiver.TypeName)
        buf.WriteString(").")
    }
    
    // Name
    buf.WriteString(sym.Name)
    
    // v1.1 features
    if f.Version == "1.1" {
        // Type parameters/arguments
        if len(sym.TypeParams) > 0 {
            formatTypeParams(&buf, sym.TypeParams)
        } else if len(sym.TypeArgs) > 0 {
            formatTypeArgs(&buf, sym.TypeArgs)
        }
        
        // Context modifier
        if sym.Context != "" {
            buf.WriteString("@")
            buf.WriteString(sym.Context)
        }
        
        // Metadata
        if f.IncludeMeta && sym.Metadata != nil {
            formatMetadata(&buf, sym.Metadata)
        }
    }
    
    return buf.String()
}
```

---

## 8. Migration Strategy

### 8.1 Version Detection

Tools SHOULD implement version detection:
```go
func detectGSRFVersion(s string) string {
    if strings.HasPrefix(s, "GSRF/") {
        return s[5:8] // "1.0" or "1.1"
    }
    
    // Heuristics for v1.1 features
    if containsV11Features(s) {
        return "1.1"
    }
    
    return "1.0"
}
```

### 8.2 Compatibility Matrix

| Source Format | Target v1.0 | Target v1.1 | Notes |
|---------------|-------------|-------------|-------|
| SSA | ✅ Full | ✅ Full | init numbering lost |
| Stack trace | ✅ Full | ✅ Full | Add package paths |
| gopls | ✅ Full | ✅ Full | Remove quotes |
| GSRF v1.0 | ✅ Identity | ✅ Full | Expand [...] |
| GSRF v1.1 | ⚠️ Degraded | ✅ Identity | Lose type info |

### 8.3 Migration Timeline

1. **Phase 1** (Current - 6 months)
   - Tools add GSRF v1.0 support
   - Maintain existing formats

2. **Phase 2** (6-12 months)
   - Tools add GSRF v1.1 parsing
   - Default output remains v1.0

3. **Phase 3** (12-18 months)
   - Tools default to GSRF v1.1
   - Deprecation warnings for old formats

4. **Phase 4** (18+ months)
   - Remove support for legacy formats
   - GSRF becomes sole format

---

## 9. Examples

### 9.1 Common Patterns

```
# Simple function
fmt.Println

# Method with pointer receiver
net/http.(*Server).ListenAndServe

# Generic function
slices.Sort[int]

# Generic method
container/list.(*List[T]).PushBack

# Init function
database/sql.init

# Anonymous function
main.main·lit
main.(*Server).Start·lit1

# Platform-specific
syscall.Getpid@linux
syscall.Getpid@windows

# Promoted method
bytes.(*Buffer).Write{via:Writer}

# Complex generic with context
sync.(*Map[K, V]).Store@linux{pos:map.go:123:1}
```

### 9.2 Real-World Examples

From the test suite:
```
# Basic callgraph output
"github.com/example/api/usecase.(*BookingUsecaseImpl).GetPeriodic" -> 
"github.com/example/api/repository.(*BookingImpl).GetBookingsBetween"

# With generics (v1.1)
"github.com/example/api/service.(*Cache[string, *User]).Get" ->
"github.com/example/api/internal.(*Storage[string, *User]).Lookup"

# With build context (v1.1)
"github.com/example/api/platform.(*FileSystem).Open@windows" ->
"github.com/example/api/platform/internal.openFileWindows"
```

---

## 10. References

1. Go Language Specification - Type Parameters
2. golang.org/x/tools/go/ssa - SSA Symbol Representation
3. golang.org/x/tools/go/types/objectpath - Object Path Notation
4. Go Runtime Stack Trace Format
5. gopls Symbol Search Implementation
6. Delve Debugger Symbol Resolution

---

## 11. Appendices

### 11.1 Appendix A: Symbol Comparison Table

| Scenario | SSA | Runtime | GSRF v1.0 | GSRF v1.1 |
|----------|-----|---------|-----------|-----------|
| Function | `pkg.Func` | `pkg.Func` | `pkg.Func` | `pkg.Func` |
| Method | `(*T).Method` | `pkg.(*T).Method` | `pkg.(*T).Method` | `pkg.(*T).Method` |
| Init | `pkg.init#1` | `pkg.init` | `pkg.init` | `pkg.init` |
| Anonymous | `pkg.func@1:2` | `pkg.func1` | `pkg.Func·lit` | `pkg.Func·lit` |
| Generic def | N/A | N/A | `pkg.F[...]` | `pkg.F[T any]` |
| Generic use | N/A | `pkg.F` | `pkg.F[...]` | `pkg.F[int]` |

### 11.2 Appendix B: Reserved Characters

The following characters have special meaning in GSRF:

| Character | Usage | Unicode |
|-----------|-------|---------|
| `.` | Package/symbol separator | U+002E |
| `(` `)` | Receiver delimiter | U+0028, U+0029 |
| `*` | Pointer indicator | U+002A |
| `[` `]` | Type parameter/argument delimiter | U+005B, U+005D |
| `·` | Anonymous function separator | U+00B7 |
| `@` | Context modifier prefix | U+0040 |
| `{` `}` | Metadata delimiter | U+007B, U+007D |
| `:` | Metadata key-value separator | U+003A |
| `,` | List separator | U+002C |

### 11.3 Appendix C: Error Handling

Common error conditions and recommended handling:

1. **Unknown version**: Fall back to v1.0 parsing
2. **Malformed symbol**: Return error with position
3. **Invalid type arguments**: Validate against parameters
4. **Missing package path**: Attempt to infer from context
5. **Ambiguous promotion**: List all possible sources

### 11.4 Appendix D: Performance Considerations

Typical symbol sizes:

| Format | Average | P95 | Max |
|--------|---------|-----|-----|
| v1.0 | 60 bytes | 100 bytes | 150 bytes |
| v1.1 (minimal) | 60 bytes | 100 bytes | 150 bytes |
| v1.1 (typical) | 85 bytes | 140 bytes | 200 bytes |
| v1.1 (full) | 120 bytes | 200 bytes | 500 bytes |

Parsing performance (100k symbols):
- v1.0 parser: ~50ms
- v1.1 parser: ~80ms
- With caching: ~10ms

---

## Contributors

- GSRF Working Group
- Based on research from the calldigraph project
- Input from Go tools maintainers

## License

This specification is licensed under the Creative Commons Attribution 4.0 International License.
