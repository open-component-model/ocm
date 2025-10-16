package internal

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed ocm_formula_template.rb.tpl
var tplFile []byte

//go:embed testdata/expected_formula.rb
var expectedResolved []byte

func TestGenerateVersionedHomebrewFormula(t *testing.T) {
	version := "1.0.0"
	architectures := []string{"amd64", "arm64"}
	operatingSystems := []string{"darwin", "linux"}
	outputDir := t.TempDir()

	templateFile := filepath.Join(outputDir, "ocm_formula_template.rb.tpl")
	if err := os.WriteFile(templateFile, tplFile, os.ModePerm); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	dummyDigest := "dummy-digest"
	// Mock server to simulate fetching digests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(dummyDigest))
	}))
	defer server.Close()
	expectedResolved = bytes.ReplaceAll(expectedResolved, []byte("$$TEST_SERVER$$"), []byte(server.URL))

	var buf bytes.Buffer

	err := GenerateVersionedHomebrewFormula(
		version,
		architectures,
		operatingSystems,
		server.URL,
		templateFile,
		outputDir,
		&buf,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	file := buf.String()

	fi, err := os.Stat(file)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if fi.Size() == 0 {
		t.Fatalf("expected file to be non-empty")
	}
	if filepath.Ext(file) != ".rb" {
		t.Fatalf("expected file to have .rb extension")
	}
	if !strings.Contains(file, version) {
		t.Fatalf("expected file to contain version")
	}

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if string(data) != string(expectedResolved) {
		t.Fatalf("expected %s, got %s", string(expectedResolved), string(data))
	}
}

func TestFetchDigest(t *testing.T) {
	expectedDigest := "dummy-digest"
	version := "1.0.0"
	targetOS, arch := "linux", "amd64"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1.0.0/ocm-1.0.0-linux-amd64.tar.gz.sha256" {
			t.Fatalf("expected path %s, got %s", fmt.Sprintf("/v%[1]s/ocm-%[1]s-%s-%s.tar.gz.sha256", version, targetOS, arch), r.URL.Path)
		}
		w.Write([]byte(expectedDigest))
	}))
	defer server.Close()

	digest, err := FetchDigestFromGithubRelease(server.URL, version, targetOS, arch)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if digest != expectedDigest {
		t.Fatalf("expected %s, got %s", expectedDigest, digest)
	}
}

func TestGenerateFormula(t *testing.T) {
	templateContent := `class {{ classname }} < Formula
version "{{ .Version }}"
end`
	templateFile := "test_template.rb.tpl"
	if err := os.WriteFile(templateFile, []byte(templateContent), 0o644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}
	defer os.Remove(templateFile)

	outputDir := t.TempDir()
	values := map[string]string{"Version": "1.0.0"}

	var buf bytes.Buffer

	if err := GenerateFormula(templateFile, outputDir, "1.0.0", values, &buf); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if buf.String() == "" {
		t.Fatalf("expected non-empty output")
	}

	outputFile := filepath.Join(outputDir, "ocm@1.0.0.rb")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("expected output file to exist")
	}
}

func TestEnsureDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := ensureDirectory(dir); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	nonDirFile := filepath.Join(dir, "file")
	if err := os.WriteFile(nonDirFile, []byte("content"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	if err := ensureDirectory(nonDirFile); err == nil {
		t.Fatalf("expected error, got nil")
	}
}
