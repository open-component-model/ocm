package cobrautils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	_ "ocm.software/ocm/api/cli/config"
	"ocm.software/ocm/api/utils/out"
)

func TweakCommand(cmd *cobra.Command, ctx out.Context) *cobra.Command {
	if ctx != nil {
		cmd.UseLine()
		cmd.SetOut(ctx.StdOut())
		cmd.SetErr(ctx.StdErr())
		cmd.SetIn(ctx.StdIn())
	}
	cobra.AddTemplateFuncs(templatefuncs)
	CleanMarkdownUsageFunc(cmd)
	SupportNestedHelpFunc(cmd)
	cmd.SetHelpTemplate(HelpTemplate)
	cmd.SetUsageTemplate(UsageTemplate)
	return cmd
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
		cmd.Long = CleanMarkdown(cmd.Long)
		cmd.Example = CleanMarkdown(cmd.Example)
		defaultHelpFunc(cmd, s)
	})
}

var center = regexp.MustCompile(" *</?(pre|center)> *")

func CleanMarkdown(s string) string {
	if strings.HasPrefix(s, "##") {
		for strings.HasPrefix(s, "#") {
			s = s[1:]
		}
	}
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

	var r []string
	found := 0
	omitted := 0
	lastIndex := -1
	mask := false
	for _, l := range strings.Split(s, "\n") {
		if strings.Contains(l, "<pre>") {
			mask = true
		}
		if strings.Contains(l, "</pre>") {
			mask = false
		}
		if mask {
			found = 0
		} else {
			if strings.HasSuffix(l, "\\") {
				l = l[:len(l)-1]
				found = 0
			} else {
				t := strings.TrimSpace(l)
				if strings.HasPrefix(t, "- ") {
					index := strings.Index(l, "-")
					if omitted > 0 && lastIndex >= 0 && lastIndex > index {
						r = append(r, "")
					}
					lastIndex = index
					found = 1
				} else {
					if t == "" {
						found++
						if found > 1 {
							omitted++
							continue
						}
					} else {
						found = 0
					}
				}
			}
		}
		omitted = 0
		if !strings.HasPrefix(strings.TrimSpace(l), "-") {
			lastIndex = -1
		}
		r = append(r, l)
	}
	return strings.Join(r, "\n")
}

func GetHelpCommand(cmd *cobra.Command) *cobra.Command {
	for _, c := range cmd.Commands() {
		if c.Name() == "help" {
			return c
		}
	}
	return nil
}

// TweakHelpCommandFor generates a help command similar to the default cobra one,
// which forwards the additional arguments to the help function.
func TweakHelpCommandFor(c *cobra.Command) *cobra.Command {
	c.InitDefaultHelpCmd()
	defhelp := GetHelpCommand(c)
	c.SetHelpCommand(nil)
	c.RemoveCommand(defhelp)

	var help *cobra.Command
	help = &cobra.Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Long: `Help provides help for any command in the application.
Simply type ` + c.Name() + ` help [path to command] for full details.`,
		ValidArgsFunction: func(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			var completions []string
			cmd, _, e := c.Root().Find(args)
			if e != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			if cmd == nil {
				// Root help command.
				cmd = c.Root()
			}
			for _, subCmd := range cmd.Commands() {
				if subCmd.IsAvailableCommand() || subCmd == help {
					if strings.HasPrefix(subCmd.Name(), toComplete) {
						completions = append(completions, fmt.Sprintf("%s\t%s", subCmd.Name(), subCmd.Short))
					}
				}
			}
			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		Run: func(c *cobra.Command, args []string) {
			cmd, subargs, e := c.Parent().Find(args)
			if cmd == nil || e != nil {
				c.Printf("Unknown help topic %#q\n", args)
				cobra.CheckErr(c.Root().Usage())
			} else {
				cmd.InitDefaultHelpFlag()    // make possible 'help' flag to be shown
				cmd.InitDefaultVersionFlag() // make possible 'version' flag to be shown
				cmd.HelpFunc()(cmd, subargs)
			}
		},
		GroupID: defhelp.GroupID,
	}

	c.SetHelpCommand(help)
	c.AddCommand(help)
	return help
}
