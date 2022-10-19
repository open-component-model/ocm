// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

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

func (h *Handler) OverwriteVersion(src ocm.ComponentVersionAccess, tgt ocm.ComponentVersionAccess) (bool, error) {
	return h.opts.IsOverwrite(), nil
}

func (h *Handler) TransferVersion(repo ocm.Repository, src ocm.ComponentVersionAccess, meta *compdesc.ComponentReference) (ocm.ComponentVersionAccess, transferhandler.TransferHandler, error) {
	if src == nil || h.opts.IsRecursive() {
		compoundResolver := ocm.NewCompoundResolver(repo, h.opts.GetResolver())
		cv, err := compoundResolver.LookupComponentVersion(meta.GetComponentName(), meta.Version)
		return cv, h, err
	}
	return nil, nil, nil
}

func (h *Handler) TransferResource(src ocm.ComponentVersionAccess, a ocm.AccessSpec, r ocm.ResourceAccess) (bool, error) {
	return h.opts.IsResourcesByValue(), nil
}

func (h *Handler) TransferSource(src ocm.ComponentVersionAccess, a ocm.AccessSpec, r ocm.SourceAccess) (bool, error) {
	return h.opts.IsSourcesByValue(), nil
}

func (h *Handler) HandleTransferResource(r ocm.ResourceAccess, m ocm.AccessMethod, hint string, t ocm.ComponentVersionAccess) error {
	return t.SetResourceBlob(r.Meta(), accessio.BlobAccessForDataAccess("", -1, m.MimeType(), m), hint, nil)
}

func (h *Handler) HandleTransferSource(r ocm.SourceAccess, m ocm.AccessMethod, hint string, t ocm.ComponentVersionAccess) error {
	return t.SetSourceBlob(r.Meta(), accessio.BlobAccessForDataAccess("", -1, m.MimeType(), m), hint, nil)
}
