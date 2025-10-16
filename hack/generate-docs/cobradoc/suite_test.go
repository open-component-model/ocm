package cobradoc_test

import (
	"bytes"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/hack/generate-docs/cobradoc"
)

var _ = Describe("spaces", func() {
	It("handles spaces at the end of a write followed by a newline", func() {
		var buf bytes.Buffer
		w := cobradoc.OmitTrailingSpaces(&buf)

		Expect(w.Write([]byte("a   "))).To(Equal(4))
		Expect(w.Write([]byte("\nbc   \n"))).To(Equal(7))

		Expect(buf.String()).To(Equal("a\nbc\n"))
	})
	It("handles spaces at the end of a write followed by spaces and a newline", func() {
		var buf bytes.Buffer
		w := cobradoc.OmitTrailingSpaces(&buf)

		Expect(w.Write([]byte("a   "))).To(Equal(4))
		Expect(w.Write([]byte("  \nbc   \n"))).To(Equal(9))

		Expect(buf.String()).To(Equal("a\nbc\n"))
	})
	It("handles spaces at the end of a write followed by spaces followed by spaces and a newline", func() {
		var buf bytes.Buffer
		w := cobradoc.OmitTrailingSpaces(&buf)

		Expect(w.Write([]byte("a   "))).To(Equal(4))
		Expect(w.Write([]byte("   "))).To(Equal(3))
		Expect(w.Write([]byte("  \nbc   \n"))).To(Equal(9))

		Expect(buf.String()).To(Equal("a\nbc\n"))
	})
	It("handles spaces at the end of a write followed by intermediate followed by spaces and a newline", func() {
		var buf bytes.Buffer
		w := cobradoc.OmitTrailingSpaces(&buf)

		Expect(w.Write([]byte("a   "))).To(Equal(4))
		Expect(w.Write([]byte(" x "))).To(Equal(3))
		Expect(w.Write([]byte("  \nbc   \n"))).To(Equal(9))

		Expect(buf.String()).To(Equal("a    x\nbc\n"))
	})
	It("handles spaces at the end of a write not followed by a newline", func() {
		var buf bytes.Buffer
		w := cobradoc.OmitTrailingSpaces(&buf)

		Expect(w.Write([]byte("a   "))).To(Equal(4))
		Expect(w.Write([]byte("bc   \n"))).To(Equal(6))

		Expect(buf.String()).To(Equal("a   bc\n"))
	})
	It("handles spaces at the end", func() {
		var buf bytes.Buffer
		w := cobradoc.OmitTrailingSpaces(&buf)

		Expect(w.Write([]byte("a   "))).To(Equal(4))
		Expect(buf.String()).To(Equal("a"))
	})
})

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cobra Doc")
}
