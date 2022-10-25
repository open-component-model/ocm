// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"github.com/mandelsoft/vfs/pkg/vfs"
)

func init() {
	Register("none", func(_ vfs.FileSystem) Templater { return NewSubst() }, `do not do any substitution.
`)
}

type None struct{}

var _ Templater = (*None)(nil)

func NewNone() Templater {
	return &None{}
}

// Template templates a string with the parsed vars.
func (s *None) Process(data string, values Values) (string, error) {
	return data, nil
}
