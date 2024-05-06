package main

import (
	"fmt"
	"os"
)

// CFG is the path to the file containing the credentials
var CFG = "examples/lib/cred.yaml"

func main() {
	err := GettingStarted()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
