package cobradoc

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Generate(name string, cmd *cobra.Command, outputDir string, clear bool) {
	if clear {
		// clear generated
		check(os.RemoveAll(outputDir))
		check(os.MkdirAll(outputDir, os.ModePerm))
	}
	check(GenMarkdownTree(cmd, outputDir))
	//nolint: forbidigo // Intentional Println because this is a supplementary tool.
	fmt.Printf("Successfully written %s docs to %s\n", name, outputDir)
}

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}
