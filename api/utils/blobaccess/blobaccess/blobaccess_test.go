package blobaccess_test

import (
	"os"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
	"ocm.software/ocm/api/utils/blobaccess/file"
	"ocm.software/ocm/api/utils/mime"
)

var _ = Describe("blob access ref counting", func() {
	It("handles ref count less access", func() {
		blob := blobaccess.ForString(mime.MIME_TEXT, "test")
		dup := Must(blob.Dup())
		MustBeSuccessful(blob.Close())
		MustBeSuccessful(blob.Close())
		Expect(dup.Get()).To(Equal([]byte("test")))
		MustBeSuccessful(dup.Close())
	})

	It("handles ref count ", func() {
		blob := bpi.NewBlobAccessForBase(bpi.BaseAccessForDataAccess(mime.MIME_TEXT, blobaccess.DataAccessForString("test")))
		dup := Must(blob.Dup())
		MustBeSuccessful(blob.Close())
		ExpectError(blob.Close()).To(Equal(blobaccess.ErrClosed))
		ExpectError(blob.Get()).To(Equal(blobaccess.ErrClosed))
		ExpectError(blob.Reader()).To(Equal(blobaccess.ErrClosed))
		Expect(dup.Get()).To(Equal([]byte("test")))
		ExpectError(dup.Digest().String()).To(Equal("sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"))
		ExpectError(dup.Size()).To(Equal(int64(4)))
		MustBeSuccessful(dup.Close())
	})

	It("releases temp file", func() {
		temp := Must(os.CreateTemp("", "testfile*"))
		path := temp.Name()
		temp.Close()
		blob := file.BlobAccessForTemporaryFilePath(mime.MIME_TEXT, path)

		Expect(vfs.FileExists(osfs.OsFs, path)).To(BeTrue())

		dup := Must(blob.Dup())
		MustBeSuccessful(blob.Close())
		Expect(vfs.FileExists(osfs.OsFs, path)).To(BeTrue())
		MustBeSuccessful(dup.Close())
		Expect(vfs.FileExists(osfs.OsFs, path)).To(BeFalse())
	})
})
