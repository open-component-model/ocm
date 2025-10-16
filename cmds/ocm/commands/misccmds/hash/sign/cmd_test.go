package sign_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm/tools/signing/signingtest"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const ISSUER = "mandelsoft"

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv(signingtest.TestData())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("signs a hash", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("sign", "hash",
			"/testdata/compat/rsa.priv",
			"SHA-256:b06e4c1a68274b876661f9fbf1f100526d289745f6ee847bfef702007b5b14cf")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
algorithm: RSASSA-PKCS1-V1_5
mediaType: application/vnd.ocm.signature.rsa
value: 4447c988ecb26352019bdfd5b3e69005eff685ed1ff699f937091ecad6a3bc2156f8d3e9cb3ba3b9d7ea203a5e7a3af5dcb756553fad33b159134cd306ff072544fe6dd991bc51b3ff5de8dcd3d0a401e768523545c5e343c5010073570975dcaa15217b1338810983b68fb4eb878ad26a1c6c3bc441b0f2029cab6b2a352ae40c979b6098a1a880a2ea0e4747c7f88863c4c91543bf457b0f3a396334e4f6155075fc46b8d45cfa34171108c42bc1f779a03bc0844bf91d50d2b6eeba71806a34d18230b911e80871a544d2b499c20d170e8a2b86f85a799b60cf540925425f5f6ff8ef36d8b63dfe0ffd82e635faa5be2a58e5078925d66e494d43fa10948e
`))
	})
})
