package signing

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"time"

	parse "github.com/mandelsoft/spiff/dynaml/x509"

	"github.com/open-component-model/ocm/pkg/signing/signutils"
)

// Deprecated: use signutils.GetCertificate.
func GetCertificate(in interface{}) (*x509.Certificate, error) {
	c, _, err := signutils.GetCertificate(in, false)
	return c, err
}

// Deprecated: use signutils.ParsePublicKey.
func ParsePublicKey(data []byte) (interface{}, error) {
	return parse.ParsePublicKey(string(data))
}

// Deprecated: use signutils.ParsePrivateKey.
func ParsePrivateKey(data []byte) (interface{}, error) {
	return parse.ParsePrivateKey(string(data))
}

// Deprecated: use signutils.SystemCertPool.
func BaseRootPool() (*x509.CertPool, error) {
	return signutils.SystemCertPool()
}

// Deprecated: use signutils.CreateCertificate.
func CreateCertificate(subject pkix.Name, validFrom *time.Time,
	validity time.Duration, pub interface{},
	ca *x509.Certificate, priv interface{}, isCA bool, names ...string,
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
