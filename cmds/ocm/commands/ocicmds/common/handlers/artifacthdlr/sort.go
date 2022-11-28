// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package artifacthdlr

import (
	"strings"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/pkg/common"
)

func Compare(a, b interface{}) int {
	aa := a.(*Object)
	ab := b.(*Object)

	c := strings.Compare(aa.Spec.UniformRepositorySpec.String(), ab.Spec.UniformRepositorySpec.String())
	if c == 0 {
		return common.CompareHistoryElement(aa, ab)
	}
	return c
}

// Sort is a processing chain sorting original objects provided by type handler.
var Sort = processing.Sort(Compare)
