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

package standard

import (
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
)

type Handler struct {
	opts *Options
}

func NewDefaultHandler(opts *Options) *Handler {
	if opts == nil {
		opts = &Options{}
	}
	return &Handler{opts: opts}
}

func New(opts ...transferhandler.TransferOption) (transferhandler.TransferHandler, error) {
	defaultOpts := &Options{}
	err := transferhandler.ApplyOptions(defaultOpts, opts...)
	if err != nil {
		return nil, err
	}
	return NewDefaultHandler(defaultOpts), nil
}

func (h *Handler) TransferVersion(repo ocm.Repository, src ocm.ComponentVersionAccess, meta *compdesc.ElementMeta) (ocm.Repository, transferhandler.TransferHandler, error) {
	if src == nil || h.opts.IsRecursive() {
		return repo, h, nil
	}
	return nil, nil, nil
}

func (h *Handler) TransferResource(src ocm.ComponentVersionAccess, a ocm.AccessSpec, r ocm.ResourceAccess) (bool, error) {
	return h.opts.IsResourcesByValue(), nil
}

func (h *Handler) TransferSource(src ocm.ComponentVersionAccess, a ocm.AccessSpec, r ocm.SourceAccess) (bool, error) {
	return h.opts.IsSourcesByValue(), nil
}

func (h *Handler) HandleTransferResource(r ocm.ResourceAccess, m ocm.AccessMethod, t ocm.ComponentVersionAccess) error {
	return t.SetResourceBlob(r.Meta(), accessio.BlobAccessForDataAccess("", -1, m.MimeType(), m), "", nil)
}

func (h *Handler) HandleTransferSource(r ocm.SourceAccess, m ocm.AccessMethod, t ocm.ComponentVersionAccess) error {
	return t.SetSourceBlob(r.Meta(), accessio.BlobAccessForDataAccess("", -1, m.MimeType(), m), "", nil)
}
