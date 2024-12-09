package accessio_test

import (
	"bytes"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/utils/accessio"
)

func CheckEOF(r io.Reader, err error) {
	var (
		buf [20]byte
		n   int
	)

	if err == nil {
		n, err = r.Read(buf[:])
		ExpectWithOffset(1, n).To(Equal(0))
	}
	ExpectWithOffset(1, err).To(Equal(io.EOF))
}

var _ = Describe("ChunkedReader", func() {
	in := "12345678901234567890"
	var buf *bytes.Buffer
	var chunked *accessio.ChunkedReader

	BeforeEach(func() {
		buf = bytes.NewBuffer([]byte(in))
	})

	Context("max preread", func() {
		BeforeEach(func() {
			chunked = accessio.NewChunkedReader(buf, 5)
		})

		It("reports EOF and splits reader", func() {
			var buf [30]byte
			cnt := 0

			n, err := chunked.Read(buf[:])
			Expect(n).To(Equal(5))
			Expect(string(buf[:n])).To(Equal(in[cnt : cnt+n]))
			CheckEOF(chunked, err)
			Expect(chunked.ChunkDone()).To(Equal(true))
			cnt += n
			Expect(chunked.Next()).To(Equal(true))

			for i := 0; i < 3; i++ {
				n, err := chunked.Read(buf[:])
				Expect(n).To(Equal(5))
				Expect(string(buf[:n])).To(Equal(in[cnt : cnt+n]))
				CheckEOF(chunked, err)
				Expect(chunked.ChunkDone()).To(Equal(true))
				cnt += n

				Expect(chunked.Next()).To(Equal(i != 2))
			}
			Expect(chunked.Next()).To(Equal(false))
		})

		It("keeps reporting EOF", func() {
			var buf [30]byte
			cnt := 0

			n, err := chunked.Read(buf[:])
			Expect(n).To(Equal(5))
			Expect(string(buf[:n])).To(Equal(in[cnt : cnt+n]))
			CheckEOF(chunked, err)
			cnt += n

			n, err = chunked.Read(buf[:])
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(io.EOF))

			Expect(chunked.ChunkDone()).To(Equal(true))
			Expect(chunked.Next()).To(Equal(true))

			for i := 0; i < 3; i++ {
				n, err := chunked.Read(buf[:])
				Expect(n).To(Equal(5))
				Expect(string(buf[:n])).To(Equal(in[cnt : cnt+n]))
				CheckEOF(chunked, err)
				Expect(chunked.ChunkDone()).To(Equal(true))
				cnt += n

				Expect(chunked.Next()).To(Equal(i != 2))
			}
			Expect(chunked.Next()).To(Equal(false))
		})
	})

	Context("complete", func() {
		BeforeEach(func() {
			chunked = accessio.NewChunkedReader(buf, 100)
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

	Context("non-matching chunk size", func() {
		BeforeEach(func() {
			chunked = accessio.NewChunkedReader(buf, 15)
		})
		It("reports EOF and Next with non-matching size", func() {
			var buf [20]byte
			n, err := chunked.Read(buf[:])
			Expect(n).To(Equal(15))
			CheckEOF(chunked, err)
			Expect(string(buf[:n])).To(Equal(in[:15]))
			Expect(chunked.ChunkDone()).To(Equal(true))
			Expect(chunked.Next()).To(Equal(true))

			n, err = chunked.Read(buf[:])
			Expect(n).To(Equal(5))
			Expect(string(buf[:n])).To(Equal(in[15:20]))
			CheckEOF(chunked, err)
			Expect(chunked.ChunkDone()).To(Equal(true))
			Expect(chunked.Next()).To(Equal(false))

			n, err = chunked.Read(buf[:])
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(io.EOF))
		})
	})

	Context("chunk size matches read size", func() {
		BeforeEach(func() {
			chunked = accessio.NewChunkedReader(buf, 20)
		})

		It("reports EOF with matched size", func() {
			var buf [20]byte
			n, err := chunked.Read(buf[:])
			Expect(n).To(Equal(20))
			Expect(string(buf[:n])).To(Equal(in))
			CheckEOF(chunked, err)
			Expect(chunked.ChunkDone()).To(Equal(true))
			Expect(chunked.Next()).To(Equal(false))

			n, err = chunked.Read(buf[:])
			Expect(n).To(Equal(0))
			Expect(err).To(Equal(io.EOF))
		})
	})

	Context("split", func() {
		BeforeEach(func() {
			chunked = accessio.NewChunkedReader(buf, 5)
		})

		It("reports EOF and splits reader", func() {
			var buf [30]byte
			cnt := 0

			n, err := chunked.Read(buf[:])
			Expect(n).To(Equal(5))
			Expect(string(buf[:n])).To(Equal(in[cnt : cnt+n]))
			CheckEOF(chunked, err)
			Expect(chunked.ChunkDone()).To(Equal(true))
			cnt += n
			Expect(chunked.Next()).To(Equal(true))

			for i := 0; i < 3; i++ {
				n, err := chunked.Read(buf[:])
				Expect(string(buf[:n])).To(Equal(in[cnt : cnt+n]))
				Expect(n).To(Equal(5))
				CheckEOF(chunked, err)
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
