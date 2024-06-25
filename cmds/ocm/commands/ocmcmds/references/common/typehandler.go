package common

import (
	"github.com/open-component-model/ocm/api/clictx"
	"github.com/open-component-model/ocm/api/ocm"
	"github.com/open-component-model/ocm/api/ocm/compdesc"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/elemhdlr"
	"github.com/open-component-model/ocm/cmds/ocm/common/output"
	"github.com/open-component-model/ocm/cmds/ocm/common/utils"
)

func Elem(e interface{}) *compdesc.ComponentReference {
	return e.(*elemhdlr.Object).Element.(*compdesc.ComponentReference)
}

var OptionsFor = elemhdlr.OptionsFor

////////////////////////////////////////////////////////////////////////////////

type TypeHandler struct {
	*elemhdlr.TypeHandler
}

func NewTypeHandler(octx clictx.OCM, opts *output.Options, repo ocm.Repository, session ocm.Session, compspecs []string, hopts ...elemhdlr.Option) (utils.TypeHandler, error) {
	return elemhdlr.NewTypeHandler(octx, opts, repo, session, ocm.KIND_REFERENCE, compspecs, func(access ocm.ComponentVersionAccess) compdesc.ElementAccessor {
		return access.GetDescriptor().References
	}, hopts...)
}
