package genericocireg_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg"
)

func TestToTag(t *testing.T) {
	tests := []struct {
		version   string
		expected  string
		expectErr bool
	}{
		{"1.0.0", "1.0.0", false},
		{"1.0.0+build", "1.0.0.build-", false},
		{"1.0.0+build.metadata", "1.0.0.build-metadata", false},
		{"0.0.1-20250108132333+af79499", "0.0.1-20250108132333+af79499", false},
		{"invalid_version", "", true},
	}

	for _, test := range tests {
		tag, err := genericocireg.ToTag(test.version)
		if test.expectErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, tag)
		}
	}
}

func TestToVersion(t *testing.T) {
	tests := []struct {
		tag      string
		expected string
	}{
		{"1.0.0", "1.0.0"},
		{"1.0.0.build-", "1.0.0+build"},
		{"1.0.0.build-metadata", "1.0.0+build.metadata"},
	}

	for _, test := range tests {
		version := genericocireg.ToVersion(test.tag)
		assert.Equal(t, test.expected, version)
	}
}
