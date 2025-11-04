package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	dsl "dsl-ob-poc/internal/hf-investor/dsl"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "dsl-validate":
		dslValidate(os.Args[2:])
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Print(`Usage:
  go run ./cmd/hf-cli dsl-validate [flags] < file.json

Commands:
  dsl-validate    Validate a Hedge Fund Investor DSL runbook from stdin

Flags (dsl-validate):
  -pretty         Print normalized JSON on success
`)
}

func dslValidate(args []string) {
	fs := flag.NewFlagSet("dsl-validate", flag.ContinueOnError)
	pretty := fs.Bool("pretty", false, "Print normalized JSON on success")
	fs.SetOutput(io.Discard) // keep stderr clean unless we fail parsing flags
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, "invalid flags:", err)
		os.Exit(2)
	}

	// Read from stdin
	stat, _ := os.Stdin.Stat()
	if stat.Mode()&os.ModeCharDevice != 0 {
		fmt.Fprintln(os.Stderr, "no stdin detected. Example: cat runbook.json | go run ./cmd/hf-cli dsl-validate")
		os.Exit(2)
	}
	raw, err := io.ReadAll(bufio.NewReader(os.Stdin))
	if err != nil {
		fmt.Fprintln(os.Stderr, "read error:", err)
		os.Exit(1)
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		fmt.Fprintln(os.Stderr, "empty input on stdin")
		os.Exit(2)
	}

	var rb dsl.Runbook
	if unmarshalErr := json.Unmarshal([]byte(trimmed), &rb); unmarshalErr != nil {
		fmt.Fprintln(os.Stderr, "json parse error:", unmarshalErr)
		os.Exit(1)
	}

	if validateErr := rb.Validate(); validateErr != nil {
		fmt.Fprintln(os.Stderr, "validation error:", validateErr)
		os.Exit(1)
	}

	if *pretty {
		out, _ := json.MarshalIndent(rb, "", "  ")
		fmt.Println(string(out))
	} else {
		fmt.Println("OK ✔ runbook is valid")
	}
}
