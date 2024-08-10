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

type repository struct {
	cpi.RepositoryImplBase
	logger logging.UnboundLogger
	spec   *RepositorySpec
	ctf    *ctf.Repository
	client git.Client
}

var (
	_ cpi.RepositoryImpl                   = (*repository)(nil)
	_ credentials.ConsumerIdentityProvider = &repository{}
)

func New(ctx cpi.Context, spec *RepositorySpec) (Repository, error) {
	urs := spec.UniformRepositorySpec()
	i := &repository{
		RepositoryImplBase: cpi.NewRepositoryImplBase(ctx),
		logger:             logging.DynamicLogger(ctx, logging.NewAttribute(ocmlog.ATTR_HOST, urs.Host)),
		spec:               spec,
	}

	var err error
	if i.client, err = git.NewClient(spec.URL); err != nil {
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

	return cpi.NewRepository(i), nil
}

func (r *repository) GetSpecification() cpi.RepositorySpec {
	return r.spec
}

func (r *repository) Close() error {
	return r.ctf.Close()
}

func (r *repository) GetIdentityMatcher() string {
	return identity.CONSUMER_TYPE
}

func (r *repository) NamespaceLister() cpi.NamespaceLister {
	return r.ctf.NamespaceLister()
}

func (r *repository) IsReadOnly() bool {
	return false
}

func (r *repository) ExistsArtifact(name string, version string) (bool, error) {
	if err := r.client.Refresh(context.Background()); err != nil {
		return false, err
	}
	return r.ctf.ExistsArtifact(name, version)
}

func (r *repository) LookupArtifact(name string, version string) (cpi.ArtifactAccess, error) {
	if err := r.client.Refresh(context.Background()); err != nil {
		return nil, err
	}
	return r.ctf.LookupArtifact(name, version)
}

func (r *repository) LookupNamespace(name string) (cpi.NamespaceAccess, error) {
	if err := r.client.Refresh(context.Background()); err != nil {
		return nil, err
	}
	return NewNamespace(r, name)
}

func (r *repository) GetConsumerId(ctx ...cpicredentials.UsageContext) cpicredentials.ConsumerIdentity {
	return nil
}
