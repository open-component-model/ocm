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
	cmd.SetHelpTemplate(HelpTemplate)

	cmd.SetUsageTemplate(UsageTemplate)
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
	s = strings.ReplaceAll(s, "<code>", "")
	s = strings.ReplaceAll(s, "</code>", "")
	s = strings.ReplaceAll(s, "<center>", "")
	s = strings.ReplaceAll(s, "</center>", "")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	return s
}
