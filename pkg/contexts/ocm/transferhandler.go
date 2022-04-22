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
	"github.com/open-component-model/ocm/pkg/common/accessio"
)

type TransferHandler interface {
	TransferVersion(repo Repository, name, version string) (Repository, TransferHandler)
	TransferResource(src ResourceAccess, tgt ComponentVersionAccess) error
	TransferSource(src SourceAccess, tgt ComponentVersionAccess) error
}

type DefaultTransferHandler struct {
	opts *DefaultTransferOptions
}

func NewDefaultTransferHandler(opts *DefaultTransferOptions) TransferHandler {
	if opts == nil {
		opts = &DefaultTransferOptions{}
	}
	return &DefaultTransferHandler{opts: opts}
}

func NewTransferHandler(opts ...TransferOption) TransferHandler {
	defaultOpts := DefaultTransferOptions{}
	for _, opt := range opts {
		opt.Apply(defaultOpts)
	}
	return NewDefaultTransferHandler(&defaultOpts)
}

func (h *DefaultTransferHandler) TransferVersion(repo Repository, name, version string) (Repository, TransferHandler) {
	if h.opts.IsRecursive() {
		return repo, h
	}
	return nil, nil
}

func (h *DefaultTransferHandler) TransferResource(r ResourceAccess, t ComponentVersionAccess) error {
	a, err := r.Access()
	if err != nil {
		return err
	}
	if a.IsLocal(t.GetContext()) || h.opts.IsResourcesByValue() {
		m, err := r.AccessMethod()
		if err != nil {
			return err
		}
		return t.SetResourceBlob(r.Meta(), accessio.BlobAccessForDataAccess("", -1, m.MimeType(), m), "", nil)
	}
	return nil
}

func (h *DefaultTransferHandler) TransferSource(r SourceAccess, t ComponentVersionAccess) error {
	a, err := r.Access()
	if err != nil {
		return err
	}
	if a.IsLocal(t.GetContext()) || h.opts.IsSourcesByValue() {
		m, err := r.AccessMethod()
		if err != nil {
			return err
		}
		return t.SetSourceBlob(r.Meta(), accessio.BlobAccessForDataAccess("", -1, m.MimeType(), m), "", nil)
	}
	return nil
}
