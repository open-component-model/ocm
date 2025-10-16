package builder

import (
	"fmt"
	"slices"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/file"
)

const T_OCIARTIFACTSET = "artifact set"

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) ArtifactSet(path string, fmt accessio.FileFormat, f ...func()) {
	r, err := artifactset.Open(accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, path, 0o777, fmt, accessio.PathFileSystem(b.FileSystem()))
	b.failOn(err)

	b.configure(&ociNamespace{NamespaceAccess: r, kind: T_OCIARTIFACTSET, annofunc: func(name, value string) {
		r.Annotate(name, value)
	}}, f)
}

func (b *Builder) ArtifactSetBlob(main string, f ...func()) {
	b.expect(b.blob, T_BLOBACCESS)
	arch, err := vfs.TempFile(b.FileSystem(), "", "artifact-*.tgz")
	b.failOn(err)
	b.FileSystem().Remove(arch.Name())

	r, err := artifactset.Open(accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, arch.Name(), 0o777, accessio.FormatTGZ, accessio.PathFileSystem(b.FileSystem()))
	b.failOn(err)

	b.configure(&ociNamespace{NamespaceAccess: &artifactSink{b, arch.Name(), main, r, ""}, kind: T_OCIARTIFACTSET, annofunc: func(name, value string) {
		r.Annotate(name, value)
	}}, append(f, func() {
		b.Annotation(artifactset.MAINARTIFACT_ANNOTATION, main)
	}))
}

type artifactSink struct {
	builder *Builder
	arch    string
	main    string
	oci.NamespaceAccess
	mime string
}

func (s *artifactSink) AddArtifact(art cpi.Artifact, tags ...string) (blobaccess.BlobAccess, error) {
	if slices.Contains(tags, s.main) {
		s.mime = art.Artifact().MimeType()
	}
	return s.NamespaceAccess.AddArtifact(art, tags...)
}

func (s *artifactSink) Close() error {
	if s.mime == "" {
		s.NamespaceAccess.Close()
		return fmt.Errorf("main artifact not defined")
	}
	*s.builder.blob = blobaccess.ForTemporaryFilePath(artifactset.MediaType(s.mime), s.arch, file.WithFileSystem(s.builder.FileSystem()))
	return s.NamespaceAccess.Close()
}
