package comparch_test

import (
	"encoding/json"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/digester/digesters/blob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/testhelper"
	"github.com/open-component-model/ocm/pkg/env"
	"github.com/open-component-model/ocm/pkg/finalizer"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha256"
	. "github.com/open-component-model/ocm/pkg/testutils"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
)

const (
	TEST_FILEPATH     = "testfilepath"
	TAR_COMPARCH      = "testdata/common"
	DIR_COMPARCH      = "testdata/directory"
	RESOURCE_NAME     = "test"
	COMPONENT_NAME    = "example.com/root"
	COMPONENT_VERSION = "1.0.0"
)

var _ = Describe("Repository", func() {

	It("marshal/unmarshal simple", func() {
		octx := ocm.DefaultContext()
		spec := Must(comparch.NewRepositorySpec(accessobj.ACC_READONLY, TEST_FILEPATH))
		data := Must(json.Marshal(spec))
		Expect(string(data)).To(Equal("{\"type\":\"ComponentArchive\",\"filePath\":\"testfilepath\",\"accessMode\":1}"))
		_ = Must(octx.RepositorySpecForConfig(data, runtime.DefaultJSONEncoding)).(*comparch.RepositorySpec)
		// spec will not equal r as the filesystem cannot be serialized
	})

	It("component archive with resource stored as tar", func() {
		// this is the typical use case
		octx := ocm.DefaultContext()
		spec := Must(comparch.NewRepositorySpec(accessobj.ACC_READONLY, TAR_COMPARCH))
		repo := Must(spec.Repository(octx, nil))
		defer Close(repo, "repo")
		cv := Must(repo.LookupComponentVersion(COMPONENT_NAME, COMPONENT_VERSION))
		defer Close(cv, "compvers")
		res := Must(cv.GetResourcesByName(RESOURCE_NAME))
		acc := Must(res[0].AccessMethod())
		defer Close(acc, "method")
		bytesA := Must(acc.Get())

		bytesB := Must(vfs.ReadFile(osfs.New(), filepath.Join(TAR_COMPARCH, "blobs", "sha256.3ed99e50092c619823e2c07941c175ea2452f1455f570c55510586b387ec2ff2")))
		Expect(bytesA).To(Equal(bytesB))
	})

	It("component archive with a resource stored in a directory", func() {
		octx := ocm.DefaultContext()
		spec := Must(comparch.NewRepositorySpec(accessobj.ACC_READONLY, DIR_COMPARCH))
		repo := Must(spec.Repository(octx, nil))
		defer Close(repo)
		cv := Must(repo.LookupComponentVersion(COMPONENT_NAME, COMPONENT_VERSION))
		defer Close(cv)
		res := Must(cv.GetResourcesByName(RESOURCE_NAME))
		acc := Must(res[0].AccessMethod())
		defer Close(acc)
		data := Must(acc.Reader())
		defer Close(data)

		mfs := memoryfs.New()
		_, _, err := tarutils.ExtractTarToFsWithInfo(mfs, data)
		Expect(err).ToNot(HaveOccurred())
		bufferA := Must(vfs.ReadFile(mfs, "testfile"))
		bufferB := Must(vfs.ReadFile(osfs.New(), filepath.Join(DIR_COMPARCH, "blobs", "root", "testfile")))
		Expect(bufferA).To(Equal(bufferB))
	})

	It("creates component archive", func() {
		octx := ocm.DefaultContext()
		memfs := memoryfs.New()

		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize)

		arch := Must(comparch.Create(octx, accessobj.ACC_WRITABLE, "test", 0o0700, accessio.PathFileSystem(memfs)))
		finalize.Close(arch, "comparch)")

		arch.SetName("acme.org/test")
		arch.SetVersion("v1.0.1")

		MustBeSuccessful(arch.SetResourceBlob(compdesc.NewResourceMeta("blob", resourcetypes.PLAIN_TEXT, metav1.LocalRelation),
			blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))

		res := Must(arch.GetResourcesByName("blob"))
		Expect(res[0].Meta().Digest).To(DeepEqual(&metav1.DigestSpec{
			HashAlgorithm:          sha256.Algorithm,
			NormalisationAlgorithm: blob.GenericBlobDigestV1,
			Value:                  D_TESTDATA,
		}))

		MustBeSuccessful(finalize.Finalize())

		arch = Must(comparch.Open(octx, accessobj.ACC_WRITABLE, "test", 0o0700, accessio.PathFileSystem(memfs)))
		finalize.Close(arch, "comparch)")

		res = Must(arch.GetResourcesByName("blob"))
		Expect(res[0].Meta().Digest).To(DeepEqual(&metav1.DigestSpec{
			HashAlgorithm:          sha256.Algorithm,
			NormalisationAlgorithm: blob.GenericBlobDigestV1,
			Value:                  D_TESTDATA,
		}))
	})

	It("closing a resource before actually reading it", func() {
		octx := ocm.DefaultContext()
		spec := Must(comparch.NewRepositorySpec(accessobj.ACC_READONLY, TAR_COMPARCH))
		repo := Must(spec.Repository(octx, nil))
		defer Close(repo)
		cv := Must(repo.LookupComponentVersion(COMPONENT_NAME, COMPONENT_VERSION))
		defer Close(cv)
		res := Must(cv.GetResourcesByName(RESOURCE_NAME))
		acc := Must(res[0].AccessMethod())
		defer Close(acc)
	})

	It("modifies component archive from spec", func() {
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize, "finalizer")

		env := env.NewEnvironment(env.ModifiableTestData())
		finalize.With(env.Cleanup)

		octx := env.OCMContext()
		spec := Must(comparch.NewRepositorySpec(accessobj.ACC_WRITABLE, TAR_COMPARCH, accessio.PathFileSystem(env)))
		repo := Must(spec.Repository(octx, nil))
		finalize.Close(repo, "repo")
		cv := Must(repo.LookupComponentVersion(COMPONENT_NAME, COMPONENT_VERSION))
		finalize.Close(cv, "cv")
		cv.GetDescriptor().Provider.Name = "modified provider"
		MustBeSuccessful(finalize.Finalize())

		spec = Must(comparch.NewRepositorySpec(accessobj.ACC_READONLY, TAR_COMPARCH, accessio.PathFileSystem(env)))
		repo = Must(spec.Repository(octx, nil))
		finalize.Close(repo, "repo")
		cv = Must(repo.LookupComponentVersion(COMPONENT_NAME, COMPONENT_VERSION))
		finalize.Close(cv, "cv")
		Expect(cv.GetDescriptor().Provider.Name).To(Equal(metav1.ProviderName("modified provider")))
	})

	It("component archive from spec with New/AddVersion", func() {
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize)

		env := env.NewEnvironment(env.ModifiableTestData())
		finalize.With(env.Cleanup)

		octx := env.OCMContext()
		spec := Must(comparch.NewRepositorySpec(accessobj.ACC_WRITABLE, TAR_COMPARCH, accessio.PathFileSystem(env)))
		repo := Must(spec.Repository(octx, nil))
		finalize.Close(repo, "repo")
		comp := Must(repo.LookupComponent(COMPONENT_NAME))
		finalize.Close(comp, "component")
		cv := Must(comp.NewVersion(COMPONENT_VERSION, true))
		finalize.Close(cv, "compvers")

		MustBeSuccessful(cv.SetResourceBlob(compdesc.NewResourceMeta("blob", resourcetypes.PLAIN_TEXT, metav1.LocalRelation),
			blobaccess.ForString(mime.MIME_TEXT, S_OTHERDATA), "", nil))

		MustBeSuccessful(comp.AddVersion(cv))

		MustBeSuccessful(finalize.Finalize())

		arch := Must(comparch.Open(octx, accessobj.ACC_READONLY, TAR_COMPARCH, 0o0700, accessio.PathFileSystem(env)))
		finalize.Close(arch, "comparch)")

		res := Must(arch.GetResourcesByName("blob"))
		Expect(res[0].Meta().Digest).To(DeepEqual(&metav1.DigestSpec{
			HashAlgorithm:          sha256.Algorithm,
			NormalisationAlgorithm: blob.GenericBlobDigestV1,
			Value:                  D_OTHERDATA,
		}))
	})

	It("handle multiple lookups", func() {
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize)

		env := env.NewEnvironment(env.ModifiableTestData())
		finalize.With(env.Cleanup)

		octx := env.OCMContext()
		spec := Must(comparch.NewRepositorySpec(accessobj.ACC_WRITABLE, TAR_COMPARCH, accessio.PathFileSystem(env)))
		repo := Must(spec.Repository(octx, nil))
		finalize.Close(repo, "repo")
		cv1 := Must(repo.LookupComponentVersion(COMPONENT_NAME, COMPONENT_VERSION))
		finalize.Close(cv1, "version1")

		cv2 := Must(repo.LookupComponentVersion(COMPONENT_NAME, COMPONENT_VERSION))

		MustBeSuccessful(cv2.Close())

		MustBeSuccessful(cv1.SetResourceBlob(compdesc.NewResourceMeta("blob", resourcetypes.PLAIN_TEXT, metav1.LocalRelation),
			blobaccess.ForString(mime.MIME_TEXT, S_OTHERDATA), "", nil))

		MustBeSuccessful(finalize.Finalize())

		arch := Must(comparch.Open(octx, accessobj.ACC_READONLY, TAR_COMPARCH, 0o0700, accessio.PathFileSystem(env)))
		finalize.Close(arch, "comparch)")

		res := Must(arch.GetResourcesByName("blob"))
		Expect(res[0].Meta().Digest).To(DeepEqual(&metav1.DigestSpec{
			HashAlgorithm:          sha256.Algorithm,
			NormalisationAlgorithm: blob.GenericBlobDigestV1,
			Value:                  D_OTHERDATA,
		}))
	})
})
