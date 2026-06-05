// Command agent is the entrypoint for the Agentic AI Lab runtime.
package main

import (
	"fmt"
	"os"
)

// Version is set at build time via -ldflags.
var Version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	fmt.Printf("agentic-ai-lab agent %s — Go is ready.\n", Version)
	return nil
}
// trivial change
