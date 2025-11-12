package maxworkersattr_test

import (
	goruntime "runtime"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/datacontext"
	me "ocm.software/ocm/api/ocm/extensions/attrs/maxworkersattr"
	"ocm.software/ocm/api/utils/runtime"
)

var _ = Describe("maxworkers attribute", func() {
	var ctx datacontext.Context

	BeforeEach(func() {
		ctx = datacontext.New(nil)
	})

	Context("resolution logic", func() {
		It("defaults to single worker", func() {
			val, err := me.Get(ctx)
			Expect(err).To(Succeed())
			Expect(val).To(Equal(me.SingleWorker))
		})

		It("uses explicit uint attribute", func() {
			Expect(me.Set(ctx, uint(8))).To(Succeed())
			val, err := me.Get(ctx)
			Expect(err).To(Succeed())
			Expect(val).To(Equal(uint(8)))
		})

		It("uses explicit string number attribute", func() {
			Expect(me.Set(ctx, "6")).To(Succeed())
			val, err := me.Get(ctx)
			Expect(err).To(Succeed())
			Expect(val).To(Equal(uint(6)))
		})

		It("uses explicit int attribute", func() {
			Expect(me.Set(ctx, 3)).To(Succeed())
			val, err := me.Get(ctx)
			Expect(err).To(Succeed())
			Expect(val).To(Equal(uint(3)))
		})

		It("uses 'auto' attribute to detect CPU count", func() {
			Expect(me.Set(ctx, "auto")).To(Succeed())
			val, err := me.Get(ctx)
			Expect(err).To(Succeed())
			Expect(val).To(Equal(uint(goruntime.NumCPU())))
		})

		It("treats 0 as single worker", func() {
			Expect(me.Set(ctx, 0)).To(Succeed())
			val, err := me.Get(ctx)
			Expect(err).To(Succeed())
			Expect(val).To(Equal(me.SingleWorker))
		})

		It("prefers attribute over environment variable", func() {
			GinkgoT().Setenv(me.TransferWorkersEnvVar, "99")
			Expect(me.Set(ctx, 7)).To(Succeed())
			val, err := me.Get(ctx)
			Expect(err).To(Succeed())
			Expect(val).To(Equal(uint(7)))
		})

		It("uses environment variable when no attribute is set", func() {
			GinkgoT().Setenv(me.TransferWorkersEnvVar, "5")
			val, err := me.Get(ctx)
			Expect(err).To(Succeed())
			Expect(val).To(Equal(uint(5)))
		})

		It("supports 'auto' in environment variable", func() {
			GinkgoT().Setenv(me.TransferWorkersEnvVar, "auto")
			val, err := me.Get(ctx)
			Expect(err).To(Succeed())
			Expect(val).To(Equal(uint(goruntime.NumCPU())))
		})

		It("rejects invalid string attribute", func() {
			Expect(me.Set(ctx, "foo")).ToNot(Succeed())
		})

		It("rejects negative int attribute", func() {
			Expect(me.Set(ctx, -2)).ToNot(Succeed())
		})
	})

	Context("encoding and decoding", func() {
		It("encodes and decodes uint correctly", func() {
			data, err := me.AttributeType{}.Encode(uint(4), runtime.DefaultJSONEncoding)
			Expect(err).To(Succeed())

			val, err := me.AttributeType{}.Decode(data, runtime.DefaultJSONEncoding)
			Expect(err).To(Succeed())
			Expect(val).To(Equal(uint(4)))
		})

		It("encodes and decodes string 'auto' correctly", func() {
			data, err := me.AttributeType{}.Encode("auto", runtime.DefaultJSONEncoding)
			Expect(err).To(Succeed())

			val, err := me.AttributeType{}.Decode(data, runtime.DefaultJSONEncoding)
			Expect(err).To(Succeed())
			Expect(val).To(Equal("auto"))
		})

		It("decodes string number correctly", func() {
			val, err := me.AttributeType{}.Decode([]byte(`"8"`), runtime.DefaultJSONEncoding)
			Expect(err).To(Succeed())
			Expect(val).To(Equal(uint(8)))
		})

		It("rejects invalid string decode", func() {
			_, err := me.AttributeType{}.Decode([]byte(`"invalid"`), runtime.DefaultJSONEncoding)
			Expect(err).To(HaveOccurred())
		})

		It("rejects unsupported encode type", func() {
			_, err := me.AttributeType{}.Encode([]string{"x"}, runtime.DefaultJSONEncoding)
			Expect(err).To(HaveOccurred())
		})

		It("rejects negative int encode", func() {
			_, err := me.AttributeType{}.Encode(-1, runtime.DefaultJSONEncoding)
			Expect(err).To(HaveOccurred())
		})

		It("encodes and decodes int correctly", func() {
			data, err := me.AttributeType{}.Encode(10, runtime.DefaultJSONEncoding)
			Expect(err).To(Succeed())

			val, err := me.AttributeType{}.Decode(data, runtime.DefaultJSONEncoding)
			Expect(err).To(Succeed())
			Expect(val).To(Equal(uint(10)))
		})

		It("decodes plain uint JSON number", func() {
			val, err := me.AttributeType{}.Decode([]byte(strconv.Itoa(12)), runtime.DefaultJSONEncoding)
			Expect(err).To(Succeed())
			Expect(val).To(Equal(uint(12)))
		})
	})
})
