package iotools_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/iotools"
)

var _ = Describe("digests.go tests", func() {
	It("DecodeBase64ToHex", func() {
		hex, err := iotools.DecodeBase64ToHex("04cXMnFlKzgudf//lH/VqGUtFkplvGv0BCmPREEJsVYTJrxyiBFlsOiZIrjPENBkHWPnK6kOG53VTiqtsILNgw==")
		Expect(err).To(BeNil())
		Expect(hex).To(Equal("d387173271652b382e75ffff947fd5a8652d164a65bc6bf404298f444109b1561326bc72881165b0e89922b8cf10d0641d63e72ba90e1b9dd54e2aadb082cd83"))

		hex, err = iotools.DecodeBase64ToHex("sha512-04cXMnFlKzgudf//lH/VqGUtFkplvGv0BCmPREEJsVYTJrxyiBFlsOiZIrjPENBkHWPnK6kOG53VTiqtsILNgw==")
		Expect(err).To(BeNil())
		Expect(hex).To(Equal("d387173271652b382e75ffff947fd5a8652d164a65bc6bf404298f444109b1561326bc72881165b0e89922b8cf10d0641d63e72ba90e1b9dd54e2aadb082cd83"))

		hex, err = iotools.DecodeBase64ToHex("SHA512-04cXMnFlKzgudf//lH/VqGUtFkplvGv0BCmPREEJsVYTJrxyiBFlsOiZIrjPENBkHWPnK6kOG53VTiqtsILNgw==")
		Expect(err).To(BeNil())
		Expect(hex).To(Equal("d387173271652b382e75ffff947fd5a8652d164a65bc6bf404298f444109b1561326bc72881165b0e89922b8cf10d0641d63e72ba90e1b9dd54e2aadb082cd83"))
	})
})
