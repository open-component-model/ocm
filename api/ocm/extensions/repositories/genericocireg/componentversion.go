package genericocireg

import (
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/set"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/cpi/repocpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localociblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/relativeociref"
	"ocm.software/ocm/api/ocm/extensions/attrs/compatattr"
	ocihdlr "ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/oci"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/errkind"
	"ocm.software/ocm/api/utils/mime"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/refmgmt"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/api/version"
)

// newComponentVersionAccess creates a component access for the artifact access, if this fails the artifact access is closed.
func newComponentVersionAccess(mode accessobj.AccessMode, comp *componentAccessImpl, version string, access oci.ArtifactAccess, persistent bool) (*repocpi.ComponentVersionAccessInfo, error) {
	c, err := newComponentVersionContainer(mode, comp, version, access)
	if err != nil {
		return nil, err
	}
	return &repocpi.ComponentVersionAccessInfo{c, true, persistent}, nil
}

////////////////////////////////////////////////////////////////////////////////

type ComponentVersionContainer struct {
	bridge repocpi.ComponentVersionAccessBridge

	comp     *componentAccessImpl
	version  string
	access   oci.ArtifactAccess
	manifest oci.ManifestAccess
	state    accessobj.State
}

var _ repocpi.ComponentVersionAccessImpl = (*ComponentVersionContainer)(nil)

func newComponentVersionContainer(mode accessobj.AccessMode, comp *componentAccessImpl, version string, access oci.ArtifactAccess) (*ComponentVersionContainer, error) {
	m := access.ManifestAccess()
	if m == nil {
		return nil, errors.ErrInvalid("artifact type")
	}
	state, err := NewState(mode, comp.name, version, m, compatattr.Get(comp.GetContext()))
	if err != nil {
		access.Close()
		return nil, err
	}

	return &ComponentVersionContainer{
		comp:     comp,
		version:  version,
		access:   access,
		manifest: m,
		state:    state,
	}, nil
}

func (c *ComponentVersionContainer) SetBridge(impl repocpi.ComponentVersionAccessBridge) {
	c.bridge = impl
}

func (c *ComponentVersionContainer) GetParentBridge() repocpi.ComponentAccessBridge {
	return c.comp.bridge
}

func (c *ComponentVersionContainer) Close() error {
	if c.manifest == nil {
		return accessio.ErrClosed
	}
	c.manifest = nil
	return c.access.Close()
}

func (c *ComponentVersionContainer) SetReadOnly() {
	c.state.SetReadOnly()
}

func (c *ComponentVersionContainer) Check() error {
	if c.version != c.GetDescriptor().Version {
		// check if version contained '+' which has been replaced by META_SEPARATOR to create OCI compliant tag
		if replaced, _ := toTag(c.GetDescriptor().Version); replaced != c.GetDescriptor().Version && replaced == c.version {
			Logger(c.GetContext()).Warn(fmt.Sprintf(
				"checked version %q contains %q, this is discouraged and you should prefer the original component version %q", c.version, META_SEPARATOR, c.GetDescriptor().Version))
			return nil
		}
		return errors.ErrInvalid("component version", c.GetDescriptor().Version)
	}
	if c.comp.name != c.GetDescriptor().Name {
		return errors.ErrInvalid("component name", c.GetDescriptor().Name)
	}
	return nil
}

func (c *ComponentVersionContainer) Repository() cpi.Repository {
	return c.comp.repo.nonref
}

func (c *ComponentVersionContainer) GetContext() cpi.Context {
	return c.comp.GetContext()
}

func (c *ComponentVersionContainer) IsReadOnly() bool {
	return c.state.IsReadOnly()
}

func (c *ComponentVersionContainer) IsClosed() bool {
	return c.manifest == nil
}

func (c *ComponentVersionContainer) AccessMethod(a cpi.AccessSpec, cv refmgmt.ExtendedAllocatable) (cpi.AccessMethod, error) {
	accessSpec, err := c.comp.GetContext().AccessSpecForSpec(a)
	if err != nil {
		return nil, err
	}

	switch a.GetKind() {
	case localblob.Type:
		return newLocalBlobAccessMethod(accessSpec.(*localblob.AccessSpec), c.comp.namespace, c.access, cv)
	case localociblob.Type:
		return newLocalOCIBlobAccessMethod(accessSpec.(*localblob.AccessSpec), c.comp.namespace, c.access, cv)
	case relativeociref.Type:
		m, err := ociartifact.NewMethod(c.GetContext(), a, accessSpec.(*relativeociref.AccessSpec).Reference, c.comp.repo.ocirepo)
		if err == nil {
			impl := accspeccpi.GetAccessMethodImplementation(m).(ociartifact.AccessMethodImpl)
			cv.BeforeCleanup(refmgmt.CleanupHandlerFunc(impl.Cache))
		}
		return m, err
	}

	return nil, errors.ErrNotSupported(errkind.KIND_ACCESSMETHOD, a.GetType(), "oci registry")
}

