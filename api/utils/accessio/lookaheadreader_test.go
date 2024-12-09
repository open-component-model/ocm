package accessio_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/utils/accessio"
)

var _ = Describe("LookAheadReader", func() {
	in := "12345678901234567890"
	var buf *bytes.Buffer
	var lookup *accessio.LookAheadReader

	BeforeEach(func() {
		buf = bytes.NewBuffer([]byte(in))
	})

	Context("read", func() {
		BeforeEach(func() {
			lookup = accessio.NewLookAheadReader(buf)
		})

		It("reads all", func() {
			var buf [30]byte

			n, err := lookup.Read(buf[:])
			Expect(n).To(Equal(20))
			Expect(string(buf[:n])).To(Equal(in))
			CheckEOF(lookup, err)
		})

		It("looksup", func() {
			var buf [30]byte

			n, err := lookup.Read(buf[:2])
			Expect(err).To(BeNil())
			Expect(n).To(Equal(2))
			Expect(string(buf[:n])).To(Equal(in[:2]))

			n, err = lookup.LookAhead(buf[:5])
			Expect(err).To(BeNil())
			Expect(n).To(Equal(5))
			Expect(string(buf[:n])).To(Equal(in[2:7]))

			n, err = lookup.Read(buf[:3])
			Expect(err).To(BeNil())
			Expect(n).To(Equal(3))
			Expect(string(buf[:n])).To(Equal(in[2:5]))

			n, err = lookup.Read(buf[:])
			Expect(err).To(BeNil())
			Expect(n).To(Equal(15))
			Expect(string(buf[:n])).To(Equal(in[5:20]))
		})

		It("interferring lookup", func() {
			var buf [30]byte

			n, err := lookup.Read(buf[:2])
			Expect(err).To(BeNil())
			Expect(n).To(Equal(2))
			Expect(string(buf[:n])).To(Equal(in[:2]))

			n, err = lookup.LookAhead(buf[:5])
			Expect(err).To(BeNil())
			Expect(n).To(Equal(5))
			Expect(string(buf[:n])).To(Equal(in[2:7]))

			n, err = lookup.Read(buf[:3])
			Expect(err).To(BeNil())
			Expect(n).To(Equal(3))
			Expect(string(buf[:n])).To(Equal(in[2:5]))

			n, err = lookup.LookAhead(buf[:5])
			Expect(err).To(BeNil())
			Expect(n).To(Equal(5))
			Expect(string(buf[:n])).To(Equal(in[5:10]))

			n, err = lookup.Read(buf[:])
			Expect(err).To(BeNil())
			Expect(n).To(Equal(15))
			Expect(string(buf[:n])).To(Equal(in[5:20]))
		})

	})
})
