// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package labelmergehandler

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

type (
	LabelMergeHandler         = cpi.LabelMergeHandler
	LabelMergeHandlerConfig   = cpi.LabelMergeHandlerConfig
	LabelMergeHandlerRegistry = cpi.LabelMergeHandlerRegistry
)

func For(ctx cpi.ContextProvider) LabelMergeHandlerRegistry {
	return ctx.OCMContext().LabelMergeHandlers()
}
