// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/elemhdlr"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
)

func Elem(e interface{}) *compdesc.Resource {
	return e.(*elemhdlr.Object).Element.(*compdesc.Resource)
}

var WithVersionConstraints = elemhdlr.WithVersionConstraints
var LatestOnly = elemhdlr.LatestOnly
var OptionsFor = elemhdlr.OptionsFor

type typeFilter struct {
	types []string
}

func (t typeFilter) ApplyToElemHandler(handler *elemhdlr.TypeHandler) {
	if len(t.types) > 0 {
		handler.SetFilter(t)
	}
}

func (t typeFilter) Accept(e compdesc.ElementMetaAccessor) bool {
	if len(t.types) == 0 {
		return true
	}
	typ := e.(*compdesc.Resource).GetType()
	for _, a := range t.types {
		if a == typ {
			return true
		}
	}
	return false
}

func WithTypes(types []string) elemhdlr.Option {
	return typeFilter{types}
}

////////////////////////////////////////////////////////////////////////////////

func NewTypeHandler(octx clictx.OCM, opts *output.Options, repo ocm.Repository, session ocm.Session, compspecs []string, hopts ...elemhdlr.Option) (utils.TypeHandler, error) {
	return elemhdlr.NewTypeHandler(octx, opts, repo, session, ocm.KIND_RESOURCE, compspecs, func(access ocm.ComponentVersionAccess) compdesc.ElementAccessor {
		return access.GetDescriptor().Resources
	}, hopts...)
}
