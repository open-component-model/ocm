package helm

import (
	"testing"
)

func TestAssureArchiveSuffix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"chart", "chart.tgz"},
		{"chart.tgz", "chart.tgz"},
		{"chart.tar.gz", "chart.tar.gz"},
		{"archive.zip", "archive.zip.tgz"},
	}

	for _, tt := range tests {
		result := AssureArchiveSuffix(tt.input)
		if result != tt.expected {
			t.Errorf("AssureArchiveSuffix(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestTrimExtN(t *testing.T) {
	tests := []struct {
		input    string
		n        int
		expected string
	}{
		{"chart.tgz", 1, "chart"},
		{"chart.tar.gz", 1, "chart.tar"},
		{"chart.tar.gz", 2, "chart"},
		{"archive.zip", 1, "archive"},
		{"archive", 1, "archive"},
	}

	for _, tt := range tests {
		result := trimExtN(tt.input, tt.n)
		if result != tt.expected {
			t.Errorf("trimExtN(%q, %d) = %q; want %q", tt.input, tt.n, result, tt.expected)
		}
	}
}
