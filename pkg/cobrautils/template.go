// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cobrautils

const HelpTemplate = "{{.CommandPath}} \u2014 {{title .Short}}" + `{{if .IsAvailableCommand}}

Synopsis:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{if or .Runnable (soleCommand .Use)}}{{if .HasAvailableLocalFlags}}{{.CommandPath}} [<options>] <sub-command> ...{{else}}{{.CommandPath}} <sub-command> ...{{end}}{{else}}{{.UseLine}}{{end}}{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if and .HasAvailableLocalFlags .IsAvailableCommand}}

Flags:
{{ flagUsages .LocalFlags | trimTrailingWhitespaces}}{{end}}{{if and .HasAvailableInheritedFlags .IsAvailableCommand}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}

Description:
{{with (or .Long .Short)}}{{. | substituteCommandLinks | trimTrailingWhitespaces | indent 2}}{{end}}{{if .HasExample}}

Examples:
{{.Example | indent 2}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
{{end}}
`

const UsageTemplate = `Synopsis:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}{{if .HasAvailableLocalFlags}}
  {{.CommandPath}} [<options>] <sub-command> ...{{else}}{{.CommandPath}} <sub-command> ...{{end}}{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{ flagUsages .LocalFlags | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
