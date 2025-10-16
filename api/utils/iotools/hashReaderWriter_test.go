package iotools_test

import (
	"bytes"
	"crypto"
	"io"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/utils/iotools"
)

var _ = Describe("Hash Reader Writer tests", func() {
	It("Ensure interface implementation", func() {
		var _ io.Reader = &iotools.HashReader{}
		var _ io.Reader = (*iotools.HashReader)(nil)
		var _ io.Reader = new(iotools.HashReader)

		var _ io.Writer = &iotools.HashWriter{}
		var _ io.Writer = (*iotools.HashWriter)(nil)
		var _ io.Writer = new(iotools.HashWriter)
	})

	It("test HashWriter", func() {
		s := "Hello Hash!"
		var b bytes.Buffer
		hw := iotools.NewHashWriter(io.Writer(&b))
		hw.Write([]byte(s))
		hashes := hw.Hashes()
		Expect(b.String()).To(Equal(s))
		Expect(hashes.GetBytes(0)).To(BeNil())
		b.Reset()

		w := io.Writer(&b)
		hw = iotools.NewHashWriter(w, crypto.SHA1)
		hw.Write([]byte(s))
		hashes = hw.Hashes()
		Expect(b.String()).To(Equal(s))
		Expect(hashes.GetBytes(0)).To(BeNil())
		Expect(hashes.GetString(crypto.SHA1)).To(Equal("5c075ed604db0adc524edd3516e8f0258ca6e58d"))
		b.Reset()

		hw = iotools.NewHashWriter(io.Writer(&b), crypto.SHA1, crypto.MD5)
		hw.Write([]byte(s))
		hashes = hw.Hashes()
		Expect(b.String()).To(Equal(s))
		Expect(hashes.GetBytes(0)).To(BeNil())
		Expect(hashes.GetString(crypto.MD5)).To(Equal("c10e8df2e378a1584359b0e546cf0149"))
		Expect(hashes.GetString(crypto.SHA1)).To(Equal("5c075ed604db0adc524edd3516e8f0258ca6e58d"))
	})

	It("test HashReader", func() {
		s := "Hello Hash!"
		hr := iotools.NewHashReader(strings.NewReader(s))
		buf := make([]byte, len(s))
		hr.Read(buf)
		hashes := hr.Hashes()
		Expect(hashes.GetBytes(0)).To(BeNil())
		Expect(string(buf)).To(Equal(s))

		hr = iotools.NewHashReader(strings.NewReader(s), crypto.SHA1)
		hr.Read(buf)
		hashes = hr.Hashes()
		Expect(hashes.GetBytes(0)).To(BeNil())
		Expect(hashes.GetString(crypto.SHA1)).To(Equal("5c075ed604db0adc524edd3516e8f0258ca6e58d"))

		hr = iotools.NewHashReader(strings.NewReader(s), crypto.SHA1)
		cnt, err := hr.CalcHashes()
		hashes = hr.Hashes()
		Expect(err).To(BeNil())
		Expect(cnt).To(Equal(int64(len(s))))
		Expect(hashes.GetBytes(0)).To(BeNil())
		Expect(hashes.GetString(crypto.SHA1)).To(Equal("5c075ed604db0adc524edd3516e8f0258ca6e58d"))

		hr = iotools.NewHashReader(strings.NewReader(s), crypto.SHA1, crypto.MD5)
		hr.Read(buf)
		hashes = hr.Hashes()
		Expect(hashes.GetBytes(crypto.SHA256)).To(BeNil())
		Expect(hashes.GetString(crypto.MD5)).To(Equal("c10e8df2e378a1584359b0e546cf0149"))
		Expect(hashes.GetString(crypto.SHA1)).To(Equal("5c075ed604db0adc524edd3516e8f0258ca6e58d"))
	})
})
