// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package cobrautils

import (
	"regexp"
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

var center = regexp.MustCompile(" *</?(pre|center)> *\n?")

func cleanMarkdown(s string) string {
	s = strings.ReplaceAll(s, "<code>", "\u00ab")
	s = strings.ReplaceAll(s, "</code>", "\u00bb")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "**", "")
	s = string(center.ReplaceAll([]byte(s), nil))
	return s
}
