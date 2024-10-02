package git

import (
	"fmt"
	"io"
	"ocm.software/ocm/api/tech/git/identity"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/mandelsoft/goutils/errors"
	giturls "github.com/whilp/git-urls"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/internal"
	techgit "ocm.software/ocm/api/tech/git"
	"ocm.software/ocm/api/utils/accessio"
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

func newMethod(componentVersionAccess internal.ComponentVersionAccess, accessSpec *AccessSpec) (accspeccpi.AccessMethodImpl, error) {
	u, err := giturls.Parse(accessSpec.RepoURL)
	if err != nil {
		return nil, errors.ErrInvalidWrap(err, "repository repoURL", accessSpec.RepoURL)
	}
	if err := plumbing.ReferenceName(accessSpec.Ref).Validate(); err != nil {
		return nil, errors.ErrInvalidWrap(err, "commit hash", accessSpec.Ref)
	}

	creds, cid, err := getCreds(accessSpec.RepoURL, componentVersionAccess.GetContext().CredentialsContext())
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials for repository %s: %w", accessSpec.RepoURL, err)
	}

	auth, err := techgit.AuthFromCredentials(creds)
	if err != nil && !errors.Is(err, techgit.ErrNoValidGitCredentials) {
		return nil, fmt.Errorf("failed to get auth method for repository %s: %w", accessSpec.RepoURL, err)
	}

	gitDownloader := git.NewDownloader(u.String(), accessSpec.Ref, accessSpec.PathSpec, auth)
	cachedGitBlobAccessor := accessobj.CachedBlobAccessForWriter(
		componentVersionAccess.GetContext(),
		mime.MIME_OCTET,
		accessio.NewWriteAtWriter(gitDownloader.Download),
	)
	jointCloser := func() error {
		return errors.Join(gitDownloader.Close(), cachedGitBlobAccessor.Close())
	}

	return &accessMethod{
		spec:   accessSpec,
		access: cachedGitBlobAccessor,
		close:  jointCloser,
		cid:    cid,
	}, nil
}

type accessMethod struct {
	spec   *AccessSpec
	access blobaccess.BlobAccess
	close  func() error

	cid credentials.ConsumerIdentity
}

var _ accspeccpi.AccessMethodImpl = &accessMethod{}

func (m *accessMethod) Close() error {
	if m.access == nil {
		return nil
	}

	var err error
	if m.close != nil {
		err = m.close()
	}
	err = errors.Join(err, m.access.Close())

	return err
}

func (m *accessMethod) Get() ([]byte, error) {
	return m.access.Get()
}

func (m *accessMethod) Reader() (io.ReadCloser, error) {
	return m.access.Reader()
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

func (m *accessMethod) GetConsumerId(_ ...credentials.UsageContext) credentials.ConsumerIdentity {
	return m.cid
}

func (m *accessMethod) GetIdentityMatcher() string {
	return identity.CONSUMER_TYPE
}

func getCreds(repoURL string, cctx credentials.Context) (credentials.Credentials, credentials.ConsumerIdentity, error) {
	id, err := identity.GetConsumerId(repoURL)
	if err != nil {
		return nil, nil, err
	}
	creds, err := credentials.CredentialsForConsumer(cctx.CredentialsContext(), id, identity.IdentityMatcher)
	if creds == nil || err != nil {
		return nil, id, err
	}
	return creds, id, nil
}
