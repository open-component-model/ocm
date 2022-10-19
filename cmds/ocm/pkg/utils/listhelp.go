// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/utils"
)

type StringElementList []string

func (l StringElementList) Size() int                { return len(l) }
func (l StringElementList) Key(i int) string         { return l[i] }
func (l StringElementList) Description(i int) string { return "" }

func FormatList(def string, elems ...string) string {
	return FormatListElements(def, StringElementList(elems))
}

type ListElements interface {
	Size() int
	Key(i int) string
	Description(i int) string
}

func FormatListElements(def string, elems ListElements) string {
	names := ""
	size := elems.Size()

	for i := 0; i < size; i++ {
		key := elems.Key(i)
		names = fmt.Sprintf("%s\n  - <code>%s</code>", names, key)
		if key == def {
			names += " (default)"
		}
		desc := elems.Description(i)
		names += ": " + utils.IndentLines(desc, "    ", true)
	}
	return names + "\n"
}

type IdentityMatcherList []credentials.IdentityMatcherInfo

func (l IdentityMatcherList) Size() int                { return len(l) }
func (l IdentityMatcherList) Key(i int) string         { return l[i].Type }
func (l IdentityMatcherList) Description(i int) string { return l[i].Description }
