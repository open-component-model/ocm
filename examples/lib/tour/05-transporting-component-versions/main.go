package main

import (
	"fmt"
	"os"
	"strings"

	"ocm.software/ocm/examples/lib/helper"
)

// CFG is the path to the file containing the credentials
var CFG = "examples/lib/cred.yaml"

var current_version string

func init() {
	data, err := os.ReadFile("VERSION")
	if err != nil {
		panic("VERSION not found")
	}
	current_version = strings.TrimSpace(string(data))
}

func main() {
	arg := 1
	if len(os.Args) > 1 {
		if os.Args[1] == "--config" {
			if len(os.Args) > 2 {
				CFG = os.Args[2]
				arg = 3
			} else {
				fmt.Fprintf(os.Stderr, "error: config file missing\n")
				os.Exit(1)
			}
		}
	}
	cfg, err := helper.ReadConfig(CFG)
	if err == nil {
		cmd := "basic"

		if len(os.Args) > arg {
			cmd = os.Args[arg]
		}
		switch cmd {
		case "basic":
			err = TransportingComponentVersions(cfg)
		default:
			err = fmt.Errorf("unknown example %q", cmd)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
