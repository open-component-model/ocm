package substutils

import (
	"encoding/json"
	"os"
	"strings"
)

type SubstitutionMap map[string]string

func (s SubstitutionMap) Substitute(variable string) (string, bool) {
	r, ok := s[variable]
	return r, ok
}

// MergeMapSubstitution merges SubstitutionMap objects.
// Hereby, later definitions will override previous ones.
func MergeMapSubstitution(subst ...SubstitutionMap) SubstitutionMap {
	r := SubstitutionMap{}
	for _, s := range subst {
		for k, v := range s {
			r[k] = v
		}
	}
	return r
}

// SubstList provides a SubstitutionMap for a list of
// key/values pairs.
func SubstList(values ...string) SubstitutionMap {
	r := SubstitutionMap{}
	for i := 0; i+1 < len(values); i += 2 {
		r[values[i]] = values[i+1]
	}
	return r
}

// SubstFrom provides a SubstitutionMap
// for the serializable attributes of an object
// with string attributes.
func SubstFrom(v interface{}, prefix ...string) SubstitutionMap {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	var values map[string]string
	err = json.Unmarshal(data, &values)
	if err != nil {
		panic(err)
	}
	if len(prefix) > 0 {
		p := strings.Join(prefix, "")
		n := map[string]string{}
		for k, v := range values {
			n[p+k] = v
		}
		values = n
	}
	return values
}

////////////////////////////////////////////////////////////////////////////////

func EnvSubstitution(prefixes ...string) Substitution {
	prefix := strings.Join(prefixes, "")
	return SubstitutionFunc(func(variable string) (string, bool) {
		if len(prefix) > 0 && !strings.HasPrefix(variable, prefix) {
			return "", false
		}
		variable = variable[len(prefix):]
		for _, s := range os.Environ() {
			i := strings.Index(s, "=")
			if i > 0 && s[:i] == variable {
				return s[i+1:], true
			}
		}
		return "", false
	})
}
