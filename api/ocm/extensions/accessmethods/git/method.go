package git

import (
	"fmt"

	giturls "github.com/chainguard-dev/git-urls"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/mandelsoft/goutils/errors"
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
	Type         = "git"
	TypeV1Alpha1 = Type + runtime.VersionSeparator + "v1alpha1"
)

func init() {
	// If we remove the default registration, also the docs are gone.
	// so we leave the default registration in.
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](Type, accspeccpi.WithDescription(usage)))
	accspeccpi.RegisterAccessType(accspeccpi.NewAccessSpecType[*AccessSpec](TypeV1Alpha1, accspeccpi.WithFormatSpec(formatV1), accspeccpi.WithConfigHandler(ConfigHandler())))
}

// AccessSpec describes the access for a GitHub registry.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// Repository is the repository URL
	Repository string `json:"repository"`

	// Ref defines the hash of the commit
	Ref string `json:"ref,omitempty"`

	// Commit defines the hash of the commit in string format to checkout from the Ref
	Commit string `json:"commit,omitempty"`
}

// AccessSpecOptions defines a set of options which can be applied to the access spec.
type AccessSpecOptions func(s *AccessSpec)

func WithCommit(commit string) AccessSpecOptions {
	return func(s *AccessSpec) {
		s.Commit = commit
	}
}

func WithRef(ref string) AccessSpecOptions {
	return func(s *AccessSpec) {
		s.Ref = ref
	}
}

// New creates a new git registry access spec version v1.
func New(url string, opts ...AccessSpecOptions) *AccessSpec {
	s := &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedTypedObject(Type),
		Repository:          url,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

func (a *AccessSpec) Describe(internal.Context) string {
	return fmt.Sprintf("git commit %s[%s]", a.Repository, a.Ref)
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
	_, err := giturls.Parse(a.Repository)
	if err != nil {
		return nil, errors.ErrInvalidWrap(err, "repository repoURL", a.Repository)
	}
	if err := plumbing.ReferenceName(a.Ref).Validate(); err != nil {
		return nil, errors.ErrInvalidWrap(err, "commit hash", a.Ref)
	}
	creds, _, err := getCreds(a.Repository, cva.GetContext().CredentialsContext())
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials for repository %s: %w", a.Repository, err)
	}

	octx := cva.GetContext()

	opts := []gitblob.Option{
		gitblob.WithLoggingContext(octx),
		gitblob.WithCredentialContext(octx),
		gitblob.WithURL(a.Repository),
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
