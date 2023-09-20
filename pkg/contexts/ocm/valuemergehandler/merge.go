// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package valuemergehandler

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/hpi"
)

func Merge(ctx cpi.Context, m *Specification, hint string, local Value, inbound *Value) (bool, error) {
	return hpi.Merge(ctx, m, hint, local, inbound)
}

func LabelHint(name string, optversion ...string) string {
	return hpi.LabelHint(name, optversion...)
}
