// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package vershdlr

import (
	"strings"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
)

func Compare(a, b interface{}) int {
	aa := a.(*Object)
	ab := b.(*Object)

	c := strings.Compare(aa.Component, ab.Component)
	if c != 0 {
		return c
	}
	return strings.Compare(aa.Version, ab.Version)
}

// Sort is a processing chain sorting original objects provided by type handler.
var Sort = processing.Sort(Compare)