func (c *ComponentVersionContainer) SetDescriptor(cd *compdesc.ComponentDescriptor) (bool, error) {
	cur := c.GetDescriptor()
	*cur = *cd
	return c.Update()
}

type LayerAnnotations []ArtifactInfo

type ArtifactInfo struct {
	// Kind specifies whether the artifact is a source, resource or a label
	Kind     string          `json:"kind"`
	Identity metav1.Identity `json:"identity"`
}

const (
	OCM_COMPONENTVERSION = "software.ocm.componentversion"
	OCM_CREATOR          = "software.ocm.creator"
	OCM_ARTIFACT         = "software.ocm.artifact"
	ARTKIND_RESOURCE     = "resource"
	ARTKIND_SOURCE       = "source"
)

func (c *ComponentVersionContainer) Update() (bool, error) {
	logger := Logger(c.GetContext()).WithValues("cv", common.NewNameVersion(c.comp.name, c.version))
	err := c.Check()
	if err != nil {
		return false, fmt.Errorf("check failed: %w", err)
	}

	if c.state.HasChanged() {
		layerAnnotations := map[int]LayerAnnotations{}

		logger.Debug("update component version")
		desc := c.GetDescriptor()
		layers := set.Set[int]{}
		for i := range c.manifest.GetDescriptor().Layers {
			layers.Add(i)
		}
		for i, r := range desc.Resources {
			s, list, err := c.evalLayer(r.Access)
			if err != nil {
				return false, fmt.Errorf("failed resource layer evaluation: %w", err)
			}
			for _, l := range list {
				layerAnnotations[l] = append(layerAnnotations[l], ArtifactInfo{
					Kind:     ARTKIND_RESOURCE,
					Identity: r.GetIdentity(desc.Resources),
				})
				layers.Delete(l)
			}
			if s != r.Access {
				desc.Resources[i].Access = s
			}
		}
		for i, r := range desc.Sources {
			s, list, err := c.evalLayer(r.Access)
			if err != nil {
				return false, fmt.Errorf("failed source layer evaluation: %w", err)
			}
			for _, l := range list {
				layerAnnotations[l] = append(layerAnnotations[l], ArtifactInfo{
					Kind:     ARTKIND_SOURCE,
					Identity: r.GetIdentity(desc.Sources),
				})
				layers.Delete(l)
			}
			if s != r.Access {
				desc.Sources[i].Access = s
			}
		}
		m := c.manifest.GetDescriptor()

		if m.Annotations == nil {
			m.Annotations = map[string]string{}
		}
		m.Annotations[OCM_COMPONENTVERSION] = common.VersionedElementKey(c.bridge).String()
		m.Annotations[OCM_CREATOR] = "OCM Go Library " + version.Current()

		for layer, info := range layerAnnotations {
			data, err := runtime.DefaultJSONEncoding.Marshal(info)
			if err != nil {
				return false, err
			}
			if m.Layers[layer].Annotations == nil {
				m.Layers[layer].Annotations = map[string]string{}
			}
			m.Layers[layer].Annotations[OCM_ARTIFACT] = string(data)
		}
		i := len(m.Layers) - 1

		for i > 0 {
			if layers.Contains(i) {
				logger.Debug("removing unused layer", "layer", i)
				m.Layers = append(m.Layers[:i], m.Layers[i+1:]...)
			}
			i--
		}
		if _, err := c.state.Update(); err != nil {
			return false, fmt.Errorf("failed to update state: %w", err)
		}

		logger.Debug("add oci artifact")
		tag, err := toTag(c.version)
		if err != nil {
			return false, err
		}
		if _, err := c.comp.namespace.AddArtifact(c.manifest, tag); err != nil {
			return false, fmt.Errorf("unable to add artifact: %w", err)
		}
		return true, nil
	}

	return false, nil
}

func (c *ComponentVersionContainer) evalLayer(s compdesc.AccessSpec) (compdesc.AccessSpec, []int, error) {
	var (
		d         *artdesc.Descriptor
		layernums []int
	)

	spec, err := c.GetContext().AccessSpecForSpec(s)
	if err != nil {
		return s, nil, err
	}
	if a, ok := spec.(*localblob.AccessSpec); ok {
		if ok, _ := artdesc.IsDigest(a.LocalReference); !ok {
			return s, nil, errors.ErrInvalid("digest", a.LocalReference)
		}
		refs := strings.Split(a.LocalReference, ",")
		media := a.GetMimeType()
		if len(refs) > 1 {
			media = mime.MIME_OCTET
		}
		for _, ref := range refs {
			d = &artdesc.Descriptor{Digest: digest.Digest(strings.TrimSpace(ref)), MediaType: media}
			// find layer
			layers := c.manifest.GetDescriptor().Layers
			maxLen := len(layers) - 1
			found := false
			for i := maxLen; i > 0; i-- { // layer 0 is the component descriptor
				l := layers[i]
				if l.Digest == d.Digest {
					layernums = append(layernums, i)
					found = true
					break
				}
			}
			if !found {
				return s, nil, fmt.Errorf("resource access %s: no layer found for local blob %s[%s]", spec.Describe(c.GetContext()), d.Digest, d.MediaType)
			}
		}
	}
	return s, layernums, nil
}

