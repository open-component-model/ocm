// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

// Package hpi contains the Handler Programming Interface for
// value merge handlers
package hpi

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/internal"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/utils"
)

// resolve package cycle among default merge handler and
// labelmergehandler by separating commonly used types
// into this package

// same problem for the embedding into the OCM environment
// required for the ocm.Context access.
// Because of this cycle, the registry implementation and the
// required types have to be placed into the internal package of
// ocm and forwarded to the cpi packages. From there it can be consumed
// here to break the dependency cycle.

type (
	Context       = internal.Context
	Handler       = internal.Handler
	Config        = internal.Config
	Registry      = internal.Registry
	Specification = internal.Specification
	Value         = internal.Value
)

func Register(h Handler) {
	internal.Register(h)
}

func Assign(hint string, spec *Specification) {
	internal.Assign(hint, spec)
}

func NewSpecification(algo string, cfg ...Config) (*Specification, error) {
	raw, err := runtime.AsRawMessage(utils.Optional(cfg...))
	if err != nil {
		return nil, err
	}
	return &Specification{
		Algorithm: algo,
		Config:    raw,
	}, nil
}

func NewRegistry(base ...Registry) Registry {
	return internal.NewRegistry(base...)
}
