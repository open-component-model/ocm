// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/spiff/features"
	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

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

func NewSpiff(fs vfs.FileSystem) Templater {
	return &Spiff{
		spiffing.New().WithFileSystem(fs).WithFeatures(features.CONTROL, features.INTERPOLATION),
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
