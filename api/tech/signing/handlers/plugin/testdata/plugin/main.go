package main

import (
	"os"

	"ocm.software/ocm/api/tech/signing/handlers/plugin/testdata/plugin/app"
)

func main() {
	err := app.Run(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
