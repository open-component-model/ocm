package cobrautils

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func LinkForCmd(cmd *cobra.Command) string {
	return LinkForPath(cmd.CommandPath())
}

func LinkForPath(path string) string {
	link := path + ".md"
	link = strings.ReplaceAll(link, " ", "_")
	return link
}

func FormatLink(pname string, linkhandler func(string) string) string {
	return fmt.Sprintf("[%s](%s)", pname, linkhandler((pname)))
}

func FormatLinkWithHandler(linkhandler func(string) string) func(string) string {
	return func(pname string) string {
		return FormatLink(pname, linkhandler)
	}
}

func SubstituteCommandLinks(desc string, linkformat func(string) string) ([]string, string) {
	var links []string
	for {
		link := strings.Index(desc, "<CMD>")
		if link < 0 {
			return links, desc
		}
		end := strings.Index(desc, "</CMD>")
		if end < 0 {
			panic("missing </CMD> in:\n" + desc)
		}
		path := desc[link+5 : end]
		links = append(links, path)
		desc = desc[:link] + linkformat(path) + desc[end+6:]
	}
}
