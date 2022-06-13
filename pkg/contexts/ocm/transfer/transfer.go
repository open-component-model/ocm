// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package transfer

import (
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	ocmcpi "github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type TransportClosure = common.NameVersionInfo

func TransferVersion(printer common.Printer, closure TransportClosure, repo ocmcpi.Repository, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler transferhandler.TransferHandler) error {
	if closure == nil {
		closure = TransportClosure{}
	}
	if printer == nil {
		printer = common.NewPrinter(nil)
	}
	state := common.WalkingState{Closure: closure}
	return transferVersion(printer, state, repo, src, tgt, handler)
}

func transferVersion(printer common.Printer, state common.WalkingState, repo ocmcpi.Repository, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler transferhandler.TransferHandler) error {
	nv := common.VersionedElementKey(src)
	if ok, err := state.Add(ocm.KIND_COMPONENTVERSION, nv); !ok {
		return err
	}
	printer.Printf("transferring version %q...\n", nv)
	if handler == nil {
		handler = standard.NewDefaultHandler(nil)
	}

	d := src.GetDescriptor()

	comp, err := tgt.LookupComponent(src.GetName())
	if err != nil {
		return errors.Wrapf(err, "%s: lookup target component", state.History)
	}

	t, err := comp.LookupVersion(src.GetVersion())
	if err != nil {
		if errors.IsErrNotFound(err) {
			t, err = comp.NewVersion(src.GetVersion())
		}
	}
	if err != nil {
		return errors.Wrapf(err, "%s: creating target version", state.History)
	}
	defer t.Close()

	err = CopyVersion(printer, state.History, src, t, handler)
	if err != nil {
		return err
	}
	subp := printer.AddGap("  ")
	list := errors.ErrListf("component references for %s", nv)
	for _, r := range d.References {
		srepo, shdlr, err := handler.TransferVersion(repo, src, &r.ElementMeta)
		if err != nil {
			return err
		}
		if srepo != nil {
			c, err := srepo.LookupComponentVersion(r.ComponentName, r.GetVersion())
			if err != nil {
				return errors.Wrapf(err, "%s: nested component %s[%s:%s]", state.History, r.GetName(), r.ComponentName, r.GetVersion())
			}
			list.Add(transferVersion(subp, state, srepo, c, tgt, shdlr))
			if srepo != repo {
				srepo.Close()
			}
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
					ok, err = handler.TransferResource(src, a, r)
				}
				if ok {
					hint := ""
					if h, ok := a.(ocm.HintProvider); ok {
						hint = h.GetReferenceName()
					}
					if printer != nil {
						printer.Printf("...resource %d...\n", i)
					}
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
					ok, err = handler.TransferSource(src, a, r)
				}
				if ok {
					hint := ""
					if h, ok := a.(ocm.HintProvider); ok {
						hint = h.GetReferenceName()
					}
					if printer != nil {
						printer.Printf("...source %d...\n", i)
					}
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
