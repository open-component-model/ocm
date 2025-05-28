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
