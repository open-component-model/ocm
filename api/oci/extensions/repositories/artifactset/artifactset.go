package artifactset

import (
	"maps"
	"slices"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/oci/annotations"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/oci/cpi/support"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
)

const (
	MAINARTIFACT_ANNOTATION = annotations.MAINARTIFACT_ANNOTATION

	TAGS_ANNOTATION = annotations.TAGS_ANNOTATION
	TYPE_ANNOTATION = annotations.TYPE_ANNOTATION

	OCITAG_ANNOTATION = annotations.OCITAG_ANNOTATION
)

func RetrieveMainArtifactFromIndex(index *artdesc.Index) string {
	if index.Annotations != nil {
		main := index.Annotations[MAINARTIFACT_ANNOTATION]
		if main != "" {
			return main
		}
	}
	if len(index.Manifests) == 1 {
		return index.Manifests[0].Digest.String()
	}
	return ""
}

func RetrieveMainArtifact(m map[string]string) string {
	return m[MAINARTIFACT_ANNOTATION]
}

func RetrieveTags(m map[string]string) string {
	return m[TAGS_ANNOTATION]
}

func RetrieveType(m map[string]string) string {
	return m[TYPE_ANNOTATION]
}

func RetrieveDigest(idx *artdesc.Index, ref string) digest.Digest {
	match := matcher(ref)
	for i, e := range idx.Manifests {
		if match(&idx.Manifests[i]) {
			return e.Digest
		}
	}
	return digest.Digest("")
}

////////////////////////////////////////////////////////////////////////////////

type ArtifactSet struct {
	cpi.NamespaceAccess
	container *namespaceContainer
}

func (a *ArtifactSet) Close() error { // why???
	return a.NamespaceAccess.Close()
}

func (a *ArtifactSet) GetBlobData(digest digest.Digest) (int64, blobaccess.DataAccess, error) {
	return a.container.GetBlobData(digest)
}

func (a *ArtifactSet) Annotate(name string, value string) {
	a.container.Annotate(name, value)
}

func (a *ArtifactSet) GetIndex() *artdesc.Index {
	return a.container.GetIndex()
}

func (a *ArtifactSet) GetMain() digest.Digest {
	return a.container.GetMain()
}

func (a *ArtifactSet) GetDigest(ref string) digest.Digest {
	return a.container.GetDigest(ref)
}

func (a *ArtifactSet) GetAnnotation(name string) string {
	return a.container.GetAnnotation(name)
}

func (a *ArtifactSet) HasAnnotation(name string) bool {
	return a.container.HasAnnotation(name)
}

func (a *ArtifactSet) SetMainArtifact(version string) {
	if version != "" {
		a.Annotate(MAINARTIFACT_ANNOTATION, version)
	}
}

func AsArtifactSet(ns cpi.NamespaceAccess) (*ArtifactSet, error) {
	i, err := cpi.GetNamespaceAccessImplementation(ns)
	if err != nil {
		return nil, errors.ErrInvalid()
	}
	c, err := support.GetArtifactSetContainer(i)
	if err != nil {
		return nil, errors.ErrInvalid()
	}
	if a, ok := c.(*namespaceContainer); ok {
		n, err := ns.Dup()
		if err != nil {
			return nil, err
		}
		return &ArtifactSet{
			NamespaceAccess: n,
			container:       a,
		}, nil
	}
	return nil, errors.ErrInvalid()
}

type namespaceContainer struct {
	base *FileSystemBlobAccess
	impl cpi.NamespaceAccessImpl
}

// New returns a new representation based element.
func New(acc accessobj.AccessMode, fs vfs.FileSystem, setup accessobj.Setup, closer accessobj.Closer, mode vfs.FileMode, formatVersion string) (*ArtifactSet, error) {
	return _Wrap(accessobj.NewAccessObject(NewAccessObjectInfo(formatVersion), acc, fs, setup, closer, mode))
}

