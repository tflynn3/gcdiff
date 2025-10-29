package main

import (
	"fmt"
	"os"

	"github.com/tflynn3/gcdiff/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
