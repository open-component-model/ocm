// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
)

type KeyInfo interface {
	GetKey() string
	GetDescription() string
}

func FormatList(def string, elems ...KeyInfo) string {
	names := ""
	for _, n := range elems {
		add := ""
		if n.GetKey() == def {
			add = " (default)"
		}
		names = fmt.Sprintf("%s\n  - <code>%s</code>:%s %s", names, n.GetKey(), add, n.GetDescription())
	}
	return names
}
