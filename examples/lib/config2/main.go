package main

import (
	"fmt"
	"os"
)

func main() {
	if err := UsingConfigs(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
