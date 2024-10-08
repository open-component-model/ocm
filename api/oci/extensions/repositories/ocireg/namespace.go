package ocireg

import (
	"context"
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/opencontainers/go-digest"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry"

	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/oci/cpi/support"
	"ocm.software/ocm/api/oci/extensions/actions/oci-repository-prepare"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/logging"
	common "ocm.software/ocm/api/utils/misc"
)

type NamespaceContainer struct {
	impl    support.NamespaceAccessImpl
	repo    *RepositoryImpl
	checked bool
	ociRepo registry.Repository
}

var _ support.NamespaceContainer = (*NamespaceContainer)(nil)

func NewNamespace(repo *RepositoryImpl, name string) (cpi.NamespaceAccess, error) {
	ref := repo.GetRef(name, "")
	ociRepo, err := repo.getResolver(ref, name)
	if err != nil {
		return nil, err
	}

	c := &NamespaceContainer{
		repo:    repo,
		ociRepo: ociRepo,
	}
	return support.NewNamespaceAccess(name, c, repo)
}

func (n *NamespaceContainer) Close() error {
	return n.repo.Close()
}

func (n *NamespaceContainer) SetImplementation(impl support.NamespaceAccessImpl) {
	n.impl = impl
}

func (n *NamespaceContainer) IsReadOnly() bool {
	return n.repo.IsReadOnly()
}

func (n *NamespaceContainer) GetBlobData(digest digest.Digest) (int64, cpi.DataAccess, error) {
	n.repo.GetContext().Logger().Debug("getting blob", "digest", digest)

	acc, err := NewDataAccess(n.ociRepo, digest, false)
	if err != nil {
		return -1, nil, fmt.Errorf("failed to construct data access: %w", err)
	}

	n.repo.GetContext().Logger().Debug("getting blob done", "digest", digest, "size", blobaccess.BLOB_UNKNOWN_SIZE, "error", logging.ErrorMessage(err))
	return blobaccess.BLOB_UNKNOWN_SIZE, acc, err
}

func (n *NamespaceContainer) AddBlob(blob cpi.BlobAccess) error {
	log := n.repo.GetContext().Logger()
	log.Debug("adding blob", "digest", blob.Digest())

	if err := n.assureCreated(); err != nil {
		return err
	}

	if err := push(dummyContext, n.ociRepo, blob); err != nil {
		return err
	}

	log.Debug("adding blob done", "digest", blob.Digest())
	return nil
}

func (n *NamespaceContainer) ListTags() ([]string, error) {
	var result []string
	if err := n.ociRepo.Tags(dummyContext, "", func(tags []string) error {
		result = append(result, tags...)

		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func (n *NamespaceContainer) GetArtifact(i support.NamespaceAccessImpl, vers string) (cpi.ArtifactAccess, error) {
	ref := n.repo.GetRef(n.impl.GetNamespace(), vers)
	n.repo.GetContext().Logger().Debug("get artifact", "ref", ref)
	desc, err := n.ociRepo.Resolve(context.Background(), ref)
	n.repo.GetContext().Logger().Debug("done", "digest", desc.Digest, "size", desc.Size, "mimetype", desc.MediaType, "error", logging.ErrorMessage(err))
	if err != nil {
		if errors.Is(err, errdef.ErrNotFound) {
			return nil, errors.ErrNotFound(cpi.KIND_OCIARTIFACT, ref, n.impl.GetNamespace())
		}
		return nil, err
	}

	acc, err := NewDataAccess(n.ociRepo, desc.Digest, false)
	if err != nil {
		return nil, fmt.Errorf("failed to construct data access: %w", err)
	}

	return support.NewArtifactForBlob(i, blobaccess.ForDataAccess(desc.Digest, desc.Size, desc.MediaType, acc))
}

func (n *NamespaceContainer) HasArtifact(vers string) (bool, error) {
	ref := n.repo.GetRef(n.impl.GetNamespace(), vers)
	n.repo.GetContext().Logger().Debug("check artifact", "ref", ref)
	desc, err := n.ociRepo.Resolve(context.Background(), ref)
	n.repo.GetContext().Logger().Debug("done", "digest", desc.Digest, "size", desc.Size, "mimetype", desc.MediaType, "error", logging.ErrorMessage(err))
	if err != nil {
		if errors.Is(err, errdef.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (n *NamespaceContainer) assureCreated() error {
	if n.checked {
		return nil
	}
	var props common.Properties
	if creds, err := n.repo.getCreds(n.impl.GetNamespace()); err == nil && creds != nil {
		props = creds.Properties()
	}
	r, err := oci_repository_prepare.Execute(n.repo.GetContext().GetActions(), n.repo.info.HostPort(), n.impl.GetNamespace(), props)
	n.checked = true
	if err != nil {
		return err
	}
	if r != nil {
		n.repo.GetContext().Logger().Debug("prepare action executed", "message", r.Message)
	}
	return nil
}

func (n *NamespaceContainer) AddArtifact(artifact cpi.Artifact, tags ...string) (access blobaccess.BlobAccess, err error) {
	blob, err := artifact.Blob()
	if err != nil {
		return nil, err
	}

	if n.repo.info.Legacy {
		blob = artdesc.MapArtifactBlobMimeType(blob, true)
	}

	n.repo.GetContext().Logger().Debug("adding artifact", "digest", blob.Digest(), "mimetype", blob.MimeType())

	if err := n.assureCreated(); err != nil {
		return nil, err
	}

	if len(tags) > 0 {
		for _, tag := range tags {
			if err := n.pushTag(blob, tag); err != nil {
				return nil, fmt.Errorf("failed to push tag %s: %w", tag, err)
			}
		}
	}

	return blob, err
}

func (n *NamespaceContainer) pushTag(blob blobaccess.BlobAccess, tag string) error {
	reader, err := blob.Reader()
	if err != nil {
		return err
	}
	expectedDescriptor := *artdesc.DefaultBlobDescriptor(blob)
	if err := n.ociRepo.PushReference(context.Background(), expectedDescriptor, reader, tag); err != nil {
		return fmt.Errorf("unable to push: %w", err)
	}
	return nil
}

func (n *NamespaceContainer) AddTags(digest digest.Digest, tags ...string) error {
	ref := n.repo.GetRef(n.impl.GetNamespace(), digest.String())
	desc, err := n.ociRepo.Resolve(context.Background(), ref)
	if err != nil {
		return fmt.Errorf("unable to resolve: %w", err)
	}

	acc, err := NewDataAccess(n.ociRepo, desc.Digest, false)
	if err != nil {
		return fmt.Errorf("error creating new data access: %w", err)
	}

	if err := n.assureCreated(); err != nil {
		return err
	}

	blob := blobaccess.ForDataAccess(desc.Digest, desc.Size, desc.MediaType, acc)
	for _, tag := range tags {
		if err := n.pushTag(blob, tag); err != nil {
			return fmt.Errorf("failed to push tag %s: %w", tag, err)
		}
	}

	return nil
}

func (n *NamespaceContainer) NewArtifact(i support.NamespaceAccessImpl, art ...cpi.Artifact) (cpi.ArtifactAccess, error) {
	if n.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	return support.NewArtifact(i, art...)
}
