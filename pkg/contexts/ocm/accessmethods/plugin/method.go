// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type AccessSpec struct {
	runtime.UnstructuredVersionedTypedObject `json:",inline"`
	handler                                  *PluginHandler
}

var (
	_ cpi.AccessSpec   = &AccessSpec{}
	_ cpi.HintProvider = &AccessSpec{}
)

func (s *AccessSpec) AccessMethod(cv cpi.ComponentVersionAccess) (cpi.AccessMethod, error) {
	return s.handler.AccessMethod(s, cv)
}

func (s *AccessSpec) Describe(ctx cpi.Context) string {
	return s.handler.Describe(s, ctx)
}

func (_ *AccessSpec) IsLocal(cpi.Context) bool {
	return false
}

func (s *AccessSpec) GlobalAccessSpec(cpi.Context) cpi.AccessSpec {
	return s
}

func (s *AccessSpec) GetMimeType() string {
	return s.handler.GetMimeType(s)
}

func (s *AccessSpec) GetReferenceHint(cv cpi.ComponentVersionAccess) string {
	return s.handler.GetReferenceHint(s, cv)
}

func (s *AccessSpec) Handler() *PluginHandler {
	return s.handler
}

////////////////////////////////////////////////////////////////////////////////

type accessMethod struct {
	lock sync.Mutex
	blob accessio.BlobAccess
	ctx  ocm.Context

	handler *PluginHandler
	spec    *AccessSpec
	info    *ppi.AccessSpecInfo
	creds   json.RawMessage
}

var _ cpi.AccessMethod = (*accessMethod)(nil)

func newMethod(p *PluginHandler, spec *AccessSpec, ctx ocm.Context, info *ppi.AccessSpecInfo, creds json.RawMessage) *accessMethod {
	return &accessMethod{
		ctx:     ctx,
		handler: p,
		spec:    spec,
		info:    info,
		creds:   creds,
	}
}

func (m *accessMethod) GetKind() string {
	return m.spec.GetKind()
}

func (m *accessMethod) AccessSpec() cpi.AccessSpec {
	return m.spec
}

func (m *accessMethod) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.blob != nil {
		m.blob.Close()
		m.blob = nil
	}
	return nil
}

func (m *accessMethod) Get() ([]byte, error) {
	return accessio.BlobData(m.getBlob())
}

func (m *accessMethod) Reader() (io.ReadCloser, error) {
	return accessio.BlobReader(m.getBlob())
}

func (m *accessMethod) MimeType() string {
	return m.info.MediaType
}

func (m *accessMethod) getBlob() (cpi.BlobAccess, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.blob != nil {
		return m.blob, nil
	}

	spec, err := json.Marshal(m.spec)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot marshal access spec")
	}
	m.blob = accessobj.CachedBlobAccessForWriter(m.ctx, m.MimeType(), plugin.NewAccessDataWriter(m.handler.plug, m.creds, spec))
	return m.blob, nil
}
