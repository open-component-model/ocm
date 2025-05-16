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
	// 1) Prepare output directories
	if clear {
		check(os.RemoveAll(outputDir))
	}
	flatDir := filepath.Join(outputDir, "_flat")
	check(os.RemoveAll(flatDir))
	check(os.MkdirAll(flatDir, os.ModePerm))

	// 2) Generate flat Markdown without front matter
	emptyPrepender := func(string) string { return "" }
	check(GenMarkdownTreeCustom(root, flatDir, emptyPrepender, cobrautils.LinkForPath))

	// 3) Reorganize files
	entries, err := ioutil.ReadDir(flatDir)
	check(err)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// Read flat file
		src := filepath.Join(flatDir, entry.Name())
		data, err := ioutil.ReadFile(src)
		check(err)
		content := string(data)

		// Extract and remove first header line
		parts := strings.SplitN(content, "\n", 3)
		headLine := ""
		body := ""
		if len(parts) > 0 {
			headLine = strings.TrimSpace(parts[0])
		}
		if len(parts) >= 3 {
			body = parts[2]
		} else if len(parts) == 2 {
			body = parts[1]
		}

		// Determine raw command path
		nameOnly := strings.TrimSuffix(entry.Name(), ".md")
		raw := strings.TrimPrefix(nameOnly, "ocm_")
		rawParts := strings.SplitN(raw, "_", 2)

		// Build URL
		url := "/docs/cli-reference/"
		switch {
		case raw == "":
			// root ocm.md
			url = "/docs/cli-reference/"
		case len(rawParts) == 1:
			url = fmt.Sprintf("/docs/cli-reference/%s/", rawParts[0])
		case len(rawParts) == 2:
			url = fmt.Sprintf("/docs/cli-reference/%s/%s/", rawParts[0], rawParts[1])
		}

		// Destination path
		destDir := filepath.Join(outputDir, "ocm")
		destName := "_index.md"
		switch {
		case raw == "":
			// stays destDir/ocm/_index.md
		case len(rawParts) == 1:
			destDir = filepath.Join(destDir, rawParts[0])
			destName = "_index.md"
		case len(rawParts) == 2:
			destDir = filepath.Join(destDir, rawParts[0])
			destName = rawParts[1] + ".md"
		}
		check(os.MkdirAll(destDir, os.ModePerm))

		// Prepare front matter title
		title := strings.TrimPrefix(headLine, "## ")

		// Write new file
		dst := filepath.Join(destDir, destName)
		out, err := os.Create(dst)
		check(err)
		fm := fmt.Sprintf("---\n"+
			"title: \"%s\"\n"+
			"url: \"%s\"\n"+
			"sidebar:\n  collapsed: true\n"+
			"---\n\n", title, url)
		_, err = out.WriteString(fm + body)
		check(err)
		out.Close()
	}

	// 4) Cleanup
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
