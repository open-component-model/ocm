// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/open-component-model/ocm/cmds/ocm/app"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

func main() {
	logrus.Info("> Generate Docs for OCM CLI")

	if len(os.Args) != 2 { // expect 2 as the first one is the programm name
		logrus.Infof("Expected exactly one argument, but got %d", len(os.Args)-1)
		os.Exit(1)
	}
	outputDir := os.Args[1]

	// clear generated
	check(os.RemoveAll(outputDir))
	check(os.MkdirAll(outputDir, os.ModePerm))

	cmd := app.NewCliCommand(clictx.DefaultContext())
	cmd.DisableAutoGenTag = true
	check(GenMarkdownTree(cmd, outputDir))
	logrus.Infof("Successfully written docs to %s\n", outputDir)
}

func check(err error) {
	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}
}
