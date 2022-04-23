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

package ocm

import (
	"github.com/open-component-model/ocm/pkg/common"
	ocmcpi "github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
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

func TransferVersion(closure TransportClosure, repo ocmcpi.Repository, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler TransferHandler) error {
	if closure == nil {
		closure = TransportClosure{}
	}
	return transferVersion(nil, closure, repo, src, tgt, handler)
}

func transferVersion(hist common.History, closure TransportClosure, repo ocmcpi.Repository, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler TransferHandler) error {
	nv := common.NewNameVersion(src.GetName(), src.GetVersion())
	if err := hist.Add(KIND_COMPONENTVERSION, nv); err != nil {
		return err
	}
	if !closure.Add(nv) {
		return nil
	}

	if handler == nil {
		handler = NewDefaultTransferHandler(nil)
	}

	d := src.GetDescriptor()

	comp, err := tgt.LookupComponent(src.GetName())
	if err != nil {
		return errors.Wrapf(err, "%s: lookup target component", hist)
	}

	t, err := comp.NewVersion(src.GetVersion())
	if err != nil {
		return errors.Wrapf(err, "%s: creating target version", hist)
	}
	defer t.Close()
	err = CopyVersion(hist, src, t, handler)
	if err != nil {
		return err
	}
	for _, r := range d.ComponentReferences {
		if srepo, shdlr := handler.TransferVersion(repo, r.GetName(), r.GetVersion()); srepo != nil {
			c, err := srepo.LookupComponentVersion(r.GetName(), r.GetVersion())
			if err != nil {
				return errors.Wrapf(err, "%s: nested component %s:%s", hist, r.GetName(), r.GetVersion())
			}
			err = transferVersion(hist, closure, srepo, c, tgt, shdlr)
			if err != nil {
				return err
			}
		}
	}
	return comp.AddVersion(t)
}

func CopyVersion(hist common.History, src ComponentVersionAccess, t ComponentVersionAccess, handler TransferHandler) error {
	if handler == nil {
		handler = NewDefaultTransferHandler(nil)
	}

	*t.GetDescriptor() = *src.GetDescriptor()
	for i, r := range src.GetResources() {
		var m AccessMethod
		a, err := r.Access()
		if err == nil {
			m, err = r.AccessMethod()
			if err == nil && (a.IsLocal(src.GetContext()) || handler.TransferResource(src, a, r, t)) {
				err = handler.HandleTransferResource(r, m, t)
			}
		}
		if err != nil {
			return errors.Wrapf(err, "%s: transferring resource %d", hist, i)
		}
	}
	for i, r := range src.GetSources() {
		var m AccessMethod
		a, err := r.Access()
		if err == nil {
			m, err = r.AccessMethod()
			if err == nil && (a.IsLocal(src.GetContext()) || handler.TransferSource(src, a, r, t)) {
				err = handler.HandleTransferSource(r, m, t)
			}
		}
		if err != nil {
			return errors.Wrapf(err, "%s: transferring source %d", hist, i)
		}
	}
	return nil
}