func _Wrap(obj *accessobj.AccessObject, err error) (*ArtifactSet, error) {
	if err != nil {
		return nil, err
	}
	c := &namespaceContainer{
		base: NewFileSystemBlobAccess(obj),
	}
	ns, err := support.NewNamespaceAccess("", c, nil, "artifactset namespace")
	if err != nil {
		return nil, err
	}
	return &ArtifactSet{ns, c}, nil
}

func (a *namespaceContainer) SetImplementation(impl support.NamespaceAccessImpl) {
	a.impl = impl
}

func (a *namespaceContainer) Annotate(name string, value string) {
	a.base.Lock()
	defer a.base.Unlock()

	d := a.GetIndex()
	if d.Annotations == nil {
		d.Annotations = map[string]string{}
	}
	d.Annotations[name] = value
}

func (a *namespaceContainer) GetAnnotation(name string) string {
	a.base.Lock()
	defer a.base.Unlock()

	d := a.GetIndex()
	if d.Annotations == nil {
		return ""
	}
	return d.Annotations[name]
}

func (a *namespaceContainer) HasAnnotation(name string) bool {
	a.base.Lock()
	defer a.base.Unlock()

	d := a.GetIndex()
	if d.Annotations == nil {
		return false
	}
	_, ok := d.Annotations[name]
	return ok
}

////////////////////////////////////////////////////////////////////////////////
// sink

func (a *namespaceContainer) AddTags(digest digest.Digest, tags ...string) error {
	if a.IsClosed() {
		return accessio.ErrClosed
	}
	if len(tags) == 0 {
		return nil
	}

	a.base.Lock()
	defer a.base.Unlock()

	index := a.GetIndex()

	manifests := index.Manifests

	artifacts := decodeIndexManifests(manifests)

	art := artifacts[digest]
	if art == nil {
		return errors.ErrUnknown(cpi.KIND_OCIARTIFACT, digest.String())
	}

	for _, tag := range tags {
		art.addTag(tag)
	}

	isOCI := a.base.
		FileSystemBlobAccess.
		Access().
		GetInfo().
		GetDescriptorFileName() == OCIArtifactSetDescriptorFileName

	index.Manifests = encodeIndexManifests(artifacts, isOCI)

	return nil
}

type descriptorWithTags struct {
	cpi.Descriptor
	Tags map[string]struct{}
}

func (d *descriptorWithTags) addTag(tag string) {
	if tag := strings.TrimSpace(tag); tag != "" {
		d.Tags[tag] = struct{}{}
	}
}

func decodeIndexManifests(manifests []cpi.Descriptor) map[digest.Digest]*descriptorWithTags {
	out := map[digest.Digest]*descriptorWithTags{}
	for _, m := range manifests {
		if out[m.Digest] == nil {
			out[m.Digest] = &descriptorWithTags{
				Descriptor: normalizeDecodedDescriptorWithoutTags(m),
				Tags:       map[string]struct{}{},
			}
		}
		annotated := out[m.Digest]

		// OCM multi-tag annotation
		tagsFromAnnotations := strings.Split(RetrieveTags(m.Annotations), ",")
		for _, tag := range tagsFromAnnotations {
			annotated.addTag(tag)
		}

		// OCI single-tag annotation
		if tag, ok := m.Annotations[OCITAG_ANNOTATION]; ok {
			annotated.addTag(tag)
		}
	}

	return out
}

func normalizeDecodedDescriptorWithoutTags(d cpi.Descriptor) cpi.Descriptor {
	d.Annotations = maps.Clone(d.Annotations)
	delete(d.Annotations, TAGS_ANNOTATION)
	delete(d.Annotations, OCITAG_ANNOTATION)
	if len(d.Annotations) == 0 {
		d.Annotations = nil
	}
	return d
}

