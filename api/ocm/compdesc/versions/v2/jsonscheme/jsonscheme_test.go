package jsonscheme_test

import (
	"fmt"
	"testing"

	"ocm.software/ocm/api/ocm/compdesc/versions/v2/jsonscheme"
)

func v2Doc(name string) []byte {
	return []byte(fmt.Sprintf(`
meta:
  schemaVersion: v2
component:
  name: %s
  version: 1.0.0
  provider: acme.org
  repositoryContexts: []
  resources: []
  sources: []
  componentReferences: []
`, name))
}

// TestComponentNameGrammar pins the relaxed component-name grammar: the
// letter-first rule is dropped so digit-led and digit-only path segments are
// accepted, while every previously valid name keeps validating and uppercase
// stays rejected. See open-component-model/ocm-project#1422.
func TestComponentNameGrammar(t *testing.T) {
	cases := []struct {
		name  string
		valid bool
	}{
		// newly accepted (digit-led / digit-only) after dropping the letter-first rule
		{"1password.com/cli", true},
		{"acme.org/1", true},
		{"acme.org/1foo", true},
		{"acme.org/v1/2", true},
		// previously valid names keep validating (relaxation is a superset)
		{"acme.org/foo", true},
		{"github.com/open-component-model/ocm", true},
		{"acme.org/a_b", true},
		// still rejected
		{"acme.org/Foo", false},  // uppercase
		{"Acme.org/foo", false},  // uppercase label
		{"a", false},             // no DNS domain and no path
		{"acme/foo", false},      // no dot, no top-level label
		{"acme.12/foo", false},   // numeric top-level label
		{"acme.org/-foo", false}, // path segment starts with '-'
		{"a.com/_x", false},      // path segment starts with '_'
		{"-a.com/x", false},      // DNS label starts with '-'
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := jsonscheme.Validate(v2Doc(tc.name))
			if tc.valid && err != nil {
				t.Fatalf("expected %q to validate, got: %v", tc.name, err)
			}
			if !tc.valid && err == nil {
				t.Fatalf("expected %q to be rejected, but it validated", tc.name)
			}
		})
	}
}
