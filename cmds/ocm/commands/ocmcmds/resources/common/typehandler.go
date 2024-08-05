package common

import (
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/handlers/elemhdlr"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

func Elem(e interface{}) *compdesc.Resource {
	return e.(*elemhdlr.Object).Element.(*compdesc.Resource)
}

var (
	WithVersionConstraints = elemhdlr.WithVersionConstraints
	LatestOnly             = elemhdlr.LatestOnly
	OptionsFor             = elemhdlr.OptionsFor
)

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
