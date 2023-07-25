// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/common/handlers/elemhdlr"
	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/v2/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/compdesc"
)

func Elem(e interface{}) *compdesc.Source {
	return e.(*elemhdlr.Object).Element.(*compdesc.Source)
}

var OptionsFor = elemhdlr.OptionsFor

////////////////////////////////////////////////////////////////////////////////

type TypeHandler struct {
	*elemhdlr.TypeHandler
}

func NewTypeHandler(octx clictx.OCM, opts *output.Options, repo ocm.Repository, session ocm.Session, compspecs []string, hopts ...elemhdlr.Option) (utils.TypeHandler, error) {
	return elemhdlr.NewTypeHandler(octx, opts, repo, session, ocm.KIND_SOURCE, compspecs, func(access ocm.ComponentVersionAccess) compdesc.ElementAccessor {
		return access.GetDescriptor().Sources
	}, hopts...)
}
