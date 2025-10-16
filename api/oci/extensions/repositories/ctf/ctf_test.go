package ctf_test

import (
	"archive/tar"
	"compress/gzip"
	"io"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opencontainers/go-digest"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/oci/extensions/repositories/ctf"
	. "ocm.software/ocm/api/oci/extensions/repositories/ctf/testhelper"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess"
	ocmlog "ocm.software/ocm/api/utils/logging"
	"ocm.software/ocm/api/utils/refmgmt"
)

var _ = Describe("ctf management", func() {
	var tempfs vfs.FileSystem

	var spec *ctf.RepositorySpec

	ocmlog.Context().AddRule(logging.NewConditionRule(logging.TraceLevel, refmgmt.ALLOC_REALM))

	BeforeEach(func() {
		t, err := osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		tempfs = t

		spec, err = ctf.NewRepositorySpec(accessobj.ACC_CREATE, "test", accessio.PathFileSystem(tempfs), accessobj.FormatDirectory)
		Expect(err).To(Succeed())
	})

	AfterEach(func() {
		vfs.Cleanup(tempfs)
	})

	It("instantiate filesystem ctf", func() {
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize)

		r := Must(ctf.FormatDirectory.Create(oci.DefaultContext(), "test", &spec.StandardOptions, 0o700))
		finalize.Close(r)
		Expect(vfs.DirExists(tempfs, "test/"+ctf.BlobsDirectoryName)).To(BeTrue())

		sub := finalize.Nested()
		n := Must(r.LookupNamespace("mandelsoft/test"))
		sub.Close(n)
		DefaultManifestFill(n)
		Expect(sub.Finalize()).To(Succeed())

		Expect(r.ExistsArtifact("mandelsoft/test", TAG)).To(BeTrue())

		art := Must(r.LookupArtifact("mandelsoft/test", TAG))
		Close(art, "art")

		Expect(finalize.Finalize()).To(Succeed())

		Expect(vfs.FileExists(tempfs, "test/"+ctf.ArtifactIndexFileName)).To(BeTrue())

		infos, err := vfs.ReadDir(tempfs, "test/"+artifactset.BlobsDirectoryName)
		Expect(err).To(Succeed())
		blobs := []string{}
		for _, fi := range infos {
			blobs = append(blobs, fi.Name())
		}
		Expect(blobs).To(ContainElements(
			"sha256."+DIGEST_MANIFEST,
			"sha256."+DIGEST_CONFIG,
			"sha256."+DIGEST_LAYER))
	})

	It("instantiate filesystem ctf", func() {
		r, err := spec.Repository(cpi.DefaultContext(), nil)
		Expect(err).To(Succeed())
		Expect(vfs.DirExists(tempfs, "test/"+ctf.BlobsDirectoryName)).To(BeTrue())

		n, err := r.LookupNamespace("mandelsoft/test")
		Expect(err).To(Succeed())
		DefaultManifestFill(n)

		Expect(n.Close()).To(Succeed())
		Expect(r.Close()).To(Succeed())
		Expect(vfs.FileExists(tempfs, "test/"+ctf.ArtifactIndexFileName)).To(BeTrue())

		infos, err := vfs.ReadDir(tempfs, "test/"+artifactset.BlobsDirectoryName)
		Expect(err).To(Succeed())
		blobs := []string{}
		for _, fi := range infos {
			blobs = append(blobs, fi.Name())
		}
		Expect(blobs).To(ContainElements(
			"sha256."+DIGEST_MANIFEST,
			"sha256."+DIGEST_CONFIG,
			"sha256."+DIGEST_LAYER))
	})

	It("instantiate tgz artifact", func() {
		ctf.FormatTGZ.ApplyOption(&spec.StandardOptions)
		spec.FilePath = "test.tgz"
		r, err := spec.Repository(cpi.DefaultContext(), nil)
		Expect(err).To(Succeed())

		n, err := r.LookupNamespace("mandelsoft/test")
		Expect(err).To(Succeed())
		DefaultManifestFill(n)

		Expect(n.Close()).To(Succeed())
		Expect(r.Close()).To(Succeed())
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
		Expect(files).To(ContainElements(
			ctf.ArtifactIndexFileName,
			"blobs/sha256."+DIGEST_MANIFEST,
			"blobs/sha256."+DIGEST_CONFIG,
			"blobs/sha256."+DIGEST_LAYER))
	})

	Context("manifest", func() {
		It("read from filesystem ctf", func() {
			r, err := spec.Repository(cpi.DefaultContext(), nil)
			Expect(err).To(Succeed())
			Expect(vfs.DirExists(tempfs, "test/"+ctf.BlobsDirectoryName)).To(BeTrue())
			n, err := r.LookupNamespace("mandelsoft/test")
			Expect(err).To(Succeed())
			DefaultManifestFill(n)
			Expect(n.Close()).To(Succeed())
			Expect(r.Close()).To(Succeed())

			r, err = ctf.Open(cpi.DefaultContext(), accessobj.ACC_READONLY, "test", 0, accessio.PathFileSystem(tempfs))
			Expect(err).To(Succeed())
			defer r.Close()

			n, err = r.LookupNamespace("mandelsoft/test")
			Expect(err).To(Succeed())

			art, err := n.GetArtifact("sha256:" + DIGEST_MANIFEST)
			Expect(err).To(Succeed())
			CheckArtifact(art)
			art, err = n.GetArtifact(TAG)
			Expect(err).To(Succeed())
			b, err := art.GetDescriptor().ToBlobAccess()
			Expect(err).To(Succeed())
			Expect(b.Digest()).To(Equal(digest.Digest("sha256:" + DIGEST_MANIFEST)))

			_, err = n.GetArtifact("dummy")
			Expect(err).To(Equal(errors.ErrNotFound(cpi.KIND_OCIARTIFACT, "dummy", "mandelsoft/test")))

			Expect(n.AddBlob(blobaccess.ForString("", "dummy"))).To(Equal(accessobj.ErrReadOnly))

			n, err = r.LookupNamespace("mandelsoft/other")
			Expect(err).To(Succeed())
			_, err = n.GetArtifact("sha256:" + DIGEST_MANIFEST)
			Expect(err).To(Equal(errors.ErrNotFound(cpi.KIND_OCIARTIFACT, "sha256:"+DIGEST_MANIFEST, "mandelsoft/other")))
		})
	})
})
