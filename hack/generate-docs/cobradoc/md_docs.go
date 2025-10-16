package cobradoc

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"ocm.software/ocm/api/utils/cobrautils"
	"ocm.software/ocm/api/utils/cobrautils/groups"
)

func printOptionGroups(buf *bytes.Buffer, title string, flags *pflag.FlagSet) {
	buf.WriteString(fmt.Sprintf("### %s\n\n", title))
	groups := groups.GroupedFlagUsagesWrapped(flags, 0)
	if len(groups) > 1 {
		for _, g := range groups {
			if g.Title != "" {
				buf.WriteString("\n#### " + g.Title + "\n\n")
			}
			buf.WriteString("```text\n")
			buf.WriteString(g.Usages)
			buf.WriteString("```\n\n")
		}
	} else {
		buf.WriteString("```text\n")
		buf.WriteString(groups[0].Usages)
		buf.WriteString("```\n\n")
	}
}

func printOptions(buf *bytes.Buffer, cmd *cobra.Command, name string) error {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		printOptionGroups(buf, "Options", flags)
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		printOptionGroups(buf, "Options inherited from parent commands", parentFlags)
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
		buf.WriteString(fmt.Sprintf("```bash\n%s\n```\n\n", UseLine(cmd)))
		if len(cmd.Aliases) > 0 {
			buf.WriteString("#### Aliases\n\n")
			cmd.NameAndAliases()
			buf.WriteString(fmt.Sprintf("```text\n%s\n```\n\n", cmd.NameAndAliases()))
		}
	}

	if cmd.IsAvailableCommand() {
		if err := printOptions(buf, cmd, name); err != nil {
			return err
		}
	}

	var links []string
	if len(cmd.Long) > 0 {
		var desc string

		desc = strings.ReplaceAll(desc, "\\\n", "\n")
		links, desc = cobrautils.SubstituteCommandLinks(cmd.Long, cobrautils.FormatLinkWithHandler(linkHandler))
		buf.WriteString("### Description\n")
		buf.WriteString(desc + "\n")
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("### Examples\n\n")

		if strings.Contains(cmd.Example, "<pre>") {
			buf.WriteString(fmt.Sprintf("\n%s\n\n", SanitizePre(cmd.Example)))
		} else {
			buf.WriteString(fmt.Sprintf("```%s\n%s\n```\n\n", ExampleCodeStyle(cmd), strings.TrimSpace(cmd.Example)))
		}
	}

	if len(links) > 0 || hasSeeAlso(cmd) {
		var shown_links []string
		cnt := 0
		if cmd.HasHelpSubCommands() {
			cnt++
		}
		if cmd.HasAvailableSubCommands() {
			cnt++
		}
		if len(links) > 0 {
			cnt++
		}
		header := cnt > 1
		buf.WriteString("### SEE ALSO\n\n")
		if cmd.HasParent() {
			header = true
			buf.WriteString("#### Parents\n\n")
			parent := cmd
			for parent.HasParent() {
				parent = parent.Parent()
				pname := parent.CommandPath()
				path := parent.CommandPath()
				shown_links = append(shown_links, path)
				buf.WriteString(fmt.Sprintf("* [%s](%s)\t &mdash; %s\n", pname, linkHandler(path), parent.Short))
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
			if OverviewOnly(child) || !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
				continue
			}
			if header && !subheader {
				buf.WriteString("\n\n##### Sub Commands\n\n")
				subheader = true
			}
			path := DocuCommandPath(child)
			cname := name + " " + "<b>" + child.Name() + "</b>"

			if OverviewOnly(cmd) {
				buf.WriteString(fmt.Sprintf("* %s\t &mdash; %s\n", cname, child.Short))
			} else {
				shown_links = append(shown_links, path)
				buf.WriteString(fmt.Sprintf("* [%s](%s)\t &mdash; %s\n", cname, linkHandler(path), child.Short))
			}
		}
		buf.WriteString("\n")

		subheader = false
		for _, child := range children {
			if !OverviewOnly(child) {
				continue
			}
			if header && !subheader {
				buf.WriteString("\n\n##### Area Overview\n\n")
				subheader = true
			}
			path := child.CommandPath()
			shown_links = append(shown_links, path)
			cname := name + " " + "<b>" + child.Name() + "</b>"
			buf.WriteString(fmt.Sprintf("* [%s](%s)\t &mdash; %s\n", cname, linkHandler(path), child.Short))
		}

		subheader = false
		for _, child := range children {
			if !child.IsAdditionalHelpTopicCommand() {
				continue
			}
			if header && !subheader {
				buf.WriteString("\n\n##### Additional Help Topics\n\n")
				subheader = true
			}
			path := DocuCommandPath(child)
			shown_links = append(shown_links, path)
			cname := FormattedCommandPath(child)
			buf.WriteString(fmt.Sprintf("* [%s](%s)\t &mdash; %s\n", cname, linkHandler(path), child.Short))
		}

		subheader = false
		root := cmd.Root()
	nextlink:
		for _, link := range links {
			var sub *cobra.Command
			for _, s := range shown_links {
				if s == link {
					continue nextlink
				}
			}
			path := strings.Split(link, " ")
			if len(path) > 1 {
				sub = root
			outer:
				for _, c := range path[1:] {
					for _, n := range sub.Commands() {
						if n.Name() == c {
							sub = n
							continue outer
						}
					}
					sub = nil
					break
				}
			}

			if header && !subheader {
				buf.WriteString("\n\n##### Additional Links\n\n")
				subheader = true
			}
			cname := "<b>" + link + "</b>"
			if sub != nil {
				buf.WriteString(fmt.Sprintf("* [%s](%s)\t &mdash; %s\n", cname, linkHandler(link), sub.Short))
			} else {
				buf.WriteString(fmt.Sprintf("* [%s](%s)\n", cname, linkHandler(link)))
			}
		}

		if subheader {
			buf.WriteString("\n")
		}
	}
	if !cmd.DisableAutoGenTag {
		buf.WriteString("###### Auto generated by spf13/cobra on " + time.Now().Format("2-Jan-2006") + "\n")
	}
	_, err := buf.WriteTo(OmitTrailingSpaces(w))
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
	emptyStr := func(s string) string { return "" }
	return GenMarkdownTreeCustom(cmd, dir, emptyStr, identity)
}

