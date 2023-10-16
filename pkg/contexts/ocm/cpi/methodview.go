// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"io"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessio/refmgmt"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/internal"
)

// AccessMethodView can be used map wrap an access method
// into a managed method with multiple views. The original method
// object is closed once the last view is closed.
type AccessMethodView interface {
	AccessMethod

	Base() interface{}
	Dup() (AccessMethodView, error)
}

// AccessMethodAsView wrap an access method object into
// a multi-view version. The original method is closed when
// the last view is closed.
// After an access method is used as base object, it should not
// explicitly closed anymore, because the views will stop
// functioning.
func AccessMethodAsView(acc AccessMethod, closer ...io.Closer) AccessMethodView {
	return refmgmt.WithView[AccessMethod, AccessMethodView](acc, accessMethodViewCreator, closer...)
}

func AccessMethodViewForSpec(spec AccessSpec, cv ComponentVersionAccess) (AccessMethodView, error) {
	m, err := spec.AccessMethod(cv)
	if err != nil {
		return nil, err
	}
	return AccessMethodAsView(m), nil
}

func AccessMethodViewForAccessProvider(p AccessProvider) (AccessMethodView, error) {
	m, err := p.AccessMethod()
	if err != nil {
		return nil, err
	}
	return AccessMethodAsView(m), nil
}

func accessMethodViewCreator(acc AccessMethod, view *refmgmt.View[AccessMethodView]) AccessMethodView {
	return &accessMethodView{view, acc}
}

type accessMethodView struct {
	*refmgmt.View[AccessMethodView]
	access AccessMethod
}

func (a *accessMethodView) Base() interface{} {
	return a.access
}

func (a *accessMethodView) Get() ([]byte, error) {
	var result []byte
	err := a.Execute(func() (err error) {
		result, err = a.access.Get()
		return
	})
	return result, err
}

func (a *accessMethodView) Reader() (io.ReadCloser, error) {
	var result io.ReadCloser
	err := a.Execute(func() (err error) {
		result, err = a.access.Reader()
		return
	})
	return result, err
}

func (a *accessMethodView) GetKind() string {
	return a.access.GetKind()
}

func (a accessMethodView) AccessSpec() internal.AccessSpec {
	return a.access.AccessSpec()
}

func (a accessMethodView) MimeType() string {
	return a.access.MimeType()
}

////////////////////////////////////////////////////////////////////////////////

func BlobAccessForAccessMethod(m AccessMethodView) (accessio.AnnotatedBlobAccess[AccessMethodView], error) {
	m, err := m.Dup()
	if err != nil {
		return nil, err
	}
	return accessio.BlobAccessForDataAccess("", -1, m.MimeType(), m), nil
}
