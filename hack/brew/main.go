package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"ocm.software/ocm/hack/brew/internal"
)

const (
	DefaultReleaseURL       = "https://github.com/open-component-model/ocm/releases/download"
	DefaultFormulaTemplate  = "hack/brew/internal/ocm_formula_template.rb.tpl"
	DefaultArchitectures    = "amd64,arm64"
	DefaultOperatingSystems = "darwin,linux"
)

func main() {
	version := flag.String("version", "", "version of the OCM formula")
	outputDir := flag.String("outputDirectory", ".", "path to the output directory")
	templateFile := flag.String("template", DefaultFormulaTemplate, "path to the template file")
	architecturesRaw := flag.String("arch", DefaultArchitectures, "comma-separated list of architectures")
	operatingSystemsRaw := flag.String("os", DefaultOperatingSystems, "comma-separated list of operating systems")
	releaseURL := flag.String("releaseURL", DefaultReleaseURL, "URL to fetch the release from")

	flag.Parse()

	if *version == "" {
		log.Fatalf("version is required")
	}

	if err := internal.GenerateVersionedHomebrewFormula(*version,
		strings.Split(*architecturesRaw, ","),
		strings.Split(*operatingSystemsRaw, ","),
		*releaseURL,
		*templateFile,
		*outputDir,
		os.Stdout,
	); err != nil {
		log.Fatalf("failed to generate formula: %v", err)
	}
}
