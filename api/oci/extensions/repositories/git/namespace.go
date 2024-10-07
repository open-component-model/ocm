package git

import (
	"context"

	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/tech/git"
)

func NewNamespace(repo *RepositoryImpl, name string) (cpi.NamespaceAccess, error) {
	ctfNamespace, err := repo.Repository.LookupNamespace(name)
	if err != nil {
		return nil, err
	}
	return &namespace{
		client:          repo.client,
		NamespaceAccess: ctfNamespace,
	}, nil
}

type namespace struct {
	client git.Client
	cpi.NamespaceAccess
}

var _ cpi.NamespaceAccess = (*namespace)(nil)

func (n *namespace) ListTags() ([]string, error) {
	if err := n.client.Refresh(context.Background()); err != nil {
		return nil, err
	}
	return n.NamespaceAccess.ListTags()
}

func (n *namespace) GetBlobData(digest digest.Digest) (int64, cpi.DataAccess, error) {
	if err := n.client.Refresh(context.Background()); err != nil {
		return 0, nil, err
	}
	return n.NamespaceAccess.GetBlobData(digest)
}

func (n *namespace) GetArtifact(version string) (cpi.ArtifactAccess, error) {
	if err := n.client.Refresh(context.Background()); err != nil {
		return nil, err
	}
	return n.NamespaceAccess.GetArtifact(version)
}

func (n *namespace) HasArtifact(vers string) (bool, error) {
	if err := n.client.Refresh(context.Background()); err != nil {
		return false, err
	}
	return n.NamespaceAccess.HasArtifact(vers)
}
