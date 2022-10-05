// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

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
