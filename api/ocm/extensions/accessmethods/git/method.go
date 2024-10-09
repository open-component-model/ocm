package git

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/mandelsoft/goutils/errors"
	giturls "github.com/whilp/git-urls"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/internal"
	"ocm.software/ocm/api/tech/git/identity"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	gitblob "ocm.software/ocm/api/utils/blobaccess/git"
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

	// Commit defines the hash of the commit in string format to checkout from the Ref
	Commit string `json:"commit"`

	// PathSpec is a path in the repository to download, can be a file or a regex matching multiple files
	PathSpec string `json:"pathSpec"`
}

// AccessSpecOptions defines a set of options which can be applied to the access spec.
type AccessSpecOptions func(s *AccessSpec)

// New creates a new git registry access spec version v1.
func New(url, ref, commit, pathSpec string, opts ...AccessSpecOptions) *AccessSpec {
	s := &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		RepoURL:             url,
		Ref:                 ref,
		Commit:              commit,
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

func (a *AccessSpec) AccessMethod(cva internal.ComponentVersionAccess) (internal.AccessMethod, error) {
	_, err := giturls.Parse(a.RepoURL)
	if err != nil {
		return nil, errors.ErrInvalidWrap(err, "repository repoURL", a.RepoURL)
	}
	if err := plumbing.ReferenceName(a.Ref).Validate(); err != nil {
		return nil, errors.ErrInvalidWrap(err, "commit hash", a.Ref)
	}
	creds, _, err := getCreds(a.RepoURL, cva.GetContext().CredentialsContext())
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials for repository %s: %w", a.RepoURL, err)
	}

	octx := cva.GetContext()

	opts := []gitblob.Option{
		gitblob.WithLoggingContext(octx),
		gitblob.WithCredentialContext(octx),
		gitblob.WithURL(a.RepoURL),
		gitblob.WithRef(a.Ref),
		gitblob.WithCommit(a.Commit),
		gitblob.WithCachingFileSystem(vfsattr.Get(octx)),
	}
	if creds != nil {
		opts = append(opts, gitblob.WithCredentials(creds))
	}

	factory := func() (blobaccess.BlobAccess, error) {
		return gitblob.BlobAccess(opts...)
	}

	return accspeccpi.AccessMethodForImplementation(accspeccpi.NewDefaultMethodImpl(
		cva,
		a,
		"",
		mime.MIME_TGZ,
		factory,
	), nil)
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
