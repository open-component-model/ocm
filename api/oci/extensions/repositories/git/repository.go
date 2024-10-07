package git

import (
	"context"
	"fmt"

	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/oci/extensions/repositories/ctf"
	"ocm.software/ocm/api/tech/git"
	"ocm.software/ocm/api/tech/git/identity"
	"ocm.software/ocm/api/utils/accessobj"
	ocmlog "ocm.software/ocm/api/utils/logging"
)

const CommitPrefix = "update(ocm)"

type Repository interface {
	cpi.Repository
}

type RepositoryImpl struct {
	logger logging.UnboundLogger
	spec   *RepositorySpec
	*ctf.Repository
	client git.Client
}

var _ cpi.Repository = (*RepositoryImpl)(nil)

func New(ctx cpi.Context, spec *RepositorySpec, creds credentials.Credentials) (Repository, error) {
	urs := spec.UniformRepositorySpec()
	i := &RepositoryImpl{
		logger: logging.DynamicLogger(ctx, logging.NewAttribute(ocmlog.ATTR_HOST, urs.Host)),
		spec:   spec,
	}

	opts := spec.ToClientOptions()

	if creds == nil {
		// if no credentials are provided, try to get them from the context,
		// if the credential is not provided, the client will try to use unauthenticated access, so allow error
		creds, _ = identity.GetCredentials(ctx, spec.URL)
	}

	if creds != nil {
		auth, err := git.AuthFromCredentials(creds)
		if err != nil {
			return nil, fmt.Errorf("failed to create git authentication from given credentials: %w", err)
		}
		opts.AuthMethod = auth
	}

	var err error
	if i.client, err = git.NewClient(opts); err != nil {
		return nil, fmt.Errorf("failed to create new git client for interacting with the repository: %w", err)
	}

	repo, err := ctf.New(ctx, &ctf.RepositorySpec{
		StandardOptions: spec.StandardOptions,
		AccessMode:      spec.AccessMode,
	}, i.client, &repoCloseUpdater{func() error {
		if i.IsReadOnly() {
			return nil
		}
		// on close make sure that we update and push the latest changes
		return i.client.Update(context.Background(), GenerateCommitMessage(), true)
	}}, vfs.FileMode(0o770))
	if err != nil {
		return nil, fmt.Errorf("failed to create new ctf repository within the git repository: %w", err)
	}
	i.Repository = repo

	return i, nil
}

func (r *RepositoryImpl) GetSpecification() cpi.RepositorySpec {
	return r.spec
}

func GenerateCommitMessage() string {
	return fmt.Sprintf("%s: update repository", CommitPrefix)
}

func (r *RepositoryImpl) GetIdentityMatcher() string {
	return identity.CONSUMER_TYPE
}

func (r *RepositoryImpl) IsReadOnly() bool {
	return false
}

func (r *RepositoryImpl) ExistsArtifact(name string, version string) (bool, error) {
	if err := r.client.Refresh(context.Background()); err != nil {
		return false, err
	}
	return r.Repository.ExistsArtifact(name, version)
}

func (r *RepositoryImpl) LookupArtifact(name string, version string) (cpi.ArtifactAccess, error) {
	if err := r.client.Refresh(context.Background()); err != nil {
		return nil, err
	}
	return r.Repository.LookupArtifact(name, version)
}

func (r *RepositoryImpl) LookupNamespace(name string) (cpi.NamespaceAccess, error) {
	if err := r.client.Refresh(context.Background()); err != nil {
		return nil, err
	}
	return NewNamespace(r, name)
}

// small helper to wrap accessio.Closer to allow calling an arbitrary closing logic.
type repoCloseUpdater struct {
	close func() error
}

func (r *repoCloseUpdater) Close(*accessobj.AccessObject) error {
	return r.close()
}
