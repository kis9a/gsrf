package main

import (
	"fmt"
	"log"

	"github.com/kis9a/gsrf"
	"github.com/kis9a/gsrf/adapters"
)

func main() {
	// Example 1: Parse basic function
	sym1, err := gsrf.Parse("fmt.Println")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Parsed: %+v\n", sym1)
	fmt.Printf("Formatted: %s\n\n", sym1.Format())

	// Example 2: Parse method with pointer receiver
	sym2, err := gsrf.Parse("net/http.(*Server).Serve")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Package: %s\n", sym2.PackagePath)
	fmt.Printf("Method: %s\n", sym2.Name)
	fmt.Printf("Receiver: %s (pointer: %v)\n\n", sym2.Receiver.TypeName, sym2.Receiver.IsPointer)

	// Example 3: Parse symbol with generics
	sym3, err := gsrf.Parse("github.com/user/repo.Map[string,int]")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Function with generics: %s\n", sym3.Name)
	fmt.Printf("Type arguments: %v\n\n", sym3.TypeArgs)

	// Example 4: Parse symbol with context and metadata
	sym4, err := gsrf.Parse("pkg.(*Handler[T]).Process@linux{pos:handler.go:42:10,test:integration}")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Complex symbol:\n")
	fmt.Printf("  Package: %s\n", sym4.PackagePath)
	fmt.Printf("  Method: %s\n", sym4.Name)
	fmt.Printf("  Receiver: %s[%v]\n", sym4.Receiver.TypeName, sym4.Receiver.TypeArgs)
	fmt.Printf("  Context: %s\n", sym4.Context)
	fmt.Printf("  Metadata: %v\n\n", sym4.Metadata)

	// Example 5: Convert from SSA format
	sym5, err := adapters.FromSSA("github.com/user/repo.(*Server).Start@/path/to/file.go:100:5")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("From SSA: %s\n", sym5.Format())
	fmt.Printf("Back to SSA: %s\n\n", adapters.ToSSA(sym5))

	// Example 6: Convert from stack trace
	sym6, err := adapters.FromStackTrace("github.com/user/repo.(*Handler).ServeHTTP")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("From stack trace: %s\n", sym6.Format())
	fmt.Printf("Back to stack trace: %s\n\n", adapters.ToStackTrace(sym6))

	// Example 7: Anonymous functions
	sym7, err := gsrf.Parse("main.main·lit2")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Anonymous function: parent=%s, index=%d\n", sym7.AnonParent, sym7.AnonIndex)

	// Example 8: Init function
	sym8, err := gsrf.Parse("github.com/user/repo.init")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Init function: %v\n", sym8.IsInit)

	// Example 9: Function with nested generics
	sym9, err := gsrf.Parse("pkg.Transform[Map[K,V],List[T]]")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Nested generics: %v\n", sym9.TypeParams)
}
