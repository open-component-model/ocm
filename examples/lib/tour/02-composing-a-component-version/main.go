// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
)

// CFG is the path to the file containing the credentials
var CFG = "../examples/lib/cred.yaml"

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
		err = fmt.Errorf("unknown example %q")
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
