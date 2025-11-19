package docker

import (
	"fmt"
	"strings"
	"sync"

	"github.com/containers/image/v5/image"
	"github.com/containers/image/v5/types"
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/logging"
	"github.com/moby/moby/client"
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/oci/cpi/support"
	"ocm.software/ocm/api/oci/internal"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
)

type blobHandler struct {
	accessio.BlobCache
}

var _ support.BlobProvider = (*blobHandler)(nil)

func newBlobHandler(cache accessio.BlobCache) support.BlobProvider {
	return &blobHandler{cache}
}

func (b blobHandler) AddBlob(access internal.BlobAccess) error {
	_, _, err := b.BlobCache.AddBlob(access)
	return err
}

////////////////////////////////////////////////////////////////////////////////

// namespaceContainer delegates functionality but blob access to an underlying
// handler.
// blob access is handled locally.
type namespaceContainer struct {
	*namespaceHandler
	blobs support.BlobProvider
}

var _ support.NamespaceContainer = (*namespaceContainer)(nil)

func newNamespaceContainer(handler *namespaceHandler, blobs support.BlobProvider) *namespaceContainer {
	return &namespaceContainer{
		namespaceHandler: handler,
		blobs:            blobs,
	}
}

func NewNamespace(repo *RepositoryImpl, name string) (cpi.NamespaceAccess, error) {
	h, err := newNamespaceHandler(repo)
	if err != nil {
		return nil, err
	}
	// initial container wrapper releases base cache with close of namespace
	// container on last namespace ref.
	// base cache has initial user count of 1.
	return support.NewNamespaceAccess(name, newNamespaceContainer(h, h.blobs), repo, "docker namespace")
}

func (n *namespaceContainer) Close() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	if n.blobs != nil {
		err := n.blobs.Unref()
		n.blobs = nil
		if err != nil {
			return fmt.Errorf("failed to unref: %w", err)
		}
	}
	return nil
}

func (n *namespaceContainer) GetBlobData(digest digest.Digest) (int64, cpi.DataAccess, error) {
	return n.blobs.GetBlobData(digest)
}

func (n *namespaceContainer) AddBlob(blob cpi.BlobAccess) error {
	if err := n.blobs.AddBlob(blob); err != nil {
		return fmt.Errorf("failed to add blob to cache: %w", err)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

type namespaceHandler struct {
	impl  support.NamespaceAccessImpl
	lock  sync.RWMutex
	repo  *RepositoryImpl
	blobs support.BlobProvider
	log   logging.Logger
}

func newNamespaceHandler(repo *RepositoryImpl) (*namespaceHandler, error) {
	cache, err := accessio.NewCascadedBlobCache(nil)
	if err != nil {
		return nil, err
	}

	return &namespaceHandler{
		repo:  repo,
		blobs: newBlobHandler(cache),
		log:   repo.GetContext().Logger(),
	}, nil
}

func (n *namespaceHandler) SetImplementation(impl support.NamespaceAccessImpl) {
	n.impl = impl
}

func (n *namespaceHandler) IsReadOnly() bool {
	return n.repo.IsReadOnly()
}

func (n *namespaceContainer) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	return nil
}

func (n *namespaceHandler) ListTags() ([]string, error) {
	opts := client.ImageListOptions{}
	list, err := n.repo.client.ImageList(dummyContext, opts)
	if err != nil {
		return nil, err
	}
	var result []string
	if n.impl.GetNamespace() == "" {
		for _, e := range list.Items {
			// ID is always the config digest
			// filter images without a repo tag for empty namespace
			if len(e.RepoTags) == 0 {
				d, err := digest.Parse(e.ID)
				if err == nil {
					result = append(result, d.String()[:12])
				}
			}
		}
	} else {
		prefix := n.impl.GetNamespace() + ":"
		for _, e := range list.Items {
			for _, t := range e.RepoTags {
				if strings.HasPrefix(t, prefix) {
					result = append(result, t[len(prefix):])
				}
			}
		}
	}
	return result, nil
}

func (n *namespaceHandler) GetArtifact(i support.NamespaceAccessImpl, vers string) (cpi.ArtifactAccess, error) {
	ref, err := ParseRef(n.impl.GetNamespace(), vers)
	if err != nil {
		return nil, err
	}
	src, err := ref.NewImageSource(dummyContext, n.repo.sysctx)
	if err != nil {
		return nil, err
	}

	opts := types.ManifestUpdateOptions{
		ManifestMIMEType: artdesc.MediaTypeImageManifest,
	}
	un := image.UnparsedInstance(src, nil)
	img, err := image.FromUnparsedImage(dummyContext, n.repo.sysctx, un)
	if err != nil {
		src.Close()
		return nil, err
	}

	img, err = img.UpdatedImage(dummyContext, opts)
	if err != nil {
		src.Close()
		return nil, err
	}

	data, mime, err := img.Manifest(dummyContext)
	if err != nil {
		src.Close()
		return nil, err
	}

	cache, err := accessio.NewCascadedBlobCacheForSource(n.blobs, newDockerSource(img, src))
	if err != nil {
		return nil, err
	}

	priv := i.WithContainer(newNamespaceContainer(n, newBlobHandler(cache)))
	// assure explicit close of wrapper container for artifact close
	return support.NewArtifactForBlob(priv, blobaccess.ForData(mime, data), priv)
}

func (n *namespaceHandler) HasArtifact(vers string) (bool, error) {
	list, err := n.ListTags()
	if err != nil {
		return false, err
	}
	for _, e := range list {
		if e == vers {
			return true, nil
		}
	}
	return false, nil
}

func (n *namespaceContainer) AddArtifact(artifact cpi.Artifact, tags ...string) (access blobaccess.BlobAccess, err error) {
	tag := "latest"
	if len(tags) > 0 {
		tag = tags[0]
	}
	ref, err := ParseRef(n.impl.GetNamespace(), tag)
	if err != nil {
		return nil, err
	}
	dst, err := ref.NewImageDestination(dummyContext, nil)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	blob, err := Convert(artifact, n.blobs, dst)
	if err != nil {
		return nil, err
	}
	err = dst.Commit(dummyContext, nil)
	if err != nil {
		return nil, err
	}

	return blob, nil
}

func (n *namespaceContainer) AddTags(digest digest.Digest, tags ...string) error {
	if ok, _ := artdesc.IsDigest(digest.String()); ok {
		return errors.ErrNotSupported("image access by digest")
	}

	src := n.impl.GetNamespace() + ":" + digest.String()

	if pattern.MatchString(digest.String()) {
		// this definitely no digest, but the library expects it this way
		src = digest.String()
	}

	for _, tag := range tags {
		opts := client.ImageTagOptions{Source: src, Target: n.impl.GetNamespace() + ":" + tag}
		_, err := n.repo.client.ImageTag(dummyContext, opts)
		if err != nil {
			return fmt.Errorf("failed to add image tag: %w", err)
		}
	}

	return nil
}

func (n *namespaceContainer) NewArtifact(i support.NamespaceAccessImpl, art ...cpi.Artifact) (cpi.ArtifactAccess, error) {
	if n.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	var m cpi.Artifact
	if len(art) == 0 || art[0] == nil {
		m = artdesc.NewManifest()
	} else {
		m = art[0]
		if !m.IsValid() {
			m = artdesc.NewManifest()
		}
	}
	return support.NewArtifact(i, m)
}
