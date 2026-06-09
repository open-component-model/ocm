package internal

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const ClassName = "Ocm"

// GenerateHomebrewFormula generates the canonical Homebrew formula and a
// version-pinned formula for the given version. It fetches the SHA256 digest
// for each os/arch combination and renders both files into outputDir:
//
//   - Formula/ocm.rb (class Ocm) — the canonical, unversioned formula. Overwritten
//     on every release so `brew install ocm` and `brew upgrade` operate in place.
//   - Formula/ocm@<version>.rb (class OcmAT<digits>) — opt-in pin to a specific
//     version, for users who want `brew install ocm@X.Y.Z`.
func GenerateHomebrewFormula(
	version string,
	architectures []string,
	operatingSystems []string,
	releaseURL string,
	templateFile string,
	outputDir string,
	writer io.Writer,
) error {
	values := map[string]string{
		"ReleaseURL": releaseURL,
		"Version":    version,
	}

	for _, targetOs := range operatingSystems {
		for _, arch := range architectures {
			digest, err := FetchDigestFromGithubRelease(releaseURL, version, targetOs, arch)
			if err != nil {
				return fmt.Errorf("failed to fetch digest for %s/%s: %w", targetOs, arch, err)
			}
			values[fmt.Sprintf("%s_%s_sha256", targetOs, arch)] = digest
		}
	}

	if err := GenerateFormula(templateFile, outputDir, "ocm.rb", ClassName, values, writer); err != nil {
		return fmt.Errorf("failed to generate canonical formula: %w", err)
	}
	if _, err := io.WriteString(writer, "\n"); err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}
	versionedFile := fmt.Sprintf("ocm@%s.rb", version)
	versionedClass := fmt.Sprintf("%sAT%s", ClassName, strings.ReplaceAll(version, ".", ""))
	if err := GenerateFormula(templateFile, outputDir, versionedFile, versionedClass, values, writer); err != nil {
		return fmt.Errorf("failed to generate versioned formula: %w", err)
	}

	return nil
}

// FetchDigestFromGithubRelease retrieves the SHA256 digest for a specific version, operating system, and architecture
// from the given release URL.
func FetchDigestFromGithubRelease(releaseURL, version, targetOs, arch string) (_ string, err error) {
	url := fmt.Sprintf("%s/v%s/ocm-%s-%s-%s.tar.gz.sha256", releaseURL, version, version, targetOs, arch)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get digest: %w", err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	digestBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read digest: %w", err)
	}

	return strings.TrimSpace(string(digestBytes)), nil
}

// GenerateFormula renders templateFile into outputDir/outputFile using the
// provided className for the Ruby class and values for template substitution.
func GenerateFormula(templateFile, outputDir, outputFile, className string, values map[string]string, writer io.Writer) error {
	tmpl, err := template.New(filepath.Base(templateFile)).Funcs(template.FuncMap{
		"classname": func() string {
			return className
		},
	}).ParseFiles(templateFile)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	if err := ensureDirectory(outputDir); err != nil {
		return err
	}

	formula, err := os.Create(filepath.Join(outputDir, outputFile))
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer formula.Close()

	if err := tmpl.Execute(formula, values); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	if _, err := io.WriteString(writer, formula.Name()); err != nil {
		return fmt.Errorf("failed to write output file path: %w", err)
	}

	return nil
}

// ensureDirectory checks if a directory exists and creates it if it does not.
func ensureDirectory(dir string) error {
	fi, err := os.Stat(dir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to stat directory: %w", err)
	} else if !fi.IsDir() {
		return fmt.Errorf("path is not a directory")
	}
	return nil
}
