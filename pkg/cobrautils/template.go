package cobrautils

const HelpTemplate = "{{commandPath .}} \u2014 {{title .Short}}" + `{{if .IsAvailableCommand}}

Synopsis:{{if .Runnable}}
  {{useLine .}}{{end}}{{if .HasAvailableSubCommands}}
  {{if or .Runnable (soleCommand .Use)}}{{if .HasAvailableLocalFlags}}{{commandPath .}} [<options>] <sub-command> ...{{else}}{{commandPath .}} <sub-command> ...{{end}}{{else}}{{useLine .}}{{end}}{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if and .HasAvailableLocalFlags .IsAvailableCommand}}

Flags:
{{ flagUsages .LocalFlags | trimTrailingWhitespaces}}{{end}}{{if and .HasAvailableInheritedFlags .IsAvailableCommand}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}

Description:
{{with (or .Long .Short)}}{{. | substituteCommandLinks | trimTrailingWhitespaces | indent 2}}{{end}}{{if .HasAvailableSubCommands}}
  Use {{commandPath .}} <command> -h for additional help.
{{end}}{{if .HasExample}}

Examples:
{{.Example | indent 2}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad commandPath . .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
{{end}}
`

const UsageTemplate = `Synopsis:{{if .Runnable}}
  {{useLine .}}{{end}}{{if .HasAvailableSubCommands}}{{if .HasAvailableLocalFlags}}
  {{commandPath .}} [<options>] <sub-command> ...{{else}}{{commandPath .}} <sub-command> ...{{end}}{{end}}{{if gt (len .Aliases) 0}}

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
  {{rpad commandPath . .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{commandPath .}} [command] --help" for more information about a command.{{end}}
`
