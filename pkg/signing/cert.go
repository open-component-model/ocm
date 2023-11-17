// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package signing

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"time"

	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/signing/signutils"
)

func VerifyCert(intermediate, root *x509.CertPool, cn string, cert *x509.Certificate) error {
	opts := x509.VerifyOptions{
		Intermediates: intermediate,
		Roots:         root,
		CurrentTime:   time.Now(),
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
	}
	_, err := cert.Verify(opts)
	if err != nil {
		return err
	}
	if cn != "" && cert.Subject.CommonName != cn {
		return errors.ErrInvalid("common name", cn)
	}
	if cert.KeyUsage&x509.KeyUsageDigitalSignature != 0 {
		return nil
	}
	for _, k := range cert.ExtKeyUsage {
		if k == x509.ExtKeyUsageCodeSigning {
			return nil
		}
	}
	return errors.ErrNotSupported("codesign", "", "certificate")
}

// Deprecated: use signutils.CreateCertificate.
func CreateCertificate(subject pkix.Name, validFrom *time.Time,
	validity time.Duration, pub interface{},
	ca *x509.Certificate, priv interface{}, isCA bool, names ...string,
) ([]byte, error) {
	spec := &signutils.Specification{
		RootCAs:      nil,
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
