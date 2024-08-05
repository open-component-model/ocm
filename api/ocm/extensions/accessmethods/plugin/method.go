package plugin

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	"ocm.software/ocm/api/ocm"
	cpi "ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/plugin"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/runtime"
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
	blob blobaccess.BlobAccess
	ctx  ocm.Context

	handler *PluginHandler
	spec    *AccessSpec
	info    *ppi.AccessSpecInfo
	creds   json.RawMessage
}

var _ cpi.AccessMethodImpl = (*accessMethod)(nil)

func newMethod(p *PluginHandler, spec *AccessSpec, ctx ocm.Context, info *ppi.AccessSpecInfo, creds json.RawMessage) *accessMethod {
	return &accessMethod{
		ctx:     ctx,
		handler: p,
		spec:    spec,
		info:    info,
		creds:   creds,
	}
}

func (_ *accessMethod) IsLocal() bool {
	return false
}

func (m *accessMethod) GetKind() string {
	return m.spec.GetKind()
}

func (m *accessMethod) AccessSpec() cpi.AccessSpec {
	return m.spec
}

func (m *accessMethod) Close() error {
	var err error
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.blob != nil {
		err = m.blob.Close()
		m.blob = nil
	}
	return err
}

func (m *accessMethod) Get() ([]byte, error) {
	return blobaccess.BlobData(m.getBlob())
}

func (m *accessMethod) Reader() (io.ReadCloser, error) {
	return blobaccess.BlobReader(m.getBlob())
}

func (m *accessMethod) MimeType() string {
	return m.info.MediaType
}

func (m *accessMethod) getBlob() (blobaccess.BlobAccess, error) {
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

func (m *accessMethod) GetConsumerId(uctx ...credentials.UsageContext) credentials.ConsumerIdentity {
	if len(m.info.ConsumerId) == 0 {
		return nil
	}
	return m.info.ConsumerId
}

func (m *accessMethod) GetIdentityMatcher() string {
	return hostpath.IDENTITY_TYPE
}
