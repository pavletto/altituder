// This file is kept for backward compatibility.
// The actual CLI application is in cmd/altituder/
// Build with: go build -o altituder ./cmd/altituder
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "Please build from cmd/altituder: go build -o altituder ./cmd/altituder")
	os.Exit(1)
}
