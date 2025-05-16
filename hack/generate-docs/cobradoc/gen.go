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

// Generate produces Hugo-ready CLI docs in nested folders.
// Input: Cobra command tree, outputDir, clear flag.
// Output under outputDir/ocm: hierarchical structure:
//
//	ocm/_index.md
//	ocm/<command>/_index.md
//	ocm/<command>/<subcommand>.md
//
// Each front matter includes title, url, and sidebar.collapsed
func Generate(name string, root *cobra.Command, outputDir string, clear bool) {
	// Clean output and prepare
	if clear {
		check(os.RemoveAll(outputDir))
	}

	// filePrepender injects front matter with title, url, and sidebar
	filePrepender := func(filename string) string {
		base := filepath.Base(filename)
		nameOnly := strings.TrimSuffix(base, ".md")

		// Title: drop "ocm_" prefix, replace separators, lowercase
		titleKey := strings.TrimPrefix(nameOnly, "ocm_")
		titleKey = strings.ReplaceAll(titleKey, "_", " ")
		titleKey = strings.ReplaceAll(titleKey, "-", " ")
		title := strings.ToLower(titleKey)

		// URL: build based on segments
		parts := strings.Split(nameOnly, "_")
		var url string
		switch len(parts) {
		case 1:
			// ocm.md
			url = "/docs/cli-reference/"
		case 2:
			// ocm_<command>.md
			url = fmt.Sprintf("/docs/cli-reference/%s/", parts[1])
		default:
			// ocm_<command>_<sub>.md
			url = fmt.Sprintf("/docs/cli-reference/%s/%s/", parts[1], parts[2])
		}
		// return front matter including sidebar collapsed
		return fmt.Sprintf(
			"---\n"+
				"title: \"%s\"\n"+
				"url: \"%s\"\n"+
				"sidebar:\n  collapsed: true\n"+
				"---\n\n",
			title, url,
		)
	}

	// 1) Generate flat markdown into temp directory
	flatDir := filepath.Join(outputDir, "_flat")
	check(os.RemoveAll(flatDir))
	check(os.MkdirAll(flatDir, os.ModePerm))
	check(GenMarkdownTreeCustom(root, flatDir, filePrepender, cobrautils.LinkForPath))

	// 2) Rearrange files into nested structure
	entries, err := ioutil.ReadDir(flatDir)
	check(err)
	for _, fi := range entries {
		if fi.IsDir() {
			continue
		}
		fname := fi.Name()
		nameOnly := strings.TrimSuffix(fname, ".md")
		parts := strings.Split(nameOnly, "_")[1:] // skip "ocm"

		var destDir, destName string
		switch len(parts) {
		case 0:
			// root ocm.md
			destDir = filepath.Join(outputDir, "ocm")
			destName = "_index.md"
		case 1:
			// command
			destDir = filepath.Join(outputDir, "ocm", parts[0])
			destName = "_index.md"
		default:
			// subcommand
			destDir = filepath.Join(outputDir, "ocm", parts[0])
			sub := strings.Join(parts[1:], "_")
			destName = sub + ".md"
		}
		check(os.MkdirAll(destDir, os.ModePerm))
		srcPath := filepath.Join(flatDir, fname)
		dstPath := filepath.Join(destDir, destName)
		check(os.Rename(srcPath, dstPath))
	}

	// 3) Cleanup
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
