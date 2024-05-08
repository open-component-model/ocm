package helm

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/mandelsoft/vfs/pkg/vfs"
)

// LoadCertificateBundle loads certificates from the given file.  The file should be pem encoded
// containing one or more certificates.  The expected pem type is "CERTIFICATE".
func LoadCertificateBundle(filename string, fs vfs.FileSystem) ([]*x509.Certificate, error) {
	b, err := vfs.ReadFile(fs, filename)
	if err != nil {
		return nil, err
	}
	return LoadCertificateBundleFromData(b)
}

func LoadCertificateBundleFromData(b []byte) ([]*x509.Certificate, error) {
	var block *pem.Block

	certificates := []*x509.Certificate{}
	block, b = pem.Decode(b)
	for ; block != nil; block, b = pem.Decode(b) {
		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, err
			}
			certificates = append(certificates, cert)
		} else {
			return nil, fmt.Errorf("invalid pem block type: %s", block.Type)
		}
	}
	return certificates, nil
}
