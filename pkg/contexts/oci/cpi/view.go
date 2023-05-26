// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"io"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/internal"
)

type (
	NamespaceImpl interface {
		internal.NamespaceInt
		io.Closer
	}
	_NamespaceImpl = NamespaceImpl
)

type namespaceInt struct {
	refs accessio.ReferencableCloser
	impl NamespaceImpl
}

func NewNamespaceAccess(ns NamespaceImpl) (NamespaceAccess, error) {
	i := &namespaceInt{
		refs: accessio.NewRefCloser(ns, true),
		impl: ns,
	}
	return i.Dup()
}

func (i *namespaceInt) Dup() (NamespaceAccess, error) {
	v, err := i.refs.View()
	if err != nil {
		return nil, err
	}
	return &namespaceView{
		view:           v,
		_NamespaceImpl: i.impl,
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

type namespaceView struct {
	view accessio.CloserView
	_NamespaceImpl
}

func (v *namespaceView) IsClosed() bool {
	return v.view.IsClosed()
}

func (v *namespaceView) Close() error {
	return v.view.Close()
}

func (v *namespaceView) Dup() (NamespaceAccess, error) {
	n, err := v.view.View()
	if err != nil {
		return nil, err
	}
	return &namespaceView{
		view:           n,
		_NamespaceImpl: v._NamespaceImpl,
	}, nil
}
