package main

import (
	"fmt"
	"os"
)

func main() {
	err := ResourceManagement()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