func encodeIndexManifests(
	descriptors map[digest.Digest]*descriptorWithTags,
	oci bool,
) []cpi.Descriptor {
	var manifests []cpi.Descriptor
	for _, desc := range descriptors {
		tags := slices.Sorted(maps.Keys(desc.Tags))
		joined := strings.Join(tags, ",")
		if !oci {
			d := desc.Descriptor
			d.Annotations = map[string]string{
				TAGS_ANNOTATION: joined,
			}
			manifests = append(manifests, d)
			continue
		}
		for _, t := range tags {
			d := desc.Descriptor
			d.Annotations = map[string]string{
				TAGS_ANNOTATION:   joined,
				OCITAG_ANNOTATION: t,
			}
			manifests = append(manifests, d)
		}
	}

	return manifests
}

////////////////////////////////////////////////////////////////////////////////
// forward

func (a *namespaceContainer) IsReadOnly() bool {
	return a.base.IsReadOnly()
}

func (a *namespaceContainer) Write(path string, mode vfs.FileMode, opts ...accessio.Option) error {
	return a.base.Write(path, mode, opts...)
}

func (a *namespaceContainer) Update() (bool, error) {
	return a.base.Update()
}

func (a *namespaceContainer) Close() error {
	return a.base.Close()
}

func (a *namespaceContainer) IsClosed() bool {
	return a.base.IsClosed()
}

// GetIndex returns the index of the included artifacts
// (image manifests and image indices)
// The manifst entries may describe dedicated tags
// to use for the dedicated artifact as annotation
// with the key TAGS_ANNOTATION.
func (a *namespaceContainer) GetIndex() *artdesc.Index {
	if a.IsReadOnly() {
		return a.base.GetState().GetOriginalState().(*artdesc.Index)
	}
	return a.base.GetState().GetState().(*artdesc.Index)
}

// GetMain returns the digest of the main artifact
// described by this artifact set.
// There might be more, if the main artifact is an index.
func (a *namespaceContainer) GetMain() digest.Digest {
	idx := a.GetIndex()
	if idx.Annotations == nil {
		return ""
	}
	v := RetrieveMainArtifact(idx.Annotations)
	return a.getDigest(v)
}

func (a *namespaceContainer) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	return a.GetIndex().GetBlobDescriptor(digest)
}

func (a *namespaceContainer) GetBlobData(digest digest.Digest) (int64, cpi.DataAccess, error) {
	return a.base.GetBlobData(digest)
}

func (a *namespaceContainer) AddBlob(blob cpi.BlobAccess) error {
	if a.IsClosed() {
		return accessio.ErrClosed
	}
	if a.IsReadOnly() {
		return accessio.ErrReadOnly
	}
	if blob == nil {
		return nil
	}
	a.base.Lock()
	defer a.base.Unlock()
	return a.base.AddBlob(blob)
}

func (a *namespaceContainer) ListTags() ([]string, error) {
	result := []string{}
	for _, a := range a.GetIndex().Manifests {
		if a.Annotations != nil {
			if tags := RetrieveTags(a.Annotations); tags != "" {
				result = append(result, strings.Split(tags, ",")...)
			}
		}
	}
	return result, nil
}

func (a *namespaceContainer) GetTags(digest digest.Digest) ([]string, error) {
	result := []string{}
	for _, a := range a.GetIndex().Manifests {
		if a.Digest == digest && a.Annotations != nil {
			if tags := RetrieveTags(a.Annotations); tags != "" {
				result = append(result, strings.Split(tags, ",")...)
			}
		}
	}
	return result, nil
}

func (a *namespaceContainer) HasArtifact(ref string) (bool, error) {
	if a.IsClosed() {
		return false, accessio.ErrClosed
	}
	a.base.Lock()
	defer a.base.Unlock()
	return a.hasArtifact(ref)
}

func (a *namespaceContainer) GetDigest(ref string) digest.Digest {
	a.base.Lock()
	defer a.base.Unlock()
	return a.getDigest(ref)
}

