package git

import (
	"context"
	"fmt"

	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/oci/cpi/support"
	"ocm.software/ocm/api/oci/internal"
	"ocm.software/ocm/api/tech/git"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
)

func NewNamespace(repo *RepositoryImpl, name string) (cpi.NamespaceAccess, error) {
	ctfNamespace, err := newNamespaceContainer(repo, name)
	if err != nil {
		return nil, err
	}
	return support.NewNamespaceAccess(name, ctfNamespace, repo, "Git RepositoryImpl Branch")
}

type namespaceContainer struct {
	impl   support.NamespaceAccessImpl
	name   string
	client git.Client
	ctf    cpi.NamespaceAccess
}

var _ support.NamespaceContainer = (*namespaceContainer)(nil)

func newNamespaceContainer(repo *RepositoryImpl, name string) (support.NamespaceContainer, error) {
	ctfNamespace, err := repo.ctf.LookupNamespace(name)
	if err != nil {
		return nil, err
	}
	return &namespaceContainer{
		name:   name,
		client: repo.client,
		ctf:    ctfNamespace,
	}, nil
}

func (n *namespaceContainer) SetImplementation(impl support.NamespaceAccessImpl) {
	n.impl = impl
}

func (n *namespaceContainer) IsReadOnly() bool {
	return false
}

func (n *namespaceContainer) Close() error {
	if err := n.ctf.Close(); err != nil {
		return err
	}
	return n.client.Update(context.Background(), fmt.Sprintf("namespace update %q", n.name), true)
}

func (n *namespaceContainer) ListTags() ([]string, error) {
	if err := n.client.Refresh(context.Background()); err != nil {
		return nil, err
	}
	return n.ctf.ListTags()
}

func (n *namespaceContainer) GetBlobData(digest digest.Digest) (int64, cpi.DataAccess, error) {
	if err := n.client.Refresh(context.Background()); err != nil {
		return 0, nil, err
	}

	return n.ctf.GetBlobData(digest)
}

func (n *namespaceContainer) AddBlob(blob cpi.BlobAccess) error {
	if err := n.ctf.AddBlob(blob); err != nil {
		return err
	}

	if err := n.client.Update(context.Background(), GenerateCommitMessageForBlob(OperationAdd, blob), false); err != nil {
		return err
	}
	return nil
}

func (n *namespaceContainer) GetArtifact(i support.NamespaceAccessImpl, vers string) (cpi.ArtifactAccess, error) {
	if err := n.client.Refresh(context.Background()); err != nil {
		return nil, err
	}

	return n.ctf.GetArtifact(vers)
}

func (n *namespaceContainer) HasArtifact(vers string) (bool, error) {
	if err := n.client.Refresh(context.Background()); err != nil {
		return false, err
	}
	return n.ctf.HasArtifact(vers)
}

func (n *namespaceContainer) AddArtifact(artifact cpi.Artifact, tags ...string) (access blobaccess.BlobAccess, err error) {
	blobAccess, err := n.ctf.AddArtifact(artifact, tags...)
	if err != nil {
		return nil, err
	}
	msg := GenerateCommitMessageForArtifact(OperationAdd, artifact)

	if err := n.client.Update(context.Background(), msg, false); err != nil {
		return nil, err
	}

	return blobAccess, nil
}

func (n *namespaceContainer) AddTags(digest digest.Digest, tags ...string) error {
	if err := n.ctf.AddTags(digest, tags...); err != nil {
		return err
	}

	if err := n.client.Update(context.Background(), fmt.Sprintf("added tags %s to %s", tags, digest.String()), true); err != nil {
		return err
	}

	return nil
}

func (n *namespaceContainer) NewArtifact(i support.NamespaceAccessImpl, art ...cpi.Artifact) (cpi.ArtifactAccess, error) {
	artifactAccess, err := n.ctf.NewArtifact(art...)
	if err != nil {
		return nil, err
	}
	return &artifactContainer{
		client:         n.client,
		ArtifactAccess: artifactAccess,
	}, nil
}

type Operation string

const (
	OperationAdd  Operation = "add"
	OperationMod  Operation = "mod"
	OperationSync Operation = "sync"
)

func GenerateCommitMessageForArtifact(operation Operation, artifact cpi.Artifact) string {
	a := artifact.Artifact()

	var msg string
	if artifact.IsManifest() {
		msg = fmt.Sprintf("update(ocm): %s manifest %s (%s)", operation, a.Digest(), a.MimeType())
	} else if artifact.IsIndex() {
		msg = fmt.Sprintf("update(ocm): %s index %s (%s)", operation, a.Digest(), a.MimeType())
	} else {
		msg = fmt.Sprintf("update(ocm): %s artifact %s (%s)", operation, a.Digest(), a.MimeType())
	}
	return msg
}