func OverviewOnly(cmd *cobra.Command) bool {
	if cmd.Annotations == nil {
		return false
	}
	_, ok := cmd.Annotations["overview"]
	return ok
}

func DocuCommandPath(cmd *cobra.Command) string {
	if cmd.Annotations != nil {
		if p, ok := cmd.Annotations["commandPath"]; ok {
			return p
		}
	}
	return cmd.CommandPath()
}

func ExampleCodeStyle(cmd *cobra.Command) string {
	if cmd.Annotations != nil {
		if p, ok := cmd.Annotations["ExampleCodeStyle"]; ok {
			return p
		}
	}
	return "text"
}

func FormattedCommandPath(cmd *cobra.Command) string {
	path := DocuCommandPath(cmd)

	h := strings.Split(path, " ")
	if h[len(h)-1] == cmd.Name() {
		return strings.Join(h[:len(h)-1], " ") + " " + "<b>" + cmd.Name() + "</b>"
	}
	return path
}

// GenMarkdownTreeCustom is the same as GenMarkdownTree, but
// with custom filePrepender and linkHandler.
func GenMarkdownTreeCustom(cmd *cobra.Command, dir string, filePrepender, linkHandler func(string) string) error {
	if !OverviewOnly(cmd) {
		for _, c := range cmd.Commands() {
			if c.Name() == "configfile" {
				// I have no idea what this supposed to do.
				// TrimSpace is a pure function but its return value is ignored.
				strings.TrimSpace(c.Name())
			}
			if !c.IsAvailableCommand() && !c.IsAdditionalHelpTopicCommand() {
				continue
			}
			if err := GenMarkdownTreeCustom(c, dir, filePrepender, linkHandler); err != nil {
				return err
			}
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

func GetCommandForPath(cmd *cobra.Command, path string) *cobra.Command {
	seq := strings.Split(path, " ")
	if len(seq) > 1 {
		seq = seq[1:]
	}
	cmd = cmd.Root()
outer:
	for _, s := range seq {
		for _, c := range cmd.Commands() {
			if c.Name() == s {
				cmd = c
				continue outer
			}
		}
		return nil
	}
	return cmd
}
