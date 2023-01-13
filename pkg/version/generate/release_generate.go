// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/open-component-model/ocm/pkg/version"
)

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("missing argument")
	}

	_ = semver.MustParse(version.ReleaseVersion)

	cmd := os.Args[1]
	//nolint:forbidigo // Logger not needed for this command.
	if cmd == "print-version" {
		fmt.Print(version.ReleaseVersion)
	}
}
