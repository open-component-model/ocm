// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cobrautils

import (
	"regexp"
	"strings"

	_ "github.com/open-component-model/ocm/pkg/contexts/clictx/config"

	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/out"
)

func TweakCommand(cmd *cobra.Command, ctx out.Context) {
	if ctx != nil {
		cmd.SetOut(ctx.StdOut())
		cmd.SetErr(ctx.StdErr())
		cmd.SetIn(ctx.StdIn())
	}
	cobra.AddTemplateFuncs(templatefuncs)
	CleanMarkdownUsageFunc(cmd)
	SupportNestedHelpFunc(cmd)
	cmd.SetHelpTemplate(HelpTemplate)
	cmd.SetUsageTemplate(UsageTemplate)
}

// SupportNestedHelpFunc adds support help evaluation of given nested command path.
func SupportNestedHelpFunc(cmd *cobra.Command) {
	defaultHelpFunc := cmd.HelpFunc()
	cmd.SetHelpFunc(func(cmd *cobra.Command, s []string) {
		_, s, err := cmd.Root().Find(s)
		if err == nil {
			for i := 1; i < len(s); i++ {
				var next *cobra.Command
				for _, c := range cmd.Commands() {
					if c.Name() == s[i] {
						next = c
						break
					}
				}
				if next == nil {
					break
				} else {
					cmd = next
				}
			}
		}
		defaultHelpFunc(cmd, s)
	})
}

// CleanMarkdownUsageFunc removes markdown tags from the long usage of the command.
// With this func it is possible to generate the markdown docs but still have readable commandline help func.
func CleanMarkdownUsageFunc(cmd *cobra.Command) {
	defaultHelpFunc := cmd.HelpFunc()
	cmd.SetHelpFunc(func(cmd *cobra.Command, s []string) {
		cmd.Long = cleanMarkdown(cmd.Long)
		cmd.Example = cleanMarkdown(cmd.Example)
		defaultHelpFunc(cmd, s)
	})
}

var center = regexp.MustCompile(" *</?(pre|center)> *\n?")

func cleanMarkdown(s string) string {
	s = strings.ReplaceAll(s, "<code>", "\u00ab")
	s = strings.ReplaceAll(s, "</code>", "\u00bb")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "<em>", "")
	s = strings.ReplaceAll(s, "</em>", "")
	s = strings.ReplaceAll(s, "<EXAMPLE>", "")
	s = strings.ReplaceAll(s, "</EXAMPLE>", "")
	s = string(center.ReplaceAll([]byte(s), nil))
	return s
}
