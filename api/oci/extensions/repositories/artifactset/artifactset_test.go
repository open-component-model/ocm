package artifactset_test

import (
	"archive/tar"
	"compress/gzip"
	"io"

	"github.com/mandelsoft/goutils/finalizer"
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	. "ocm.software/ocm/api/oci/extensions/repositories/artifactset/testhelper"
	. "ocm.software/ocm/api/oci/extensions/repositories/ctf/testhelper"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
)

func defaultManifestFill(a *artifactset.ArtifactSet) {
	var finalize finalizer.Finalizer
	defer Defer(finalize.Finalize)

	art := NewArtifact(a, &finalize)
	MustWithOffset(1, Calling(a.AddArtifact(art)))
}

var _ = Describe("artifact management", func() {
	var tempfs vfs.FileSystem
	var opts accessio.Options

	BeforeEach(func() {
		t, err := osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		tempfs = t
		opts, err = accessio.AccessOptions(nil, accessio.PathFileSystem(tempfs))
		Expect(err).To(Succeed())
	})

	AfterEach(func() {
		vfs.Cleanup(tempfs)
	})

	It("creates with default format", func() {
		a, err := artifactset.FormatDirectory.Create("test", opts, 0o700)
		Expect(err).To(Succeed())
		Expect(vfs.DirExists(tempfs, "test/"+artifactset.BlobsDirectoryName)).To(BeTrue())

		defaultManifestFill(a)

		Expect(a.Close()).To(Succeed())

		desc := artifactset.DescriptorFileName("")
		Expect(vfs.FileExists(tempfs, "test/"+desc)).To(BeTrue())
		Expect(vfs.FileExists(tempfs, "test/"+artifactset.OCILayouFileName)).To(Equal(desc == artifactset.OCIArtifactSetDescriptorFileName))
	})

	TestForAllFormats("instantiate filesystem artifact", func(format string) {
		opts, err := accessio.AccessOptions(&artifactset.Options{}, opts, artifactset.StructureFormat(format))
		Expect(err).To(Succeed())

		a, err := artifactset.FormatDirectory.Create("test", opts, 0o700)
		Expect(err).To(Succeed())
		Expect(vfs.DirExists(tempfs, "test/"+artifactset.BlobsDirectoryName)).To(BeTrue())

		defaultManifestFill(a)

		Expect(a.Close()).To(Succeed())

		desc := artifactset.DescriptorFileName(format)
		Expect(vfs.FileExists(tempfs, "test/"+desc)).To(BeTrue())
		Expect(vfs.FileExists(tempfs, "test/"+artifactset.OCILayouFileName)).To(Equal(desc == artifactset.OCIArtifactSetDescriptorFileName))

		infos, err := vfs.ReadDir(tempfs, "test/"+artifactset.BlobsDirectoryName)
		Expect(err).To(Succeed())
		blobs := []string{}
		for _, fi := range infos {
			blobs = append(blobs, fi.Name())
		}
		Expect(blobs).To(ContainElements(
			"sha256.3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a",
			"sha256.44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
			"sha256.810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"))
	})

	TestForAllFormats("instantiate tgz artifact", func(format string) {
		opts, err := accessio.AccessOptions(&artifactset.Options{}, opts, artifactset.StructureFormat(format))
		Expect(err).To(Succeed())

		a, err := artifactset.FormatTGZ.Create("test.tgz", opts, 0o600)
		Expect(err).To(Succeed())

		defaultManifestFill(a)

		Expect(a.Close()).To(Succeed())
		Expect(vfs.FileExists(tempfs, "test.tgz")).To(BeTrue())

		file, err := tempfs.Open("test.tgz")
		Expect(err).To(Succeed())
		defer file.Close()
		zip, err := gzip.NewReader(file)
		Expect(err).To(Succeed())
		defer zip.Close()
		tr := tar.NewReader(zip)

		files := []string{}
		for {
			header, err := tr.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				Fail(err.Error())
			}

			switch header.Typeflag {
			case tar.TypeDir:
				Expect(header.Name).To(Equal(artifactset.BlobsDirectoryName))
			case tar.TypeReg:
				files = append(files, header.Name)
			}
		}
		elems := []interface{}{
			artifactset.DescriptorFileName(format),
			"blobs/sha256.3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a",
			"blobs/sha256.44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
			"blobs/sha256.810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50",
		}
		if format == artifactset.FORMAT_OCI {
			elems = append(elems, artifactset.OCILayouFileName)
		}
		Expect(files).To(ContainElements(elems))
	})

	TestForAllFormats("instantiate tgz artifact for open file object", func(format string) {
		file, err := vfs.TempFile(opts.GetPathFileSystem(), "", "*.tgz")
		Expect(err).To(Succeed())
		defer file.Close()

		opts, err := accessio.AccessOptions(&artifactset.Options{FormatVersion: format}, opts, accessio.File(file))
		Expect(err).To(Succeed())

		a, err := artifactset.FormatTGZ.Create("", opts, 0o600)
		Expect(err).To(Succeed())

		defaultManifestFill(a)

		Expect(a.Close()).To(Succeed())
		Expect(vfs.FileExists(tempfs, file.Name())).To(BeTrue())

		_, err = file.Seek(0, io.SeekStart)
		Expect(err).To(Succeed())
		zip, err := gzip.NewReader(file)
		Expect(err).To(Succeed())
		defer zip.Close()
		tr := tar.NewReader(zip)

		files := []string{}
		for {
			header, err := tr.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				Fail(err.Error())
			}

			switch header.Typeflag {
			case tar.TypeDir:
				Expect(header.Name).To(Equal(artifactset.BlobsDirectoryName))
			case tar.TypeReg:
				files = append(files, header.Name)
			}
		}
		elems := []interface{}{
			artifactset.DescriptorFileName(format),
			"blobs/sha256.3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a",
			"blobs/sha256.44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
			"blobs/sha256.810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50",
		}
		if format == artifactset.FORMAT_OCI {
			elems = append(elems, artifactset.OCILayouFileName)
		}
		Expect(files).To(ContainElements(elems))
	})

	Context("manifest", func() {
		TestForAllFormats("read from filesystem artifact", func(format string) {
			opts := Must(accessio.AccessOptions(&artifactset.Options{}, opts, artifactset.StructureFormat(format)))

			a := Must(artifactset.FormatDirectory.Create("test", opts, 0o700))
			Expect(vfs.DirExists(tempfs, "test/"+artifactset.BlobsDirectoryName)).To(BeTrue())
			defaultManifestFill(a)
			Expect(a.Close()).To(Succeed())

			a = Must(artifactset.FormatDirectory.Open(accessobj.ACC_READONLY, "test", opts))
			defer Close(a, "artefactset")
			Expect(len(a.GetIndex().Manifests)).To(Equal(1))
			art := Must(a.GetArtifact(a.GetIndex().Manifests[0].Digest.String()))
			defer Close(art, "artefact")
			Expect(art.IsManifest()).To(BeTrue())
			blob := Must(art.GetBlob("sha256:810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"))
			Expect(blob.Get()).To(Equal([]byte("testdata")))
			Expect(blob.MimeType()).To(Equal(mime.MIME_OCTET))
		})

		TestForAllFormats("read from tgz artifact", func(format string) {
			opts := Must(accessio.AccessOptions(&artifactset.Options{}, opts, artifactset.StructureFormat(format)))

			a := Must(artifactset.FormatTGZ.Create("test.tgz", opts, 0o700))
			defaultManifestFill(a)
			Expect(a.Close()).To(Succeed())

			a = Must(artifactset.Open(accessobj.ACC_READONLY, "test.tgz", 0, opts))
			defer Close(a, "artefactset")
			Expect(len(a.GetIndex().Manifests)).To(Equal(1))
			art := Must(a.GetArtifact(a.GetIndex().Manifests[0].Digest.String()))
			defer Close(art, "artefact")
			Expect(art.IsManifest()).To(BeTrue())
			blob := Must(art.GetBlob("sha256:810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"))
			defer Close(blob, "blob")
			Expect(blob.Get()).To(Equal([]byte("testdata")))
			Expect(blob.MimeType()).To(Equal(mime.MIME_OCTET))
		})
	})
})
