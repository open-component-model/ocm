// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"strings"
)

// CFG is the path to the file containing the credentials
var CFG = "../examples/lib/cred.yaml"

var current_version string

func init() {
	data, err := os.ReadFile("VERSION")
	if err != nil {
		panic("VERSION not found")
	}
	current_version = strings.TrimSpace(string(data))
}

func main() {
	cmd := "basic"

	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	var err error
	switch cmd {
	case "basic":
		err = ComposingAComponentVersionA()
	case "compose":
		err = ComposingAComponentVersionB()
	default:
		err = fmt.Errorf("unknown example %q", cmd)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
