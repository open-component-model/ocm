package main

import (
	"fmt"
	"os"
)

func main() {
	if err := CTFExample(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
