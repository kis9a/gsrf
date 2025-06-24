# GSRF - Go Symbol Representation Format

A Go package implementing the GSRF (Go Symbol Representation Format) specification.

## Installation

```bash
go get github.com/kis9a/gsrf
```

## Usage

### As a Library

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/kis9a/gsrf"
)

func main() {
    // Parse a GSRF symbol
    sym, err := gsrf.Parse("fmt.Println")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Package: %s\n", sym.PackagePath)
    fmt.Printf("Function: %s\n", sym.Name)
    
    // Parse symbol with generics
    sym2, err := gsrf.Parse("pkg.Map[T,U]")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Type Arguments: %v\n", sym2.TypeArgs)
    
    // Format a symbol
    formatted := sym.Format()
    fmt.Println(formatted) // "fmt.Println"
}
```

### CLI Tool

```bash
# Install the CLI
go install github.com/kis9a/gsrf/cmd/gsrf@latest

# Parse a symbol
gsrf parse "net/http.(*Server).Serve"

# Convert between formats
gsrf convert "pkg.Function"

# Format from other formats
gsrf format --from ssa "pkg.init#1"
gsrf format --from stacktrace "main.(*Server).Start"

# Parse advanced features
gsrf parse "pkg.Map[T,U]#linux#amd64@{src:file.go:10:1}"

# JSON output
gsrf parse --json "fmt.Println"
```

## Features

### Core Symbol Types
- Package-level functions: `fmt.Println`
- Methods: `net/http.(*Server).Serve`
- Init functions: `pkg.init`
- Anonymous functions: `main.main·lit`, `main.main·lit2`

### Extended Features
- Generics: `pkg.Map[T,U]`, `pkg.(*List[T]).Add`
- Build contexts: `pkg.Function#linux#amd64`
- Metadata: `pkg.Function@{src:file.go:12:1}`

### Format Adapters
- SSA format conversion
- Stack trace format conversion

## API Reference

### Parsing

```go
// Parse a GSRF symbol
sym, err := gsrf.Parse("fmt.Println")

// Must parse (panics on error)
sym := gsrf.MustParse("fmt.Println")
```

### Symbol Type

```go
type Symbol struct {
    Package      string
    Function     string
    Receiver     *Receiver
    IsInit       bool
    IsAnonymous  bool
    AnonParent   string
    AnonIndex    int
    TypeParams   []string
    TypeArgs     []string
    BuildContext *BuildContext
    Metadata     map[string]string
}

type Receiver struct {
    Type      string
    IsPointer bool
    TypeArgs  []string
}

type BuildContext struct {
    OS   string
    Arch string
    Tags []string
}
```

### Formatting

```go
// Format symbol to GSRF string
formatted := sym.Format()

// String() is equivalent to Format()
formatted := sym.String()
```

### Adapters

```go
import "github.com/kis9a/gsrf/adapters"

// From SSA format
sym, err := adapters.FromSSA("pkg.init#1")

// To SSA format
ssa := adapters.ToSSA(sym)

// From stack trace
sym, err := adapters.FromStackTrace("main.(*Server).Start")

// To stack trace
trace := adapters.ToStackTrace(sym)
```

## Examples

See the [examples](examples/) directory for more usage examples.

## Testing

```bash
go test ./... -cover
```

## License

MIT License - see LICENSE file for details.