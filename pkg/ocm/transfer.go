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
	"fmt"

	"github.com/gardener/ocm/pkg/common"
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/errors"
	ocmcpi "github.com/gardener/ocm/pkg/ocm/cpi"
)

type TransferHandler interface {
	TransferVersion(repo Repository, name, version string) (Repository, TransferHandler)
	TransferResource(src ResourceAccess, tgt ComponentVersionAccess) error
	TransferSource(src SourceAccess, tgt ComponentVersionAccess) error
}

type DefaultTransferHandler struct {
	recursive bool
}

func NewDefaultTransferHandler(recursive bool) TransferHandler {
	return DefaultTransferHandler{recursive}
}

func (h DefaultTransferHandler) TransferVersion(repo Repository, name, version string) (Repository, TransferHandler) {
	if h.recursive {
		return repo, h
	}
	return nil, nil
}

func (DefaultTransferHandler) TransferResource(r ResourceAccess, t ComponentVersionAccess) error {
	a, err := r.Access()
	if err != nil {
		return err
	}
	if a.IsLocal(t.GetContext()) {
		m, err := r.AccessMethod()
		if err != nil {
			return err
		}
		return t.SetResourceBlob(r.Meta(), accessio.BlobAccessForDataAccess("", -1, m.MimeType(), m), "", nil)
	}
	return nil
}

func (DefaultTransferHandler) TransferSource(r SourceAccess, t ComponentVersionAccess) error {
	a, err := r.Access()
	if err != nil {
		return err
	}
	if a.IsLocal(t.GetContext()) {
		m, err := r.AccessMethod()
		if err != nil {
			return err
		}
		return t.SetSourceBlob(r.Meta(), accessio.BlobAccessForDataAccess("", -1, m.MimeType(), m), "", nil)
	}
	return nil
}

type history []common.NameVersion

func (h history) String() string {
	s := ""
	sep := ""
	for _, e := range h {
		s = fmt.Sprintf("%s%s%s", s, sep, e)
		sep = "->"
	}
	return s
}

func (h history) Contains(nv common.NameVersion) bool {
	for _, e := range h {
		if e == nv {
			return true
		}
	}
	return false
}

func TransferVersion(repo ocmcpi.Repository, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler TransferHandler) error {
	return transferVersion(nil, repo, src, tgt, handler)
}

func transferVersion(hist history, repo ocmcpi.Repository, src ocmcpi.ComponentVersionAccess, tgt ocmcpi.Repository, handler TransferHandler) error {
	nv := common.NewNameVersion(src.GetName(), src.GetVersion())
	if hist.Contains(nv) {
		return errors.Newf("%s: reference recursion for %s", hist, nv)
	}
	hist = append(hist, nv)

	if handler == nil {
		handler = DefaultTransferHandler{}
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
	*t.GetDescriptor() = *d
	defer t.Close()
	for i, r := range src.GetResources() {
		err = handler.TransferResource(r, t)
		if err != nil {
			return errors.Wrapf(err, "%s: transferring resource %d", hist, i)
		}
	}
	for i, r := range src.GetSources() {
		err = handler.TransferSource(r, t)
		if err != nil {
			return errors.Wrapf(err, "%s: transferring source %d", hist, i)
		}
	}
	for _, r := range d.ComponentReferences {
		if srepo, shdlr := handler.TransferVersion(repo, r.GetName(), r.GetVersion()); srepo != nil {
			c, err := srepo.LookupComponentVersion(r.GetName(), r.GetVersion())
			if err != nil {
				return errors.Wrapf(err, "%s: nested component %s:%s", hist, r.GetName(), r.GetVersion())
			}
			err = transferVersion(hist, srepo, c, tgt, shdlr)
			if err != nil {
				return err
			}
		}
	}
	return comp.AddVersion(t)
}
