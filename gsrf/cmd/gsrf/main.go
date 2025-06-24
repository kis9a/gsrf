package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kis9a/gsrf"
	"github.com/kis9a/gsrf/adapters"
	"github.com/spf13/cobra"
)

var (
	outputJSON  bool
	inputFormat string
)

var rootCmd = &cobra.Command{
	Use:   "gsrf",
	Short: "Go Symbol Representation Format tool",
	Long:  `A CLI tool for parsing and formatting Go symbols according to the GSRF specification.`,
}

var parseCmd = &cobra.Command{
	Use:   "parse [symbol]",
	Short: "Parse a GSRF symbol",
	Long:  `Parse a GSRF symbol and output its components.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]

		sym, err := gsrf.Parse(input)
		if err != nil {
			return fmt.Errorf("parse error: %w", err)
		}

		if outputJSON {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(sym)
		}

		// Human-readable output
		fmt.Printf("Package: %s\n", sym.PackagePath)
		if sym.Name != "" {
			fmt.Printf("Function: %s\n", sym.Name)
		}
		if sym.Receiver != nil {
			fmt.Printf("Receiver: ")
			if sym.Receiver.IsPointer {
				fmt.Print("*")
			}
			fmt.Print(sym.Receiver.TypeName)
			if len(sym.Receiver.TypeArgs) > 0 {
				fmt.Printf("[%v]", sym.Receiver.TypeArgs)
			}
			fmt.Println()
		}
		if sym.IsInit {
			fmt.Println("Type: init function")
		}
		if sym.IsAnonymous {
			fmt.Printf("Type: anonymous function (parent: %s, index: %d)\n", sym.AnonParent, sym.AnonIndex)
		}
		if len(sym.TypeParams) > 0 {
			fmt.Printf("Type Parameters: %v\n", sym.TypeParams)
		}
		if len(sym.TypeArgs) > 0 {
			fmt.Printf("Type Arguments: %v\n", sym.TypeArgs)
		}
		if sym.Context != "" {
			fmt.Printf("Context: %s\n", sym.Context)
		}
		// Display metadata if present
		if sym.Metadata.Via != "" || sym.Metadata.Alias != "" || sym.Metadata.Position != "" || len(sym.Metadata.Custom) > 0 {
			fmt.Println("Metadata:")
			if sym.Metadata.Via != "" {
				fmt.Printf("  Via: %s\n", sym.Metadata.Via)
			}
			if sym.Metadata.Alias != "" {
				fmt.Printf("  Alias: %s\n", sym.Metadata.Alias)
			}
			if sym.Metadata.Position != "" {
				fmt.Printf("  Position: %s\n", sym.Metadata.Position)
			}
			if len(sym.Metadata.Custom) > 0 {
				fmt.Printf("  Custom: %v\n", sym.Metadata.Custom)
			}
		}

		return nil
	},
}

var formatCmd = &cobra.Command{
	Use:   "format [symbol]",
	Short: "Format a symbol to GSRF notation",
	Long:  `Format a symbol from various formats to GSRF notation.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]

		var sym *gsrf.Symbol
		var err error

		switch inputFormat {
		case "gsrf":
			sym, err = gsrf.Parse(input)
		case "ssa":
			sym, err = adapters.FromSSA(input)
		case "stacktrace", "stack":
			sym, err = adapters.FromStackTrace(input)
		default:
			return fmt.Errorf("unknown input format: %s", inputFormat)
		}

		if err != nil {
			return fmt.Errorf("conversion error: %w", err)
		}

		if outputJSON {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(map[string]string{
				"gsrf": sym.Format(),
			})
		}

		fmt.Println(sym.Format())
		return nil
	},
}

var convertCmd = &cobra.Command{
	Use:   "convert [symbol]",
	Short: "Convert between different symbol formats",
	Long:  `Convert a GSRF symbol to other formats (SSA, stack trace).`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input := args[0]

		sym, err := gsrf.Parse(input)
		if err != nil {
			return fmt.Errorf("parse error: %w", err)
		}

		result := map[string]string{
			"gsrf":       sym.Format(),
			"ssa":        adapters.ToSSA(sym),
			"stacktrace": adapters.ToStackTrace(sym),
		}

		if outputJSON {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(result)
		}

		fmt.Printf("GSRF:       %s\n", result["gsrf"])
		fmt.Printf("SSA:        %s\n", result["ssa"])
		fmt.Printf("Stack Trace: %s\n", result["stacktrace"])

		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gsrf version 1.0.0")
		fmt.Println("GSRF Specification - Latest")
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&outputJSON, "json", false, "Output in JSON format")

	formatCmd.Flags().StringVar(&inputFormat, "from", "gsrf", "Input format (gsrf, ssa, stacktrace)")

	rootCmd.AddCommand(parseCmd)
	rootCmd.AddCommand(formatCmd)
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
