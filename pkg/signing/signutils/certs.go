// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package signutils

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/modern-go/reflect2"

	"github.com/open-component-model/ocm/pkg/errors"
)

type Usages []interface{}

type Specification struct {
	RootCAs      interface{}
	IsCA         bool
	PublicKey    interface{}
	CAPrivateKey interface{}
	CAChain      interface{}
	Subject      pkix.Name
	Usages       Usages
	Validity     time.Duration
	NotBefore    *time.Time

	Hosts []string
}

func CreateCertificate(spec *Specification) (*x509.Certificate, []byte, error) {
	var err error

	var rootCerts *x509.CertPool

	if !reflect2.IsNil(spec.RootCAs) {
		rootCerts, err = GetCertPool(spec.RootCAs, true)
		if err != nil {
			return nil, nil, err
		}
	}

	var pubKey interface{}
	if !reflect2.IsNil(spec.PublicKey) {
		pubKey, err = GetPublicKey(spec.PublicKey)
		if err != nil {
			return nil, nil, err
		}
	}

	var caChain []*x509.Certificate
	if !reflect2.IsNil(spec.CAChain) {
		caChain, err = GetCertificateChain(spec.CAChain, false)
		if err != nil {
			return nil, nil, err
		}
	}

	var caPrivKey interface{}
	if !reflect2.IsNil(spec.CAPrivateKey) {
		caPrivKey, err = GetPrivateKey(spec.CAPrivateKey)
		if err != nil {
			return nil, nil, err
		}
	}
	if reflect2.IsNil(caPrivKey) {
		return nil, nil, fmt.Errorf("private key required for signing")
	}

	if reflect2.IsNil(pubKey) {
		pubKey, err = GetPublicKey(caPrivKey)
		if err != nil {
			return nil, nil, err
		}
	}

	var notBefore time.Time
	if spec.NotBefore == nil {
		notBefore = time.Now()
	} else {
		notBefore, err = GetTime(spec.NotBefore)
		if err != nil {
			return nil, nil, err
		}
	}

	if len(caChain) > 0 {
		if rootCerts == nil {
			rootCerts, err = x509.SystemCertPool()
			if err != nil {
				return nil, nil, err
			}
		}

		intermediates, err := GetCertPool(caChain, false)
		if err != nil {
			return nil, nil, err
		}

		opts := x509.VerifyOptions{
			Intermediates:             intermediates,
			Roots:                     rootCerts,
			CurrentTime:               time.Time{},
			KeyUsages:                 []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
			MaxConstraintComparisions: 0,
		}

		_, err = caChain[0].Verify(opts)
		if err != nil {
			return nil, nil, err
		}
	}

	validity := spec.Validity
	if validity == 0 {
		validity = time.Hour*24 + 365
	}
	notAfter := notBefore.Add(validity)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to generate serial number")
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      spec.Subject,
		NotBefore:    notBefore,
		NotAfter:     notAfter,

		KeyUsage:              0,
		ExtKeyUsage:           []x509.ExtKeyUsage{},
		BasicConstraintsValid: true,
	}

	var ca *x509.Certificate
	if len(caChain) == 0 {
		ca = template // se0fl signed certificate
	} else {
		ca = caChain[0]
	}

	for _, u := range spec.Usages {
		k := GetKeyUsage(u)
		if k == nil {
			return nil, nil, fmt.Errorf("invalid usage key %q", u)
		}
		k.AddTo(template)
	}

	for _, h := range spec.Hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if spec.IsCA || (template.KeyUsage&x509.KeyUsageCertSign) != 0 {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, ca, pubKey, caPrivKey)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create certificate")
	}

	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		panic("failed to parse generated certificate:" + err.Error())
	}

	pemBytes := CertificateBytesToPem(derBytes)
	for _, c := range caChain {
		pemBytes = append(pemBytes, CertificateToPem(c)...)
	}
	return cert, pemBytes, nil
}

func CertificateToPem(c *x509.Certificate) []byte {
	return CertificateBytesToPem(c.Raw)
}

func CertificateBytesToPem(derBytes []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
}
