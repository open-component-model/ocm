package iotools_test

import (
	"crypto"
	"encoding/base64"
	"encoding/hex"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/utils/iotools"
)

var _ = Describe("digests.go tests", func() {
	It("DecodeBase64ToHex", func() {
		hx, err := iotools.DecodeBase64ToHex("04cXMnFlKzgudf//lH/VqGUtFkplvGv0BCmPREEJsVYTJrxyiBFlsOiZIrjPENBkHWPnK6kOG53VTiqtsILNgw==")
		Expect(err).To(BeNil())
		Expect(hx).To(Equal("d387173271652b382e75ffff947fd5a8652d164a65bc6bf404298f444109b1561326bc72881165b0e89922b8cf10d0641d63e72ba90e1b9dd54e2aadb082cd83"))

		hx, err = iotools.DecodeBase64ToHex("sha512-04cXMnFlKzgudf//lH/VqGUtFkplvGv0BCmPREEJsVYTJrxyiBFlsOiZIrjPENBkHWPnK6kOG53VTiqtsILNgw==")
		Expect(err).To(BeNil())
		Expect(hx).To(Equal("d387173271652b382e75ffff947fd5a8652d164a65bc6bf404298f444109b1561326bc72881165b0e89922b8cf10d0641d63e72ba90e1b9dd54e2aadb082cd83"))

		hx, err = iotools.DecodeBase64ToHex("SHA512-04cXMnFlKzgudf//lH/VqGUtFkplvGv0BCmPREEJsVYTJrxyiBFlsOiZIrjPENBkHWPnK6kOG53VTiqtsILNgw==")
		Expect(err).To(BeNil())
		Expect(hx).To(Equal("d387173271652b382e75ffff947fd5a8652d164a65bc6bf404298f444109b1561326bc72881165b0e89922b8cf10d0641d63e72ba90e1b9dd54e2aadb082cd83"))

		hx, err = iotools.DecodeBase64ToHex("Sha-512:04cXMnFlKzgudf//lH/VqGUtFkplvGv0BCmPREEJsVYTJrxyiBFlsOiZIrjPENBkHWPnK6kOG53VTiqtsILNgw==")
		Expect(err).To(BeNil())
		Expect(hx).To(Equal("d387173271652b382e75ffff947fd5a8652d164a65bc6bf404298f444109b1561326bc72881165b0e89922b8cf10d0641d63e72ba90e1b9dd54e2aadb082cd83"))

		s1 := crypto.SHA1.New()
		s1.Write([]byte("hello"))
		sum := s1.Sum(nil)
		hx, err = iotools.DecodeBase64ToHex("sHa-1:" + base64.StdEncoding.EncodeToString(sum))
		Expect(err).To(BeNil())
		Expect(hx).To(Equal(hex.EncodeToString(sum)))
	})
})
