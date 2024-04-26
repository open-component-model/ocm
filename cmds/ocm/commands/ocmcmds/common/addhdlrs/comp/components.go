package comp

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/generics"
	"github.com/open-component-model/ocm/pkg/out"
)

func ProcessComponents(ctx clictx.Context, ictx inputs.Context, repo ocm.Repository, complete ocm.ComponentVersionResolver, thdlr transferhandler.TransferHandler, h *ResourceSpecHandler, elems []addhdlrs.Element) (err error) {
	list := errors.ErrorList{}

	for _, elem := range elems {
		if r, ok := elem.Spec().(*ResourceSpec); ok {
			list.Add(addhdlrs.ValidateElementSpecIdentities("resource", elem.Source().String(), generics.ConvertSliceTo[addhdlrs.ElementSpec](r.Resources)))
			list.Add(addhdlrs.ValidateElementSpecIdentities("source", elem.Source().String(), generics.ConvertSliceTo[addhdlrs.ElementSpec](r.Sources)))
			list.Add(addhdlrs.ValidateElementSpecIdentities("reference", elem.Source().String(), generics.ConvertSliceTo[addhdlrs.ElementSpec](r.References)))
		}
	}
	if err := list.Result(); err != nil {
		return err
	}

	index := generics.Set[common.NameVersion]{}
	for _, elem := range elems {
		if r, ok := elem.Spec().(*ResourceSpec); ok {
			index.Add(common.NewNameVersion(r.Name, r.Version))
		}
	}

	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&err)

	for _, elem := range elems {
		loop := finalize.Nested()
		err := h.Add(ctx, ictx.Section("adding %s...", elem.Spec().Info()), elem, repo)
		if err != nil {
			return errors.Wrapf(err, "failed adding component %q(%s)", elem.Spec().GetName(), elem.Source())
		}

		if r, ok := elem.Spec().(*ResourceSpec); complete != nil && thdlr != nil && ok {
			cv, err := repo.LookupComponentVersion(r.Name, r.Version)
			if err != nil {
				return errors.Wrapf(err, "accessing added component version failed")
			}
			loop.Close(cv)
			if len(cv.GetDescriptor().References) > 0 {
				ictx.Printf("completing %s:%s...\n", r.Name, r.Version)
				for _, cr := range cv.GetDescriptor().References {
					loop := loop.Nested()
					nv := common.NewNameVersion(cr.ComponentName, cr.Version)
					if index.Contains(nv) {
						continue
					}
					found, err := repo.LookupComponentVersion(nv.GetName(), nv.GetVersion())
					if err == nil && found != nil {
						found.Close()
						out.Outf(ctx, "  reference %s[%s] already found\n", cr.Name, nv)
						continue
					}
					found, err = complete.LookupComponentVersion(nv.GetName(), nv.GetVersion())
					if err != nil || found == nil {
						return errors.NewEf(err, "referenced component version %s not found", nv)
					}
					loop.Close(found)
					err = transfer.TransferVersion(ictx.Printer().AddGap("  "), nil, found, repo, thdlr)
					if err != nil {
						return errors.Wrapf(err, "completing reference %s[%s] of %s:%s failed", cr.Name, nv, r.Name, r.Version)
					}
					err = loop.Finalize()
					if err != nil {
						return err
					}
				}
			}
			err = loop.Finalize()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
