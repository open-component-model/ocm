package accessio_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	common "ocm.software/ocm/api/utils/misc"
)

var _ = Describe("cache management", func() {
	var tempfs vfs.FileSystem
	var cache accessio.BlobCache
	var source accessio.BlobCache

	var td1_digest digest.Digest
	var td1_size int64

	BeforeEach(func() {
		t, err := osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		tempfs = t
		local, err := accessio.NewDefaultBlobCache(t)
		Expect(err).To(Succeed())

		source, err = accessio.NewDefaultBlobCache()
		Expect(err).To(Succeed())

		td1_size, td1_digest, err = source.AddData(blobaccess.DataAccessForData([]byte("testdata")))
		Expect(err).To(Succeed())

		cache, err = accessio.CachedAccess(source, nil, local)
		Expect(err).To(Succeed())

		_ = td1_size
	})

	AfterEach(func() {
		cache.Unref()
		source.Unref()
		vfs.Cleanup(tempfs)
	})

	It("blob copied to cache", func() {
		Expect(vfs.FileExists(tempfs, common.DigestToFileName(td1_digest))).To(BeFalse())
		_, data, err := cache.GetBlobData(td1_digest)
		Expect(err).To(Succeed())
		Expect(vfs.FileExists(tempfs, common.DigestToFileName(td1_digest))).To(BeFalse())
		Expect(data.Get()).To(Equal([]byte("testdata")))
		Expect(vfs.FileExists(tempfs, common.DigestToFileName(td1_digest))).To(BeTrue())
	})
})
