// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package valuemergehandler

import (
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/config"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/handlers"

	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/hpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/internal"
)

type (
	Context       = internal.Context
	Handler       = internal.Handler
	Handlers      = internal.Handlers
	Config        = internal.Config
	Registry      = internal.Registry
	Specification = internal.Specification
	Value         = internal.Value
)

const (
	KIND_VALUE_MERGE_ALGORITHM = hpi.KIND_VALUE_MERGE_ALGORITHM
	KIND_VALUESET              = "value set"
)

func NewSpecification(algo string, cfg Config) (*Specification, error) {
	return hpi.NewSpecification(algo, cfg)
}

func NewRegistry(base ...Registry) Registry {
	return internal.NewRegistry(base...)
}

func For(ctx cpi.ContextProvider) Registry {
	return hpi.For(ctx)
}

func SetFor(ctx datacontext.Context, registry Registry) {
	hpi.SetFor(ctx, registry)
}
