package comphdlr

import (
	"github.com/mandelsoft/goutils/errors"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/cmds/ocm/commands/common/options/closureoption"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

func Evaluate(octx clictx.OCM, session ocm.Session, repobase ocm.Repository, compspecs []string, oopts *output.Options, opts ...Option) (Objects, error) {
	h := NewTypeHandler(octx, session, repobase, opts...)

	oopts = oopts.WithSession(session)
	comps := output.NewElementOutput(oopts, closureoption.Closure(oopts, ClosureExplode, Sort))
	err := utils.HandleOutput(comps, h, utils.StringElemSpecs(compspecs...)...)
	if err != nil {
		return nil, err
	}
	components := Objects{}
	i := comps.Elems.Iterator()
	for i.HasNext() {
		components = append(components, i.Next().(*Object))
	}
	if len(components) == 0 {
		if len(compspecs) == 0 {
			return nil, errors.Newf("no component version specified")
		}
		return nil, errors.Newf("no component version found")
	}
	return components, nil
}
