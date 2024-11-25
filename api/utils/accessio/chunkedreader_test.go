package accessio_test

import (
	"bytes"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/utils/accessio"
)

var _ = FDescribe("Test Environment", func() {
	in := "12345678901234567890"
	var buf *bytes.Buffer
	var chunked *accessio.ChunkedReader

	BeforeEach(func() {
		buf = bytes.NewBuffer([]byte(in))
	})

	Context("complete", func() {
		BeforeEach(func() {
			chunked = accessio.NewChunkedReader(buf, 100, 2)
		})

		It("reports EOF", func() {
			var buf [30]byte
			n, err := chunked.Read(buf[:])
			Expect(n).To(Equal(20))
			Expect(err).To(BeNil())
			Expect(string(buf[:n])).To(Equal(in))
			Expect(chunked.ChunkDone()).To(Equal(false))
			Expect(chunked.Next()).To(Equal(false))

			n, err = chunked.Read(buf[:])
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(io.EOF))
			Expect(chunked.ChunkDone()).To(Equal(true))
			Expect(chunked.Next()).To(Equal(false))

			n, err = chunked.Read(buf[:])
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(io.EOF))
		})

		It("reports EOF with matched size", func() {
			var buf [20]byte
			n, err := chunked.Read(buf[:])
			Expect(n).To(Equal(20))
			Expect(err).To(BeNil())
			Expect(string(buf[:n])).To(Equal(in))
			Expect(chunked.ChunkDone()).To(Equal(false))
			Expect(chunked.Next()).To(Equal(false))

			n, err = chunked.Read(buf[:])
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(io.EOF))
			Expect(chunked.ChunkDone()).To(Equal(true))
			Expect(chunked.Next()).To(Equal(false))

			n, err = chunked.Read(buf[:])
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(io.EOF))
		})
	})

	Context("chunk size matches read size", func() {
		BeforeEach(func() {
			chunked = accessio.NewChunkedReader(buf, 20, 2)
		})

		It("reports EOF with matched size", func() {
			var buf [20]byte
			n, err := chunked.Read(buf[:])
			Expect(n).To(Equal(20))
			Expect(err).To(Equal(io.EOF))
			Expect(string(buf[:n])).To(Equal(in))
			Expect(chunked.ChunkDone()).To(Equal(true))
			Expect(chunked.Next()).To(Equal(false))

			n, err = chunked.Read(buf[:])
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(io.EOF))
		})
	})

	Context("split", func() {
		BeforeEach(func() {
			chunked = accessio.NewChunkedReader(buf, 5, 2)
		})

		It("reports EOF and splits reader", func() {
			var buf [30]byte
			cnt := 0

			n, err := chunked.Read(buf[:])
			Expect(n).To(Equal(5))
			Expect(err).To(Equal(io.EOF))
			Expect(string(buf[:n])).To(Equal(in[cnt : cnt+n]))
			Expect(chunked.ChunkDone()).To(Equal(true))
			cnt += n
			Expect(chunked.Next()).To(Equal(true))

			for i := 0; i < 3; i++ {
				n, err := chunked.Read(buf[:])
				Expect(n).To(Equal(2))
				Expect(err).To(BeNil())
				Expect(string(buf[:n])).To(Equal(in[cnt : cnt+n]))
				Expect(chunked.ChunkDone()).To(Equal(false))
				cnt += n

				n, err = chunked.Read(buf[:])
				Expect(n).To(Equal(3))
				Expect(err).To(Equal(io.EOF))
				Expect(string(buf[:n])).To(Equal(in[cnt : cnt+n]))
				Expect(chunked.ChunkDone()).To(Equal(true))
				cnt += n

				Expect(chunked.Next()).To(Equal(i != 2))
			}
			Expect(chunked.Next()).To(Equal(false))
		})
	})
})

func check(chunked *accessio.ChunkedReader, n int, buf []byte, exp int, data string, done, next bool) {
	ExpectWithOffset(1, n).To(Equal(exp))
	ExpectWithOffset(1, string(buf[:n])).To(Equal(data))

	ExpectWithOffset(1, chunked.ChunkDone()).To(Equal(done))
	ExpectWithOffset(1, chunked.Next()).To(Equal(next))

}
