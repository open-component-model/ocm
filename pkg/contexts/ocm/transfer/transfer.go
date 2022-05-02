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
)

type TransportClosure map[common.NameVersion]struct{}

func (c TransportClosure) Add(nv common.NameVersion) bool {
	if _, ok := c[nv]; !ok {
		c[nv] = struct{}{}
		return true
	}
	return false
}

func (c TransportClosure) Contains(nv common.NameVersion) bool {
	_, ok := c[nv]
	return ok
}

func TransferVersion(printer Printer, closure TransportClosure, repo ocmcpi.Repository, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler transferhandler.TransferHandler) error {
	if closure == nil {
		closure = TransportClosure{}
	}
	if printer == nil {
		printer = NewPrinter(nil)
	}
	return transferVersion(printer, nil, closure, repo, src, tgt, handler)
}

func transferVersion(printer Printer, hist common.History, closure TransportClosure, repo ocmcpi.Repository, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler transferhandler.TransferHandler) error {
	nv := common.NewNameVersion(src.GetName(), src.GetVersion())
	if err := hist.Add(ocm.KIND_COMPONENTVERSION, nv); err != nil {
		return err
	}
	if !closure.Add(nv) {
		return nil
	}
	printer.Printf("transferring version %q...\n", common.VersionedElementKey(src))
	if handler == nil {
		handler = standard.NewDefaultHandler(nil)
	}

	d := src.GetDescriptor()

	comp, err := tgt.LookupComponent(src.GetName())
	if err != nil {
		return errors.Wrapf(err, "%s: lookup target component", hist)
	}

	t, err := comp.LookupVersion(src.GetVersion())
	if err != nil {
		if errors.IsErrNotFound(err) {
			t, err = comp.NewVersion(src.GetVersion())
		}
	}
	if err != nil {
		return errors.Wrapf(err, "%s: creating target version", hist)
	}
	defer t.Close()
	err = CopyVersion(hist, src, t, handler)
	if err != nil {
		return err
	}
	subp := printer.AddGap("  ")
	list := errors.ErrListf("component references for %s", nv)
	for _, r := range d.ComponentReferences {
		srepo, shdlr, err := handler.TransferVersion(repo, src, &r.ElementMeta)
		if err != nil {
			return err
		}
		if srepo != nil {
			c, err := srepo.LookupComponentVersion(r.GetName(), r.GetVersion())
			if err != nil {
				return errors.Wrapf(err, "%s: nested component %s:%s", hist, r.GetName(), r.GetVersion())
			}
			list.Add(transferVersion(subp, hist, closure, srepo, c, tgt, shdlr))
			if srepo != repo {
				srepo.Close()
			}
		}
	}
	return list.Add(comp.AddVersion(t)).Result()
}

func CopyVersion(hist common.History, src ocm.ComponentVersionAccess, t ocm.ComponentVersionAccess, handler transferhandler.TransferHandler) error {
	if handler == nil {
		handler = standard.NewDefaultHandler(nil)
	}

	*t.GetDescriptor() = *src.GetDescriptor()
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
					err = handler.HandleTransferResource(r, m, t)
				}
			}
		}
		if err != nil {
			return errors.Wrapf(err, "%s: transferring resource %d", hist, i)
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
					err = handler.HandleTransferSource(r, m, t)
				}
			}
		}
		if err != nil {
			return errors.Wrapf(err, "%s: transferring source %d", hist, i)
		}
	}
	return nil
}
