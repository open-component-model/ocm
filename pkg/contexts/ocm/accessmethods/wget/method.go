package wget

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/wget/identity"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/runtime"
	"io"
	"net/http"
	"sync"
)

// Type is the access type for a blob on an http server .
const (
	Type   = "wget"
	TypeV1 = Type + runtime.VersionSeparator + "v1"

	CACHE_CONTENT_THRESHOLD = 4096
)

func init() {
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](Type))
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](TypeV1))
}

func Is(spec accspeccpi.AccessSpec) bool {
	return spec != nil && spec.GetKind() == Type
}

// New creates a new Helm Chart accessor for helm repositories.
func New(url string) *AccessSpec {
	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		URL:                 url,
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

	log := Logger(m.comp, "URL", m.spec.URL)

	if m.blob != nil {
		return m.blob, nil
	}

	creds, err := credentials.CredentialsForConsumer(m.comp.GetContext(), identity.GetConsumerId(m.spec.URL), identity.IdentityMatcher)
	if err != nil {
		log.Debug("no credentials found for", "url", m.spec.URL)
		return nil, err
	}

	rootCAs := credentials.GetRootCAs(m.comp.GetContext(), creds)
	clientCerts, err := credentials.GetClientCerts(m.comp.GetContext(), creds)
	if err != nil {
		return nil, errors.New("client certificate and private key provided in credentials could not be loaded " +
			"as tls certificate")
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      rootCAs,
			Certificates: clientCerts,
		},
	}

	client := &http.Client{
		Transport: transport,
	}

	request, err := http.NewRequest(http.MethodGet, m.spec.URL, nil)
	if err != nil {
		return nil, err
	}

	user := creds.GetProperty(identity.ATTR_USERNAME)
	password := creds.GetProperty(identity.ATTR_PASSWORD)
	token := creds.GetProperty(identity.ATTR_IDENTITY_TOKEN)

	if user != "" && password != "" {
		request.SetBasicAuth(user, password)
	} else if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Get(m.spec.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Debug("http status code", "", resp.StatusCode)
	if resp.ContentLength < 0 || resp.ContentLength > CACHE_CONTENT_THRESHOLD {
		log.Debug("download to file because content length is", "unkown or greater than", CACHE_CONTENT_THRESHOLD)
		f, err := blobaccess.NewTempFile("", "wget")
		if err != nil {
			return nil, err
		}
		defer f.Close()

		n, err := io.Copy(f.Writer(), resp.Body)
		if err != nil {
			return nil, err
		}
		log.Debug("downloaded", "size", n, "to", f.Name())
		m.blob = f.AsBlob(m.spec.GetMimeType())
	} else {
		log.Debug("download to memory because content length is", "less than", CACHE_CONTENT_THRESHOLD)
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		m.blob = blobaccess.ForData(m.spec.GetMimeType(), buf)
	}

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
