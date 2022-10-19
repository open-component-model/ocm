// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/none"
	ocmcpi "github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type TransportClosure = common.NameVersionInfo

func TransferVersion(printer common.Printer, closure TransportClosure, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler transferhandler.TransferHandler) error {
	if closure == nil {
		closure = TransportClosure{}
	}
	if printer == nil {
		printer = common.NewPrinter(nil)
	}
	state := common.WalkingState{Closure: closure}
	return transferVersion(printer, state, src, tgt, handler)
}

func transferVersion(printer common.Printer, state common.WalkingState, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler transferhandler.TransferHandler) error {
	nv := common.VersionedElementKey(src)
	if ok, err := state.Add(ocm.KIND_COMPONENTVERSION, nv); !ok {
		return err
	}
	printer.Printf("transferring version %q...\n", nv)
	if handler == nil {
		var err error
		handler, err = standard.New(standard.Overwrite())
		if err != nil {
			return err
		}
	}

	d := src.GetDescriptor()

	comp, err := tgt.LookupComponent(src.GetName())
	if err != nil {
		return errors.Wrapf(err, "%s: lookup target component", state.History)
	}
	defer comp.Close()

	t, err := comp.LookupVersion(src.GetVersion())
	defer accessio.Close(t)
	if err != nil {
		if errors.IsErrNotFound(err) {
			t, err = comp.NewVersion(src.GetVersion())
			defer accessio.Close(t)
		}
	} else {
		var ok bool
		ok, err = handler.OverwriteVersion(src, t)
		if !ok {
			printer.Printf("  version %q already present -> skip transport\n", nv)
			return nil
		}
	}
	if err != nil {
		return errors.Wrapf(err, "%s: creating target version", state.History)
	}

	err = CopyVersion(printer, state.History, src, t, handler)
	if err != nil {
		return err
	}
	subp := printer.AddGap("  ")
	list := errors.ErrListf("component references for %s", nv)
	for _, r := range d.References {
		cv, shdlr, err := handler.TransferVersion(src.Repository(), src, &r)
		if err != nil {
			return errors.Wrapf(err, "%s: nested component %s[%s:%s]", state.History, r.GetName(), r.ComponentName, r.GetVersion())
		}
		if cv != nil {
			list.Add(transferVersion(subp, state, cv, tgt, shdlr))
			cv.Close()
		}
	}

	var unstr *runtime.UnstructuredTypedObject
	if !ocm.IsIntermediate(tgt.GetSpecification()) {
		unstr, err = runtime.ToUnstructuredTypedObject(tgt.GetSpecification())
		if err != nil {
			unstr = nil
		}
	}
	cd := t.GetDescriptor()
	if unstr != nil {
		cd.RepositoryContexts = append(cd.RepositoryContexts, unstr)
	}
	cd.Signatures = src.GetDescriptor().Signatures.Copy()
	printer.Printf("...adding component version...\n")
	return list.Add(comp.AddVersion(t)).Result()
}

func CopyVersion(printer common.Printer, hist common.History, src ocm.ComponentVersionAccess, t ocm.ComponentVersionAccess, handler transferhandler.TransferHandler) error {
	if handler == nil {
		handler = standard.NewDefaultHandler(nil)
	}

	*t.GetDescriptor() = *src.GetDescriptor().Copy()
	for i, r := range src.GetResources() {
		var m ocm.AccessMethod
		a, err := r.Access()
		if err == nil {
			m, err = r.AccessMethod()
			if err == nil {
				defer m.Close()
				ok := a.IsLocal(src.GetContext())
				if !ok {
					if a.GetKind() != none.Type {
						ok, err = handler.TransferResource(src, a, r)
					}
				}
				if ok {
					hint := ocmcpi.ArtefactNameHint(a, src)
					printArtefactInfo(printer, "resource", i, hint)
					err = handler.HandleTransferResource(r, m, hint, t)
				}
			}
		}
		if err != nil {
			if !errors.IsErrUnknownKind(err, errors.KIND_ACCESSMETHOD) {
				return errors.Wrapf(err, "%s: transferring resource %d", hist, i)
			}
			printer.Printf("WARN: %s: transferring resource %d: %s (enforce transport by reference)\n", hist, i, err)
		}
	}
	for i, r := range src.GetSources() {
		var m ocm.AccessMethod
		a, err := r.Access()
		if err == nil {
			m, err = r.AccessMethod()
			if err == nil {
				defer m.Close()
				ok := a.IsLocal(src.GetContext())
				if !ok {
					if a.GetKind() != none.Type {
						ok, err = handler.TransferSource(src, a, r)
					}
				}
				if ok {
					hint := ocmcpi.ArtefactNameHint(a, src)
					printArtefactInfo(printer, "source", i, hint)
					err = handler.HandleTransferSource(r, m, hint, t)
				}
			}
		}
		if err != nil {
			if !errors.IsErrUnknownKind(err, errors.KIND_ACCESSMETHOD) {
				return errors.Wrapf(err, "%s: transferring source %d", hist, i)
			}
			printer.Printf("WARN: %s: transferring source %d: %s (enforce transport by reference)\n", hist, i, err)
		}
	}
	return nil
}

func printArtefactInfo(printer common.Printer, kind string, index int, hint string) {
	if printer != nil {
		if hint != "" {
			printer.Printf("...resource %d(%s)...\n", index, hint)
		} else {
			printer.Printf("...resource %d...\n", index)
		}
	}
}
