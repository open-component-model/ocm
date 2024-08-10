package git

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/internal"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessio/downloader"
	"ocm.software/ocm/api/utils/accessio/downloader/git"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	Type   = "git"
	TypeV1 = Type + runtime.VersionSeparator + "v1"
)

func init() {
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](Type, accspeccpi.WithDescription(usage)))
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](TypeV1, accspeccpi.WithFormatSpec(formatV1), accspeccpi.WithConfigHandler(ConfigHandler())))
}

// AccessSpec describes the access for a GitHub registry.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// RepoURL is the repository URL
	RepoURL string `json:"repoUrl"`

	// Ref defines the hash of the commit
	Ref string `json:"ref"`

	// PathSpec is a path in the repository to download, can be a file or a regex matching multiple files
	PathSpec string `json:"pathSpec"`

	client     *http.Client
	downloader downloader.Downloader
}

// AccessSpecOptions defines a set of options which can be applied to the access spec.
type AccessSpecOptions func(s *AccessSpec)

// New creates a new git registry access spec version v1.
func New(url, ref string, pathSpec string, opts ...AccessSpecOptions) *AccessSpec {
	s := &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		RepoURL:             url,
		Ref:                 ref,
		PathSpec:            pathSpec,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

func (a *AccessSpec) Describe(internal.Context) string {
	return fmt.Sprintf("git commit %s[%s]", a.RepoURL, a.Ref)
}

func (*AccessSpec) IsLocal(internal.Context) bool {
	return false
}

func (a *AccessSpec) GlobalAccessSpec(accspeccpi.Context) accspeccpi.AccessSpec {
	return a
}

func (*AccessSpec) GetType() string {
	return Type
}

func (a *AccessSpec) AccessMethod(c internal.ComponentVersionAccess) (internal.AccessMethod, error) {
	return accspeccpi.AccessMethodForImplementation(newMethod(c, a))
}

func newMethod(c internal.ComponentVersionAccess, a *AccessSpec) (accspeccpi.AccessMethodImpl, error) {
	u, err := url.Parse(a.RepoURL)
	if err != nil {
		return nil, errors.ErrInvalidWrap(err, "repository repoURL", a.RepoURL)
	}
	if err := plumbing.ReferenceName(a.Ref).Validate(); err != nil {
		return nil, errors.ErrInvalidWrap(err, "commit hash", a.Ref)
	}

	return &accessMethod{
		repoURL:  u.String(),
		compvers: c,
		spec:     a,
		ref:      a.Ref,
	}, nil
}

type accessMethod struct {
	lock   sync.Mutex
	access blobaccess.BlobAccess

	compvers accspeccpi.ComponentVersionAccess
	spec     *AccessSpec

	repoURL string
	path    string
	ref     string
}

var _ accspeccpi.AccessMethodImpl = &accessMethod{}

func (m *accessMethod) Close() error {
	if m.access == nil {
		return nil
	}
	return m.access.Close()
}

func (m *accessMethod) Get() ([]byte, error) {
	if err := m.setup(); err != nil {
		return nil, err
	}
	return m.access.Get()
}

func (m *accessMethod) Reader() (io.ReadCloser, error) {
	if err := m.setup(); err != nil {
		return nil, err
	}
	return m.access.Reader()
}

func (m *accessMethod) setup() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.access != nil {
		return nil
	}

	d := git.NewDownloader(m.repoURL, m.ref, m.path)
	defer d.Close()

	cacheBlobAccess := accessobj.CachedBlobAccessForWriter(
		m.compvers.GetContext(),
		m.MimeType(),
		accessio.NewWriteAtWriter(d.Download),
	)

	m.access = cacheBlobAccess

	return nil
}

func (m *accessMethod) MimeType() string {
	return mime.MIME_OCTET
}

func (*accessMethod) IsLocal() bool {
	return false
}

func (m *accessMethod) GetKind() string {
	return Type
}

func (m *accessMethod) AccessSpec() internal.AccessSpec {
	return m.spec
}
