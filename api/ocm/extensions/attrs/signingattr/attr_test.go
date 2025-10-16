package signingattr_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/rootcertsattr"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
)

const NAME = "test"

var _ = Describe("attribute", func() {
	var cfgctx config.Context
	var ocmctx ocm.Context

	BeforeEach(func() {
		ocmctx = ocm.New(datacontext.MODE_EXTENDED)
		cfgctx = ocmctx.ConfigContext()
	})

	It("marshal/unmarshal", func() {
		cfg := signingattr.New()
		cfg.AddPublicKeyData(NAME, []byte("keydata"))

		data, err := json.Marshal(cfg)
		Expect(err).To(Succeed())

		r := &signingattr.Config{}
		Expect(json.Unmarshal(data, r)).To(Succeed())
		Expect(r).To(Equal(cfg))
	})

	It("applies public key", func() {
		cfg := signingattr.New()
		cfg.AddPublicKeyData(NAME, []byte("keydata"))

		Expect(cfgctx.ApplyConfig(cfg, "from test")).To(Succeed())
		Expect(signingattr.Get(ocmctx).GetPublicKey(NAME)).To(Equal([]byte("keydata")))
	})

	It("applies root certificate", func() {
		certdata := `
-----BEGIN CERTIFICATE-----
MIIDBDCCAeygAwIBAgIQF+kRr0G+faDEAH5Y4P1J7DANBgkqhkiG9w0BAQsFADAc
MQwwCgYDVQQKEwNPQ00xDDAKBgNVBAMTA29jbTAeFw0yMzEyMjkxMDIyMzdaFw0y
NDEyMjgxMDIyMzdaMBwxDDAKBgNVBAoTA09DTTEMMAoGA1UEAxMDb2NtMIIBIjAN
BgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvpTQIQFNy23ygef3pshdeNjT7TME
kPEuqrqcF3KIX1cX16pHMQeU+VzXAFRj3xCy3LAM8ZzLsdHSwZDsIsGdg0nAbGjz
+USez/9TGC58ktr/84Kh0gHDE28YSVhsnNSrBJcWaBlYZz4Iy89O2Xc4jbK34Cwg
Si0ES+Ru1lxLD6FSLYLe43wCIjWRJRrMFcua6nI0P4MCpcKmTkXG2/xz80QSobI3
z/isqOT54FKHW8DZZVlQMOxh+loeLksfEq7EYVkQoUWEV6xyR24TEpMGfxERgDre
l7lmx8nIFzRMXkot+P19XWfUBgqctVEiDF4DlRE+SvCZsNCrg7nQuC2AZQIDAQAB
o0IwQDAOBgNVHQ8BAf8EBAMCAqQwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQU
1iQqrWM/bCXMk+5c1bulfI5zlKcwDQYJKoZIhvcNAQELBQADggEBAAQO6lw6ePuX
E+NyhDYCulueMWHC7GRUKa1KpouFT2yM0BSQnP04VakTlwVO3w4w2KucSVVomHR3
hTY9Ypx7iGLaqdXHmUZvx3uaTM5IXQKMMWL1LJsxAvuzucehgDlOnFBD91tKsr5o
VRvRU5ya0igBCnnGpFu7NuH3C9pgF01lrQ3EhUHuNeazxleaE3/uQWmAXfxFB4ci
gHMKSEk3HuYA1raDJFv4ihwO5pXHvlDhcW0C1oMG9lOCh8TXpVzzBDZiH1kWPWSs
gW9YBu7/p/22U4++X23RyaheGuysfRAMv9cTv+8T0J8NHaAmQz4/QHFXh+0/tQgU
EVQVGDF6KNU=
-----END CERTIFICATE-----
`
		cfg := signingattr.New()
		cfg.AddRootCertificateData([]byte(certdata))

		Expect(cfgctx.ApplyConfig(cfg, "from test")).To(Succeed())
		Expect(rootcertsattr.Get(ocmctx).HasRootCertificates()).To(BeTrue())
	})
})
