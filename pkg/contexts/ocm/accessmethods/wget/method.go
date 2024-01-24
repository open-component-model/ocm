package wget

import (
	"fmt"
	"io"
	"sync"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/blobaccess/wget"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/wget/identity"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/runtime"
)

// Type is the access type for a blob on an http server .
const (
	Type   = "wget"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](Type))
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](TypeV1))
}

func Is(spec accspeccpi.AccessSpec) bool {
	return spec != nil && spec.GetKind() == Type
}

// New creates a new WGET accessor for http resources.
func New(url, mime string) *AccessSpec {
	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		URL:                 url,
		MediaType:           mime,
	}
}

// AccessSpec describes the access for files on HTTP and HTTPS servers.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// URLs to the files on a server
	URL string `json:"URL"`

	// MediaType is the media type of the object represented by the blob
	MediaType string `json:"mediaType"`
}

var _ accspeccpi.AccessSpec = (*AccessSpec)(nil)

func (a *AccessSpec) Describe(ctx accspeccpi.Context) string {
	return fmt.Sprintf("Files from %s", a.URL)
}

func (a *AccessSpec) IsLocal(ctx accspeccpi.Context) bool {
	return false
}

func (a *AccessSpec) GlobalAccessSpec(ctx accspeccpi.Context) accspeccpi.AccessSpec {
	return a
}

func (a *AccessSpec) GetMimeType() string {
	if a.MediaType == "" {
		return mime.MIME_OCTET
	}
	return a.MediaType
}

func (a *AccessSpec) AccessMethod(access accspeccpi.ComponentVersionAccess) (accspeccpi.AccessMethod, error) {
	return accspeccpi.AccessMethodForImplementation(&accessMethod{comp: access, spec: a}, nil)
}

func (a *AccessSpec) GetInexpensiveContentVersionIdentity(access accspeccpi.ComponentVersionAccess) string {
	return ""
}

///////////////////

func (a *AccessSpec) GetURL() string {
	return a.URL
}

////////////////////////////////////////////////////////////////////////////////

type accessMethod struct {
	lock sync.Mutex
	blob blobaccess.BlobAccess
	comp accspeccpi.ComponentVersionAccess
	spec *AccessSpec
}

var _ accspeccpi.AccessMethodImpl = (*accessMethod)(nil)

func (_ *accessMethod) IsLocal() bool {
	return false
}

func (m *accessMethod) GetKind() string {
	return Type
}

func (m *accessMethod) AccessSpec() accspeccpi.AccessSpec {
	return m.spec
}

func (m *accessMethod) Get() ([]byte, error) {
	return blobaccess.BlobData(m.getBlob())
}

func (m *accessMethod) Reader() (io.ReadCloser, error) {
	return blobaccess.BlobReader(m.getBlob())
}

func (m *accessMethod) MimeType() string {
	return m.spec.MediaType
}

func (m *accessMethod) getBlob() (blobaccess.BlobAccess, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	return wget.BlobAccessForWget(m.spec.URL,
		wget.WithMimeType(m.spec.GetMimeType()),
		wget.WithCredentialContext(m.comp.GetContext()),
		wget.WithLoggingContext(m.comp.GetContext()))
}

func (m *accessMethod) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	var err error
	if m.blob != nil {
		err = m.blob.Close()
		m.blob = nil
	}

	return err
}

func (m *accessMethod) GetConsumerId(uctx ...credentials.UsageContext) credentials.ConsumerIdentity {
	return identity.GetConsumerId(m.spec.URL)
}

func (m *accessMethod) GetIdentityMatcher() string {
	return identity.CONSUMER_TYPE
}
