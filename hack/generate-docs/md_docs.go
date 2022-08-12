//Copyright 2015 Red Hat Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/open-component-model/ocm/pkg/cobrautils"
	"github.com/spf13/cobra"
)

const fmTmpl = `
---
title: %s
url: %s
date: %s
draft: false
images: []
menu:
  docs:
    parent: cli-reference
toc: true
---
`

func printOptions(buf *bytes.Buffer, cmd *cobra.Command, name string) error {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("### Options\n\n```\n")
		flags.PrintDefaults()
		buf.WriteString("```\n\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("### Options inherited from parent commands\n\n```\n")
		parentFlags.PrintDefaults()
		buf.WriteString("```\n\n")
	}
	return nil
}

// GenMarkdown creates markdown output.
func GenMarkdown(cmd *cobra.Command, w io.Writer) error {
	return GenMarkdownCustom(cmd, w, cobrautils.LinkForPath)
}

// GenMarkdownCustom creates custom markdown output.
func GenMarkdownCustom(cmd *cobra.Command, w io.Writer, linkHandler func(string) string) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)
	name := cmd.CommandPath()

	buf.WriteString("## " + name + " &mdash; " + strings.Title(cmd.Short) + "\n\n")

	if cmd.Runnable() || cmd.HasAvailableSubCommands() {
		buf.WriteString("### Synopsis\n\n")
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", UseLine(cmd)))
	}

	if cmd.IsAvailableCommand() {
		if err := printOptions(buf, cmd, name); err != nil {
			return err
		}
	}

	if len(cmd.Long) > 0 {
		buf.WriteString("### Description\n\n")
		buf.WriteString(cobrautils.SubstituteCommandLinks(cmd.Long, cobrautils.FormatLinkWithHandler(linkHandler)) + "\n\n")
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("### Examples\n\n")
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmd.Example))
	}

	if hasSeeAlso(cmd) {
		header := cmd.HasHelpSubCommands() && cmd.HasAvailableSubCommands()
		buf.WriteString("### SEE ALSO\n\n")
		if cmd.HasParent() {
			header = true
			buf.WriteString("##### Parents\n\n")
			parent := cmd
			for parent.HasParent() {
				parent = parent.Parent()
				pname := parent.CommandPath()
				buf.WriteString(fmt.Sprintf("* [%s](%s)\t &mdash; %s\n", pname, linkHandler(parent.CommandPath()), parent.Short))
			}
			cmd.VisitParents(func(c *cobra.Command) {
				if c.DisableAutoGenTag {
					cmd.DisableAutoGenTag = c.DisableAutoGenTag
				}
			})
		}

		children := cmd.Commands()
		sort.Sort(byName(children))

		subheader := false
		for _, child := range children {
			if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
				continue
			}
			if header && !subheader {
				buf.WriteString("\n\n##### Sub Commands\n\n")
				subheader = true
			}
			cname := name + " " + "<b>" + child.Name() + "</b>"
			buf.WriteString(fmt.Sprintf("* [%s](%s)\t &mdash; %s\n", cname, linkHandler(child.CommandPath()), child.Short))
		}
		buf.WriteString("\n")

		subheader = false
		for _, child := range children {
			if !child.IsAdditionalHelpTopicCommand() {
				continue
			}
			if header && !subheader {
				buf.WriteString("\n\n##### Additional Help Topics\n\n")
				subheader = true
			}
			cname := name + " " + "<b>" + child.Name() + "</b>"
			buf.WriteString(fmt.Sprintf("* [%s](%s)\t &mdash; %s\n", cname, linkHandler(child.CommandPath()), child.Short))
		}
		if subheader {
			buf.WriteString("\n")
		}

	}
	if !cmd.DisableAutoGenTag {
		buf.WriteString("###### Auto generated by spf13/cobra on " + time.Now().Format("2-Jan-2006") + "\n")
	}
	_, err := buf.WriteTo(w)
	return err
}

// GenMarkdownTree will generate a markdown page for this command and all
// descendants in the directory given. The header may be nil.
// This function may not work correctly if your command names have `-` in them.
// If you have `cmd` with two subcmds, `sub` and `sub-third`,
// and `sub` has a subcommand called `third`, it is undefined which
// help output will be in the file `cmd-sub-third.1`.
func GenMarkdownTree(cmd *cobra.Command, dir string) error {
	identity := cobrautils.LinkForPath
	frontmatter := func(filename string) string {
		now := time.Now().Format(time.RFC3339)
		name := filepath.Base(filename)
		base := strings.TrimSuffix(name, path.Ext(name))
		url := "/docs/cli-reference/" + strings.ToLower(base) + "/"
		return fmt.Sprintf(fmTmpl, base, url, now)
	}

	return GenMarkdownTreeCustom(cmd, dir, frontmatter, identity)
}

// GenMarkdownTreeCustom is the the same as GenMarkdownTree, but
// with custom filePrepender and linkHandler.
func GenMarkdownTreeCustom(cmd *cobra.Command, dir string, filePrepender, linkHandler func(string) string) error {
	for _, c := range cmd.Commands() {
		if c.Name() == "configfile" {
			strings.TrimSpace(c.Name())
		}
		if !c.IsAvailableCommand() && !c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := GenMarkdownTreeCustom(c, dir, filePrepender, linkHandler); err != nil {
			return err
		}
	}

	basename := strings.Replace(cmd.CommandPath(), " ", "_", -1) + ".md"
	filename := filepath.Join(dir, basename)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.WriteString(f, filePrepender(filename)); err != nil {
		return err
	}
	if err := GenMarkdownCustom(cmd, f, linkHandler); err != nil {
		return err
	}
	return nil
}
