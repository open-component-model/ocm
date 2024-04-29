package iotools_test

import (
	"bytes"
	"crypto"
	"io"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/iotools"
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
		hr := iotools.NewHashWriter(io.Writer(&b))
		hr.Write([]byte(s))
		Expect(b.String()).To(Equal(s))
		Expect(hr.GetBytes(0)).To(BeNil())
		b.Reset()

		w := io.Writer(&b)
		hr = iotools.NewHashWriter(w, crypto.SHA1)
		hr.Write([]byte(s))
		Expect(b.String()).To(Equal(s))
		Expect(hr.GetBytes(0)).To(BeNil())
		Expect(hr.GetString(crypto.SHA1)).To(Equal("5c075ed604db0adc524edd3516e8f0258ca6e58d"))
		b.Reset()

		hr = iotools.NewHashWriter(io.Writer(&b), crypto.SHA1, crypto.MD5)
		hr.Write([]byte(s))
		Expect(b.String()).To(Equal(s))
		Expect(hr.GetBytes(0)).To(BeNil())
		Expect(hr.GetString(crypto.MD5)).To(Equal("c10e8df2e378a1584359b0e546cf0149"))
		Expect(hr.GetString(crypto.SHA1)).To(Equal("5c075ed604db0adc524edd3516e8f0258ca6e58d"))
	})

	It("test HashReader", func() {
		s := "Hello Hash!"
		hr := iotools.NewHashReader(strings.NewReader(s))
		buf := make([]byte, len(s))
		hr.Read(buf)
		Expect(hr.GetBytes(0)).To(BeNil())
		Expect(string(buf)).To(Equal(s))

		hr = iotools.NewHashReader(strings.NewReader(s), crypto.SHA1)
		hr.Read(buf)
		Expect(hr.GetBytes(0)).To(BeNil())
		Expect(hr.GetString(crypto.SHA1)).To(Equal("5c075ed604db0adc524edd3516e8f0258ca6e58d"))

		hr = iotools.NewHashReader(strings.NewReader(s), crypto.SHA1)
		cnt, err := hr.CalcHashes()
		Expect(err).To(BeNil())
		Expect(cnt).To(Equal(int64(len(s))))
		Expect(hr.GetBytes(0)).To(BeNil())
		Expect(hr.GetString(crypto.SHA1)).To(Equal("5c075ed604db0adc524edd3516e8f0258ca6e58d"))

		hr = iotools.NewHashReader(strings.NewReader(s), crypto.SHA1, crypto.MD5)
		hr.Read(buf)
		Expect(hr.GetBytes(crypto.SHA256)).To(BeNil())
		Expect(hr.GetString(crypto.MD5)).To(Equal("c10e8df2e378a1584359b0e546cf0149"))
		Expect(hr.GetString(crypto.MD5)).To(Equal("c10e8df2e378a1584359b0e546cf0149"))
		Expect(hr.GetString(crypto.SHA1)).To(Equal("5c075ed604db0adc524edd3516e8f0258ca6e58d"))
		Expect(hr.GetString(crypto.SHA1)).To(Equal("5c075ed604db0adc524edd3516e8f0258ca6e58d"))
	})
})
