package wget

import (
	"fmt"
	"io"
	"sync"

	"github.com/mandelsoft/goutils/optionutils"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/tech/wget/identity"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/wget"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/runtime"
)

// Type is the access type for a blob on an http server .
const (
	Type   = "wget"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](Type, accspeccpi.WithDescription(usage)))
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](TypeV1, accspeccpi.WithFormatSpec(formatV1), accspeccpi.WithConfigHandler(ConfigHandler())))
}

func Is(spec accspeccpi.AccessSpec) bool {
	return spec != nil && spec.GetKind() == Type
}

// New creates a new WGET accessor for http resources.
func New(url string, opts ...Option) *AccessSpec {
	eff := optionutils.EvalOptions(opts...)

	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		URL:                 url,
		MediaType:           eff.MimeType,
		Header:              eff.Header,
		Verb:                eff.Verb,
		Body:                eff.Body,
		NoRedirect:          optionutils.AsValue(eff.NoRedirect),
	}
}

// AccessSpec describes the access for files on HTTP and HTTPS servers.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// URLs to the files on a server
	URL string `json:"URL"`
	// MediaType is the media type of the object represented by the blob
	MediaType string `json:"mediaType"`
	// Header to be passed in the http request
	Header map[string][]string `json:"header"`
	// Verb is the http verb to be used for the request
	Verb string `json:"verb"`
	// Body is the body to be included in the http request
	Body io.Reader `json:"body"`
	// NoRedirect allows to disable redirects
	NoRedirect bool `json:"noRedirect"`
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

func (a *AccessSpec) AccessMethod(access accspeccpi.ComponentVersionAccess) (accspeccpi.AccessMethod, error) {
	return accspeccpi.AccessMethodForImplementation(&accessMethod{comp: access, spec: a}, nil)
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
	if m.spec.MediaType != "" {
		return m.spec.MediaType
	}
	blob, err := m.getBlob()
	if err != nil {
		return mime.MIME_OCTET
	}
	return blob.MimeType()
}

func (m *accessMethod) getBlob() (blobaccess.BlobAccess, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.blob != nil {
		return m.blob, nil
	}

	blob, err := wget.BlobAccess(m.spec.URL,
		wget.WithMimeType(m.spec.MediaType),
		wget.WithCredentialContext(m.comp.GetContext()),
		wget.WithLoggingContext(m.comp.GetContext()),
		wget.WithHeader(m.spec.Header),
		wget.WithVerb(m.spec.Verb),
		wget.WithBody(m.spec.Body),
		wget.WithNoRedirect(m.spec.NoRedirect))
	if err != nil {
		return nil, err
	}

	m.blob = blob
	return m.blob, nil
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
