// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package output

import (
	"strings"
)

func SelectBest(name string, candidates ...string) (string, int) {
	for i, c := range candidates {
		if strings.EqualFold(name, c) {
			return c, i
		}
	}
	return "", -1
}
