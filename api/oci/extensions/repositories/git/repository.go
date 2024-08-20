package git

import (
	"context"

	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/builtin/oci/identity"
	cpicredentials "ocm.software/ocm/api/credentials/cpi"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/oci/extensions/repositories/ctf"
	"ocm.software/ocm/api/tech/git"
	ocmlog "ocm.software/ocm/api/utils/logging"
)

type Repository interface {
	cpi.Repository
}

type RepositoryImpl struct {
	cpi.RepositoryImplBase
	logger logging.UnboundLogger
	spec   *RepositorySpec
	ctf    *ctf.Repository
	client git.Client
}

var (
	_ cpi.RepositoryImpl                   = (*RepositoryImpl)(nil)
	_ credentials.ConsumerIdentityProvider = &RepositoryImpl{}
)

func New(ctx cpi.Context, spec *RepositorySpec) (Repository, error) {
	urs := spec.UniformRepositorySpec()
	i := &RepositoryImpl{
		RepositoryImplBase: cpi.NewRepositoryImplBase(ctx),
		logger:             logging.DynamicLogger(ctx, logging.NewAttribute(ocmlog.ATTR_HOST, urs.Host)),
		spec:               spec,
	}

	var err error
	if i.client, err = git.NewClient(spec.ToClientOptions()); err != nil {
		return nil, err
	}

	repo, err := ctf.New(ctx, &ctf.RepositorySpec{
		StandardOptions: spec.StandardOptions,
		AccessMode:      spec.AccessMode,
	}, i.client, i.client, vfs.FileMode(0o770))
	if err != nil {
		return nil, err
	}
	i.ctf = repo

	return cpi.NewRepository(i, "git"), nil
}

func (r *RepositoryImpl) GetSpecification() cpi.RepositorySpec {
	return r.spec
}

func (r *RepositoryImpl) Close() error {
	return r.ctf.Close()
}

func (r *RepositoryImpl) GetIdentityMatcher() string {
	return identity.CONSUMER_TYPE
}

func (r *RepositoryImpl) NamespaceLister() cpi.NamespaceLister {
	return r.ctf.NamespaceLister()
}

func (r *RepositoryImpl) IsReadOnly() bool {
	return false
}

func (r *RepositoryImpl) ExistsArtifact(name string, version string) (bool, error) {
	if err := r.client.Refresh(context.Background()); err != nil {
		return false, err
	}
	return r.ctf.ExistsArtifact(name, version)
}

func (r *RepositoryImpl) LookupArtifact(name string, version string) (cpi.ArtifactAccess, error) {
	if err := r.client.Refresh(context.Background()); err != nil {
		return nil, err
	}
	return r.ctf.LookupArtifact(name, version)
}

func (r *RepositoryImpl) LookupNamespace(name string) (cpi.NamespaceAccess, error) {
	if err := r.client.Refresh(context.Background()); err != nil {
		return nil, err
	}
	return NewNamespace(r, name)
}

func (r *RepositoryImpl) GetConsumerId(ctx ...cpicredentials.UsageContext) cpicredentials.ConsumerIdentity {
	return nil
}
