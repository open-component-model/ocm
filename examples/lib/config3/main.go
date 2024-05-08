package main

import (
	"fmt"
	"os"
)

// CFG is the path to the file containing the credentials
var CFG = "examples/lib/cred.yaml"

func main() {
	if len(os.Args) > 1 {
		CFG = os.Args[1]
	}
	if err := UsingConfigs(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
