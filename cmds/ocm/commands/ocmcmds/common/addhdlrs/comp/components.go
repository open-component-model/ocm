package comp

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/goutils/sliceutils"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

func ProcessComponents(ctx clictx.Context, ictx inputs.Context, repo ocm.Repository, complete ocm.ComponentVersionResolver, thdlr transferhandler.TransferHandler, h *ResourceSpecHandler, elems []addhdlrs.Element) (err error) {
	list := errors.ErrorList{}

	for _, elem := range elems {
		if r, ok := elem.Spec().(*ResourceSpec); ok {
			if len(r.References) > 0 && len(r.OldReferences) > 0 {
				return fmt.Errorf("only field references or componentReferences (deprecated) is possible")
			}
			list.Add(addhdlrs.ValidateElementSpecIdentities("resource", elem.Source().String(), sliceutils.Convert[addhdlrs.ElementSpec](r.Resources)))
			list.Add(addhdlrs.ValidateElementSpecIdentities("source", elem.Source().String(), sliceutils.Convert[addhdlrs.ElementSpec](r.Sources)))
			list.Add(addhdlrs.ValidateElementSpecIdentities("reference", elem.Source().String(), sliceutils.Convert[addhdlrs.ElementSpec](r.References)))
			list.Add(addhdlrs.ValidateElementSpecIdentities("reference", elem.Source().String(), sliceutils.Convert[addhdlrs.ElementSpec](r.OldReferences)))
		}
	}
	if err := list.Result(); err != nil {
		return err
	}

	index := set.New[common.NameVersion]()
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
