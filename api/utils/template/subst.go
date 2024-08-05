package template

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/drone/envsubst"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

func init() {
	Register("subst", func(_ vfs.FileSystem) Templater { return NewSubst() }, `simple value substitution with the <code>drone/envsubst</code> templater.
It supports string values, only. Complex settings will be json encoded.
<pre>
  key:
    subkey: "abc ${MY_VAL}"
</pre>
`)
}

type Subst struct{}

var _ Templater = (*Subst)(nil)

func NewSubst() Templater {
	return &Subst{}
}

// Template templates a string with the parsed vars.
func (s *Subst) Process(data string, values Values) (string, error) {
	return envsubst.Eval(data, stringmapping(values))
}

// mapping is a helper function for the envsubst to provide the value for a variable name.
// It returns an empty string if the variable is not defined.
func stringmapping(values Values) func(variable string) string {
	return func(variable string) string {
		if values == nil {
			return ""
		}
		v := values[variable]
		if v == nil {
			return ""
		}
		t := reflect.TypeOf(v)
		if t.Kind() == reflect.Map || t.Kind() == reflect.Array {
			data, err := json.Marshal(v)
			if err != nil {
				return ""
			}
			return string(data)
		}
		return fmt.Sprintf("%v", v)
	}
}
