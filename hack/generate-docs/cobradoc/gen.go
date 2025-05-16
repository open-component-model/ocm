package cobradoc

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"ocm.software/ocm/api/utils/cobrautils"
)

// Generate produces Hugo-ready CLI docs by extracting titles from headers
// and reorganizing flat Markdown into nested folders with front matter.
func Generate(name string, root *cobra.Command, outputDir string, clear bool) {
	// 1) Prepare output directory
	if clear {
		check(os.RemoveAll(outputDir))
	}
	flatDir := filepath.Join(outputDir, "_flat")
	check(os.RemoveAll(flatDir))
	check(os.MkdirAll(flatDir, os.ModePerm))

	// 2) Generate flat Markdown without front matter
	emptyPrepender := func(string) string { return "" }
	check(GenMarkdownTreeCustom(root, flatDir, emptyPrepender, cobrautils.LinkForPath))

	// 3) Reorganize files and add front matter
	entries, err := ioutil.ReadDir(flatDir)
	check(err)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// Read generated file
		src := filepath.Join(flatDir, entry.Name())
		data, err := ioutil.ReadFile(src)
		check(err)
		content := string(data)

		// Extract header line and body
		parts := strings.SplitN(content, "\n", 3)
		headLine := ""
		body := ""
		if len(parts) >= 1 {
			headLine = strings.TrimSpace(parts[0])
		}
		if len(parts) == 3 {
			body = parts[2]
		} else if len(parts) == 2 {
			body = parts[1]
		}

		// Determine raw identifier: drop prefix "ocm_"
		nameOnly := strings.TrimSuffix(entry.Name(), ".md")
		raw := strings.TrimPrefix(nameOnly, "ocm_")
		rawParts := strings.SplitN(raw, "_", 2)

		// Build URL
		url := "/docs/cli-reference/"
		switch {
		case raw == "":
			// root
		case len(rawParts) == 1:
			url = fmt.Sprintf("/docs/cli-reference/%s/", rawParts[0])
		case len(rawParts) == 2:
			url = fmt.Sprintf("/docs/cli-reference/%s/%s/", rawParts[0], rawParts[1])
		}

		// Determine destination directory and filename
		destDir := filepath.Join(outputDir, "ocm")
		destName := "_index.md"
		switch {
		case raw == "":
			// stays ocm/_index.md
		case len(rawParts) == 1:
			destDir = filepath.Join(outputDir, "ocm", rawParts[0])
			destName = "_index.md"
		case len(rawParts) == 2:
			destDir = filepath.Join(outputDir, "ocm", rawParts[0])
			destName = rawParts[1] + ".md"
		}
		check(os.MkdirAll(destDir, os.ModePerm))

		// Prepare title: strip '## ' and replace HTML entity/em dash
		title := strings.TrimPrefix(headLine, "## ")
		title = strings.ReplaceAll(title, "&mdash;", "-")
		title = strings.ReplaceAll(title, "—", "-")

		// Prepare linkTitle: remove leading "ocm ", then drop after em dash
		linkTitle := strings.TrimPrefix(title, "ocm ")
		if idx := strings.Index(linkTitle, " - "); idx != -1 {
			linkTitle = linkTitle[:idx]
		}
		if linkTitle == "" {
			linkTitle = strings.ReplaceAll(raw, "_", " ")
		}

		// Write enriched Markdown
		dst := filepath.Join(destDir, destName)
		out, err := os.Create(dst)
		check(err)
		fm := fmt.Sprintf(`---
`+
			`title: "%s"
`+
			`linkTitle: "%s"
`+
			`url: "%s"
`+
			`sidebar:
  collapsed: true
`+
			`menu:
  docs:
    name: "%s"
`+
			`---

`, title, linkTitle, url, linkTitle)
		_, err = out.WriteString(fm + body)
		check(err)
		out.Close()
	}

	// 4) Cleanup temporary files
	check(os.RemoveAll(flatDir))

	fmt.Printf("Successfully written %s docs to %s\n", name, outputDir)
}

// check aborts on error
func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
