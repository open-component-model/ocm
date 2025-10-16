package signing_test

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm"
	. "ocm.software/ocm/api/ocm/tools/signing"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/api/tech/signing/signutils"
)

const NAME = "mandelsoft"

var _ = Describe("options", func() {
	defer GinkgoRecover()

	capriv, capub, err := rsa.Handler{}.CreateKeyPair()
	Expect(err).To(Succeed())

	spec := &signutils.Specification{
		RootCAs:      nil,
		IsCA:         true,
		PublicKey:    capub,
		CAPrivateKey: capriv,
		CAChain:      nil,
		Subject: pkix.Name{
			CommonName: "ca-authority",
		},
		Usages:    signutils.Usages{x509.ExtKeyUsageCodeSigning},
		Validity:  10 * time.Hour,
		NotBefore: nil,
	}

	ca, _, err := signutils.CreateCertificate(spec)
	Expect(err).To(Succeed())

	priv, pub, err := rsa.Handler{}.CreateKeyPair()
	Expect(err).To(Succeed())

	spec.Subject = pkix.Name{
		CommonName:    NAME,
		StreetAddress: []string{"some street 21"},
	}
	spec.RootCAs = ca
	spec.CAChain = ca
	spec.PublicKey = pub
	spec.IsCA = false

	cert, _, err := signutils.CreateCertificate(spec)
	Expect(err).To(Succeed())

	pool := x509.NewCertPool()
	pool.AddCert(ca)

	It("verifies options for verification", func() {
		opts := NewOptions(
			RootCertificates(pool),
			VerifySignature(NAME),
			PrivateKey(NAME, priv),
			PublicKey(NAME, cert),
		)
		Expect(opts.Complete(ocm.DefaultContext())).To(Succeed())
	})

	It("fails for options for verification without root cert", func() {
		opts := NewOptions(
			VerifySignature(NAME),
			PrivateKey(NAME, priv),
			PublicKey(NAME, cert),
		)
		Expect(opts.Complete(ocm.DefaultContext())).To(HaveOccurred())
	})

	It("succeeds for options for signing with verification with root cert", func() {
		opts := NewOptions(
			RootCertificates(pool),
			Sign(signing.DefaultRegistry().GetSigner(rsa.Algorithm), NAME),
			PrivateKey(NAME, priv),
			PublicKey(NAME, cert),
		)
		Expect(opts.Complete(ocm.DefaultContext())).To(Succeed())
	})

	It("fails for options for signing with verification without root cert", func() {
		opts := NewOptions(
			Sign(signing.DefaultRegistry().GetSigner(rsa.Algorithm), NAME),
			PrivateKey(NAME, priv),
			PublicKey(NAME, cert),
		)
		Expect(opts.Complete(ocm.DefaultContext())).To(HaveOccurred())
	})
})
