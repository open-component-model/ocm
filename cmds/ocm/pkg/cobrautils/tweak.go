// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cobrautils

import (
	"strings"

	"github.com/open-component-model/ocm/pkg/out"
	"github.com/spf13/cobra"

	_ "github.com/open-component-model/ocm/cmds/ocm/clictx/config"
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
		defaultHelpFunc(cmd, s)
	})
}

func cleanMarkdown(s string) string {
	s = strings.ReplaceAll(s, "<pre>", "")
	s = strings.ReplaceAll(s, "</pre>", "")
	s = strings.ReplaceAll(s, "<code>", "\u00ab")
	s = strings.ReplaceAll(s, "</code>", "\u00bb")
	s = strings.ReplaceAll(s, "<center>\n", "")
	s = strings.ReplaceAll(s, "<center>", "")
	s = strings.ReplaceAll(s, "\n</center>", "")
	s = strings.ReplaceAll(s, "</center>", "")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "**", "")
	return s
}
