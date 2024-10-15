package template

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/spiff/features"
	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

const SPIFF_MODE = "mode"

func init() {
	Register("spiff", NewSpiff, `[spiff templating](https://github.com/mandelsoft/spiff).
It supports complex values. the settings are accessible using the binding <code>values</code>.
<pre>
  key:
    subkey: "abc (( values.MY_VAL ))"
</pre>
`)
}

type Spiff struct {
	spiff spiffing.Spiff
}

func NewSpiff(fs vfs.FileSystem, opts TemplaterOptions) Templater {
	s := spiffing.New().WithFileSystem(fs).WithFeatures(features.CONTROL, features.INTERPOLATION)

	m := opts.Get(SPIFF_MODE)

	if mode, ok := m.(int); ok {
		s = s.WithMode(mode)
	}
	return &Spiff{
		s,
	}
}

func (s *Spiff) Process(data string, values Values) (string, error) {
	spiff, err := s.spiff.WithValues(map[string]interface{}{"values": values})
	if err != nil {
		return "", err
	}
	docs, err := SplitYamlDocuments([]byte(data))
	if err != nil {
		return "", err
	}

	result := ""
	for i, d := range docs {
		tmp, err := spiffing.Process(spiff, spiffing.NewSourceData(fmt.Sprintf("spec document %d", i), d))
		if err != nil {
			return "", err
		}
		if result != "" {
			if !strings.HasSuffix(result, "\n") {
				result += "\n"
			}
			result += "---\n"
		}
		result += string(tmp)
	}
	return result, nil
}
