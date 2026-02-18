package httptimeoutattr_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/httptimeoutattr"
	"ocm.software/ocm/api/utils/runtime"
)

var _ = Describe("httptimeout attribute", func() {
	var ctx datacontext.Context
	attr := httptimeoutattr.AttributeType{}
	enc := runtime.DefaultJSONEncoding

	BeforeEach(func() {
		ctx = datacontext.New(nil)
	})

	Context("get and set", func() {
		It("defaults to DefaultTimeout", func() {
			Expect(httptimeoutattr.Get(ctx)).To(Equal(httptimeoutattr.DefaultTimeout))
		})

		It("sets and retrieves duration", func() {
			Expect(httptimeoutattr.Set(ctx, 5*time.Minute)).To(Succeed())
			Expect(httptimeoutattr.Get(ctx)).To(Equal(5 * time.Minute))

			Expect(httptimeoutattr.Set(ctx, 2*time.Minute)).To(Succeed())
			Expect(httptimeoutattr.Get(ctx)).To(Equal(2 * time.Minute))
		})
	})

	Context("encoding values to JSON", func() {
		DescribeTable("encodes valid input",
			func(input interface{}, expected string) {
				data, err := attr.Encode(input, enc)
				Expect(err).To(Succeed())
				Expect(string(data)).To(Equal(expected))
			},
			Entry("time.Duration 30s to JSON string", 30*time.Second, `"30s"`),
			Entry("duration string 5m to JSON string", "5m", `"5m"`),
		)

		DescribeTable("rejects invalid input",
			func(input interface{}, errSubstring string) {
				_, err := attr.Encode(input, enc)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(errSubstring))
			},
			Entry("non-parseable string like notaduration", "notaduration", "invalid duration string"),
			Entry("string with unknown unit like 1Gb", "1Gb", "invalid duration string"),
			Entry("unsupported type like bool", true, "duration or duration string required"),
		)
	})

	Context("decoding values from JSON", func() {
		DescribeTable("decodes valid JSON input",
			func(input string, expected time.Duration) {
				val, err := attr.Decode([]byte(input), enc)
				Expect(err).To(Succeed())
				Expect(val).To(Equal(expected))
			},
			Entry("duration string 30s", `"30s"`, 30*time.Second),
			Entry("duration string 5m", `"5m"`, 5*time.Minute),
			Entry("duration string 1h", `"1h"`, 1*time.Hour),
			Entry("nanoseconds number 300000000000 as 5m", `300000000000`, 5*time.Minute),
		)

		DescribeTable("rejects invalid JSON input",
			func(input string, errSubstring string) {
				_, err := attr.Decode([]byte(input), enc)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(errSubstring))
			},
			Entry("non-parseable string like notaduration", `"notaduration"`, "invalid timeout value"),
			Entry("string with unknown unit like 1Gb", `"1Gb"`, "invalid timeout value"),
			Entry("digit-only string like 300000000000", `"300000000000"`, "invalid timeout value"),
			Entry("JSON boolean true", `true`, "must be a duration string or nanoseconds number"),
		)
	})
})