func (a *namespaceContainer) GetArtifact(i support.NamespaceAccessImpl, ref string) (cpi.ArtifactAccess, error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	a.base.Lock()
	defer a.base.Unlock()
	return a.getArtifact(i, ref)
}

func matcher(ref string) func(d *artdesc.Descriptor) bool {
	if ok, digest := artdesc.IsDigest(ref); ok {
		return func(desc *artdesc.Descriptor) bool {
			return desc.Digest == digest
		}
	}
	return func(d *artdesc.Descriptor) bool {
		if d.Annotations == nil {
			return false
		}
		for _, tag := range strings.Split(RetrieveTags(d.Annotations), ",") {
			if tag == ref {
				return true
			}
		}
		return false
	}
}

func (a *namespaceContainer) hasArtifact(ref string) (bool, error) {
	idx := a.GetIndex()
	match := matcher(ref)
	for i := range idx.Manifests {
		if match(&idx.Manifests[i]) {
			return true, nil
		}
	}
	return false, nil
}

func (a *namespaceContainer) getDigest(ref string) digest.Digest {
	idx := a.GetIndex()
	return RetrieveDigest(idx, ref)
}

func (a *namespaceContainer) getArtifact(impl support.NamespaceAccessImpl, ref string) (cpi.ArtifactAccess, error) {
	idx := a.GetIndex()
	match := matcher(ref)
	for i, e := range idx.Manifests {
		if match(&idx.Manifests[i]) {
			return a.base.GetArtifact(impl, e.Digest)
		}
	}
	return nil, errors.ErrNotFound(cpi.KIND_OCIARTIFACT, ref, impl.GetNamespace())
}

func (a *namespaceContainer) AnnotateArtifact(digest digest.Digest, name, value string) error {
	if a.IsClosed() {
		return accessio.ErrClosed
	}
	if a.IsReadOnly() {
		return accessio.ErrReadOnly
	}
	a.base.Lock()
	defer a.base.Unlock()
	idx := a.GetIndex()
	for i, e := range idx.Manifests {
		if e.Digest == digest {
			annos := e.Annotations
			if annos == nil {
				annos = map[string]string{}
				idx.Manifests[i].Annotations = annos
			}
			annos[name] = value
			return nil
		}
	}
	return errors.ErrUnknown(cpi.KIND_OCIARTIFACT, digest.String())
}

func (a *namespaceContainer) AddArtifact(artifact cpi.Artifact, tags ...string) (access blobaccess.BlobAccess, err error) {
	blob, err := a.AddPlatformArtifact(artifact, nil)
	if err != nil {
		return nil, err
	}
	return blob, a.AddTags(blob.Digest(), tags...)
}

func (a *namespaceContainer) AddPlatformArtifact(artifact cpi.Artifact, platform *artdesc.Platform) (access blobaccess.BlobAccess, err error) {
	if a.IsClosed() {
		return nil, blobaccess.ErrClosed
	}
	if a.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	a.base.Lock()
	defer a.base.Unlock()
	idx := a.GetIndex()
	blob, err := a.base.AddArtifactBlob(artifact)
	if err != nil {
		return nil, err
	}

	desc := cpi.Descriptor{
		Digest:      blob.Digest(),
		Size:        blob.Size(),
		URLs:        nil,
		Annotations: nil,
		Platform:    platform,
	}

	isOCI := a.base.FileSystemBlobAccess.
		Access().
		GetInfo().
		GetDescriptorFileName() == OCIArtifactSetDescriptorFileName

	if isOCI {
		desc.MediaType = blob.MimeType()
	}

	idx.Manifests = append(idx.Manifests, desc)
	return blob, nil
}

func (a *namespaceContainer) NewArtifact(i support.NamespaceAccessImpl, artifact ...cpi.Artifact) (cpi.ArtifactAccess, error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	if a.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	return support.NewArtifact(i, artifact...)
}