func GenerateCommitMessageForBlob(operation Operation, blob cpi.BlobAccess) string {
	var msg string
	if blob.DigestKnown() {
		msg = fmt.Sprintf("update(ocm): %s blob %s of type %s", operation, blob.Digest(), blob.MimeType())
	} else {
		msg = fmt.Sprintf("update(ocm): %s blob of type %s", operation, blob.MimeType())
	}
	return msg
}

type artifactContainer struct {
	client git.Client
	cpi.ArtifactAccess
}

var _ cpi.ArtifactAccess = (*artifactContainer)(nil)

func (a *artifactContainer) Close() error {
	if err := a.ArtifactAccess.Close(); err != nil {
		return err
	}
	return a.client.Update(context.Background(), GenerateCommitMessageForArtifact(OperationSync, a.ArtifactAccess), true)
}

func (a *artifactContainer) Dup() (cpi.ArtifactAccess, error) {
	access, err := a.ArtifactAccess.Dup()
	if err != nil {
		return nil, err
	}
	return &artifactContainer{
		client:         a.client,
		ArtifactAccess: access,
	}, nil
}

func (a *artifactContainer) AddBlob(access internal.BlobAccess) error {
	if err := a.ArtifactAccess.AddBlob(access); err != nil {
		return err
	}
	return a.client.Update(context.Background(), GenerateCommitMessageForBlob(OperationAdd, access), false)
}

func (a *artifactContainer) AddArtifact(artifact cpi.Artifact, platform *artdesc.Platform) (cpi.BlobAccess, error) {
	b, err := a.ArtifactAccess.AddArtifact(artifact, platform)
	if err != nil {
		return nil, err
	}
	return b, a.client.Update(context.Background(), GenerateCommitMessageForArtifact(OperationAdd, artifact), true)
}

func (a *artifactContainer) AddLayer(access cpi.BlobAccess, descriptor *artdesc.Descriptor) (int, error) {
	n, err := a.ArtifactAccess.AddLayer(access, descriptor)
	if err != nil {
		return -1, err
	}
	return n, a.client.Update(context.Background(), GenerateCommitMessageForBlob(OperationMod, access), false)
}

func (a *artifactContainer) NewArtifact(artifact ...cpi.Artifact) (cpi.ArtifactAccess, error) {
	access, err := a.ArtifactAccess.NewArtifact(artifact...)
	if err != nil {
		return nil, err
	}
	return &artifactContainer{
		client:         a.client,
		ArtifactAccess: access,
	}, nil
}

func (a *artifactContainer) ManifestAccess() cpi.ManifestAccess {
	return &manifestContainer{
		client:         a.client,
		ManifestAccess: a.ArtifactAccess.ManifestAccess(),
	}
}

func (a *artifactContainer) IndexAccess() cpi.IndexAccess {
	return &indexContainer{
		client:      a.client,
		IndexAccess: a.ArtifactAccess.IndexAccess(),
	}
}

type manifestContainer struct {
	cpi.ManifestAccess
	client git.Client
}

var _ cpi.ManifestAccess = (*manifestContainer)(nil)

func (m *manifestContainer) AddBlob(access internal.BlobAccess) error {
	if err := m.ManifestAccess.AddBlob(access); err != nil {
		return err
	}
	return m.client.Update(context.Background(), GenerateCommitMessageForBlob(OperationAdd, access), false)
}

func (m *manifestContainer) AddLayer(access internal.BlobAccess, descriptor *artdesc.Descriptor) (int, error) {
	n, err := m.ManifestAccess.AddLayer(access, descriptor)
	if err != nil {
		return -1, err
	}
	return n, m.client.Update(context.Background(), GenerateCommitMessageForBlob(OperationMod, access), false)
}

func (m *manifestContainer) SetConfigBlob(blob internal.BlobAccess, d *artdesc.Descriptor) error {
	if err := m.ManifestAccess.SetConfigBlob(blob, d); err != nil {
		return err
	}
	return m.client.Update(context.Background(), GenerateCommitMessageForBlob(OperationMod, blob), false)
}

type indexContainer struct {
	cpi.IndexAccess
	client git.Client
}

var _ cpi.IndexAccess = (*indexContainer)(nil)

func (i *indexContainer) GetArtifact(digest digest.Digest) (internal.ArtifactAccess, error) {
	a, err := i.IndexAccess.GetArtifact(digest)
	if err != nil {
		return nil, err
	}
	return &artifactContainer{
		client:         i.client,
		ArtifactAccess: a,
	}, nil
}

func (i *indexContainer) AddBlob(access internal.BlobAccess) error {
	if err := i.IndexAccess.AddBlob(access); err != nil {
		return err
	}
	return i.client.Update(context.Background(), GenerateCommitMessageForBlob(OperationAdd, access), false)
}

func (i *indexContainer) AddArtifact(artifact internal.Artifact, platform *artdesc.Platform) (internal.BlobAccess, error) {
	b, err := i.IndexAccess.AddArtifact(artifact, platform)
	if err != nil {
		return nil, err
	}
	return b, i.client.Update(context.Background(), GenerateCommitMessageForArtifact(OperationAdd, artifact), false)
}
