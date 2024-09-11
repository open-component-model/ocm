package common

import (
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/handlers/elemhdlr"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/utils"
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
	return elemhdlr.NewTypeHandler(octx, opts, repo, session, ocm.KIND_SOURCE, compspecs, func(access ocm.ComponentVersionAccess) compdesc.ElementListAccessor {
		return access.GetDescriptor().Sources
	}, hopts...)
}
