package signing_test

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/api/tech/signing/signutils"
)

// CreateCertificate creates a pem encoded certificate.
func CreateCertificate(subject pkix.Name, validFrom *time.Time, validity time.Duration,
	pub interface{}, ca *x509.Certificate, priv interface{}, isCA bool, names ...string,
) ([]byte, error) {
	spec := &signutils.Specification{
		RootCAs:      ca,
		IsCA:         isCA,
		PublicKey:    pub,
		CAPrivateKey: priv,
		CAChain:      ca,
		Subject:      subject,
		Usages:       signutils.Usages{x509.ExtKeyUsageCodeSigning},
		Validity:     validity,
		NotBefore:    validFrom,
		Hosts:        names,
	}
	_, data, err := signutils.CreateCertificate(spec)
	return data, err
}

var _ = Describe("normalization", func() {
	defer GinkgoRecover()

	capriv, capub, err := rsa.Handler{}.CreateKeyPair()
	Expect(err).To(Succeed())

	subject := pkix.Name{
		CommonName: "ca-authority",
	}
	caData, err := CreateCertificate(subject, nil, 10*time.Hour, capub, nil, capriv, true)
	Expect(err).To(Succeed())
	ca, err := signutils.ParseCertificate(caData)
	Expect(err).To(Succeed())

	priv, pub, err := rsa.Handler{}.CreateKeyPair()
	Expect(err).To(Succeed())

	subject = pkix.Name{
		CommonName:    "mandelsoft",
		StreetAddress: []string{"some street 21"},
	}

	Context("foreignly signed", func() {
		certData, err := CreateCertificate(subject, nil, 10*time.Hour, pub, ca, capriv, false)
		Expect(err).To(Succeed())

		cert, err := signutils.ParseCertificate(certData)
		Expect(err).To(Succeed())

		pool := x509.NewCertPool()
		pool.AddCert(ca)

		It("verifies for issuer", func() {
			err = signing.VerifyCert(nil, pool, "mandelsoft", cert)
			Expect(err).To(Succeed())
		})
		It("verifies for anonymous", func() {
			err = signing.VerifyCert(nil, pool, "", cert)
			Expect(err).To(Succeed())
		})
		It("fails for wrong issuer", func() {
			err = signing.VerifyCert(nil, pool, "x", cert)
			Expect(err).To(HaveOccurred())
		})
	})
	Context("self signed", func() {
		certData, err := CreateCertificate(subject, nil, 10*time.Hour, pub, nil, priv, false)
		Expect(err).To(Succeed())

		cert, err := signutils.ParseCertificate(certData)
		Expect(err).To(Succeed())

		pool := x509.NewCertPool()
		pool.AddCert(cert)

		It("verifies for issuer", func() {
			err = signing.VerifyCert(nil, pool, "mandelsoft", cert)
			Expect(err).To(Succeed())
		})
		It("verifies for anonymous", func() {
			err = signing.VerifyCert(nil, pool, "", cert)
			Expect(err).To(Succeed())
		})
		It("fails for wrong issuer", func() {
			err = signing.VerifyCert(nil, pool, "x", cert)
			Expect(err).To(HaveOccurred())
		})
	})
})
