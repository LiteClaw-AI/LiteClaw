package main

import (
	"fmt"
	"os"

	"github.com/liteclaw/liteclaw/cmd"
)

var (
	version   = "1.0.0"
	buildDate = "unknown"
	commit    = "unknown"
)

func main() {
	if err := cmd.Execute(version, buildDate, commit); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