func (c *ComponentVersionContainer) GetDescriptor() *compdesc.ComponentDescriptor {
	return c.state.GetState().(*compdesc.ComponentDescriptor)
}

func (c *ComponentVersionContainer) GetBlob(name string) (cpi.DataAccess, error) {
	return c.manifest.GetBlob(digest.Digest(name))
}

func (c *ComponentVersionContainer) GetStorageContext() cpi.StorageContext {
	return ocihdlr.New(c.comp.GetName(), c.Repository(), c.comp.repo.ocirepo.GetSpecification().GetKind(), c.comp.repo.ocirepo, c.comp.namespace, c.manifest)
}

func blobAccessForChunk(blob blobaccess.BlobAccess, fs vfs.FileSystem, r io.Reader, limit int64) (cpi.BlobAccess, bool, error) {
	f, err := blobaccess.NewTempFile("", "chunk-*", fs)
	if err != nil {
		return nil, true, err
	}
	written, err := io.CopyN(f.Writer(), r, limit)
	if err != nil && !errors.Is(err, io.EOF) {
		f.Close()
		return nil, false, err
	}
	if written <= 0 {
		f.Close()
		return nil, false, nil
	}
	return f.AsBlob(blob.MimeType()), written == limit, nil
}

func (c *ComponentVersionContainer) AddBlob(blob cpi.BlobAccess, refName string, global cpi.AccessSpec) (cpi.AccessSpec, error) {
	if blob == nil {
		return nil, errors.New("a resource has to be defined")
	}

	fs := vfsattr.Get(c.GetContext())
	size := blob.Size()
	limit := c.comp.repo.blobLimit
	var refs []string
	if limit > 0 && size != blobaccess.BLOB_UNKNOWN_SIZE && size > limit {
		reader, err := blob.Reader()
		if err != nil {
			return nil, err
		}
		defer reader.Close()

		var b blobaccess.BlobAccess
		cont := true
		for cont {
			b, cont, err = blobAccessForChunk(blob, fs, reader, limit)
			if err != nil {
				return nil, err
			}
			if b != nil {
				err = c.addLayer(b, &refs)
				b.Close()
				if err != nil {
					return nil, err
				}
			}
		}
	} else {
		err := c.addLayer(blob, &refs)
		if err != nil {
			return nil, err
		}
	}
	return localblob.New(strings.Join(refs, ","), refName, blob.MimeType(), global), nil
}

func (c *ComponentVersionContainer) addLayer(blob cpi.BlobAccess, refs *[]string) error {
	err := c.manifest.AddBlob(blob)
	if err != nil {
		return err
	}

	err = c.manifest.Modify(func(manifest *artdesc.Manifest) error {
		return ocihdlr.AssureLayerLocked(blob, manifest)
	})
	if err != nil {
		return err
	}
	*refs = append(*refs, blob.Digest().String())
	return nil
}

// AssureGlobalRef provides a global manifest for a local OCI Artifact.
func (c *ComponentVersionContainer) AssureGlobalRef(d digest.Digest, url, name string) (cpi.AccessSpec, error) {
	blob, err := c.manifest.GetBlob(d)
	if err != nil {
		return nil, err
	}
	var namespace oci.NamespaceAccess
	var version string
	var tag string
	if name == "" {
		namespace = c.comp.namespace
	} else {
		i := strings.LastIndex(name, ":")
		if i > 0 {
			version = name[i+1:]
			name = name[:i]
			tag = version
		}
		namespace, err = c.comp.repo.ocirepo.LookupNamespace(name)
		if err != nil {
			return nil, err
		}
	}
	set, err := artifactset.OpenFromBlob(accessobj.ACC_READONLY, blob)
	if err != nil {
		return nil, err
	}
	defer set.Close()
	digest := set.GetMain()
	if version == "" {
		version = digest.String()
	}
	art, err := set.GetArtifact(digest.String())
	if err != nil {
		return nil, err
	}
	err = artifactset.TransferArtifact(art, namespace, oci.AsTags(tag)...)
	if err != nil {
		return nil, err
	}

	ref := path.Join(url+namespace.GetNamespace()) + ":" + version

	global := ociartifact.New(ref)
	return global, nil
}
