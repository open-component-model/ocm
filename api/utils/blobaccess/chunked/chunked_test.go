package chunked_test

import (
	"bytes"
	"io"

	"github.com/mandelsoft/goutils/finalizer"
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/utils/blobaccess/chunked"
)

var _ = Describe("Chunked Blobs", func() {
	var blobData = []byte("a1a2a3a4a5a6a7a8a9a0b1b2b3b4b5b6b7b8b9b0")

	var src chunked.ChunkedBlobSource
	var fs vfs.FileSystem

	Context("blobs", func() {
		BeforeEach(func() {
			fs = memoryfs.New()
		})

		AfterEach(func() {
			vfs.Cleanup(fs)
		})

		It("small blobs", func() {
			src = chunked.New(bytes.NewBuffer(blobData), 50, fs)
			n := 0
			buf := bytes.NewBuffer(nil)

			var finalize finalizer.Finalizer
			defer finalize.Finalize()
			for {
				b := Must(src.Next())
				if b == nil {
					break
				}
				n++
				finalize.Close(b)
				r := Must(b.Reader())
				_, err := io.Copy(buf, r)
				r.Close()
				MustBeSuccessful(err)
			}
			Expect(n).To(Equal(1))
			Expect(buf.String()).To(Equal(string(blobData)))

			MustBeSuccessful(finalize.Finalize())
			list := Must(vfs.ReadDir(fs, "/"))
			Expect(len(list)).To(Equal(0))
		})

		It("matching blobs", func() {
			src = chunked.New(bytes.NewBuffer(blobData), 40, fs)
			n := 0
			buf := bytes.NewBuffer(nil)

			var finalize finalizer.Finalizer
			defer finalize.Finalize()
			for {
				b := Must(src.Next())
				if b == nil {
					break
				}
				n++
				finalize.Close(b)
				r := Must(b.Reader())
				_, err := io.Copy(buf, r)
				r.Close()
				MustBeSuccessful(err)
			}
			Expect(n).To(Equal(1))
			Expect(buf.String()).To(Equal(string(blobData)))

			MustBeSuccessful(finalize.Finalize())
			list := Must(vfs.ReadDir(fs, "/"))
			Expect(len(list)).To(Equal(0))
		})

		It("large blobs", func() {
			src = chunked.New(bytes.NewBuffer(blobData), 18, fs)
			n := 0
			buf := bytes.NewBuffer(nil)

			var finalize finalizer.Finalizer
			defer finalize.Finalize()
			for {
				b := Must(src.Next())
				if b == nil {
					break
				}
				n++
				finalize.Close(b)
				r := Must(b.Reader())
				_, err := io.Copy(buf, r)
				r.Close()
				MustBeSuccessful(err)
			}
			Expect(n).To(Equal(3))
			Expect(buf.String()).To(Equal(string(blobData)))

			MustBeSuccessful(finalize.Finalize())
			list := Must(vfs.ReadDir(fs, "/"))
			Expect(len(list)).To(Equal(0))
		})
	})
})
