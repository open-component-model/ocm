package cobradoc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"ocm.software/ocm/api/utils/cobrautils"
)

// Generate produces Markdown docs for the given Cobra command tree,
// injecting YAML front matter so Hugo can parse titles and menus.
func Generate(name string, cmd *cobra.Command, outputDir string, clear bool) {
	// Clear previous docs if requested
	if clear {
		check(os.RemoveAll(outputDir))
		check(os.MkdirAll(outputDir, os.ModePerm))
	}

	// filePrepender injects front matter based on filename
	filePrepender := func(filename string) string {
		base := filepath.Base(filename)
		nameOnly := strings.TrimSuffix(base, ".md")

		// Build lowercase title
		titleKey := strings.TrimPrefix(nameOnly, "ocm_")
		titleKey = strings.ReplaceAll(titleKey, "_", " ")
		titleKey = strings.ReplaceAll(titleKey, "-", " ")
		title := strings.ToLower(titleKey)

		// Determine parent: subcommands (len>=3) under main command
		parts := strings.Split(nameOnly, "_")
		parent := "cli-reference"
		if len(parts) >= 3 {
			parent = parts[1]
		}

		// YAML front matter block
		return fmt.Sprintf(
			"---\n"+
				"title: \"%s\"\n"+
				"menu:\n"+
				"  docs:\n"+
				"    parent: %s\n"+
				"---\n",
			title, parent,
		)
	}

	// Generate docs with injected front matter and correct linking
	check(GenMarkdownTreeCustom(cmd, outputDir, filePrepender, cobrautils.LinkForPath))

	fmt.Printf("Successfully written %s docs to %s\n", name, outputDir)
}

// check aborts on error
func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
}
