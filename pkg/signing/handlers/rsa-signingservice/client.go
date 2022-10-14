// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package rsa_signingservice

import (
	"bytes"
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

const (
	AcceptHeader = "Accept"

	// MediaTypePEM defines the media type for PEM formatted data.
	MediaTypePEM = "application/x-pem-file"
)

type SigningServerSigner struct {
	ServerURL string
}

func NewSigningClient(serverURL string) (*SigningServerSigner, error) {
	signer := SigningServerSigner{
		ServerURL: serverURL,
	}
	return &signer, nil
}

func (signer *SigningServerSigner) Sign(cctx credentials.Context, signatureAlgo string, hashAlgo crypto.Hash, digest string, issuer string, key interface{}) (*signing.Signature, error) {
	decodedHash, err := hex.DecodeString(digest)
	if err != nil {
		return nil, fmt.Errorf("unable to hex decode hash: %w", err)
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		fmt.Sprintf("%s/sign/%s?hashAlgorithm=%s", signer.ServerURL, strings.ToLower(signatureAlgo), hashAlgo.String()),
		bytes.NewBuffer(decodedHash),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to build http request: %w", err)
	}
	req.Header.Add(AcceptHeader, MediaTypePEM)

	// TODO: split up signing server url into protocol, host, and port for matching?
	consumerId := credentials.ConsumerIdentity{
		credentials.ID_TYPE: CONSUMER_TYPE,
		ID_HOSTNAME:         signer.ServerURL,
	}
	credSource, err := cctx.GetCredentialsForConsumer(consumerId, credentials.CompleteMatch)
	if err != nil && !errors.IsErrUnknown(err) {
		return nil, fmt.Errorf("unable to get credential source: %w", err)
	}

	var caCertPool *x509.CertPool
	var clientCerts []tls.Certificate
	if credSource != nil {
		cred, err := credSource.Credentials(cctx)
		if err != nil {
			return nil, fmt.Errorf("unable to get credentials from credential source: %w", err)
		}

		if !cred.ExistsProperty(CLIENT_CERT) {
			return nil, fmt.Errorf("credential for consumer %+v is missing property %q", consumerId, CLIENT_CERT)
		}
		if !cred.ExistsProperty(PRIVATE_KEY) {
			return nil, fmt.Errorf("credential for consumer %+v is missing property %q", consumerId, PRIVATE_KEY)
		}

		rawClientCert := []byte(cred.GetProperty(CLIENT_CERT))
		rawPrivateKey := []byte(cred.GetProperty(PRIVATE_KEY))
		clientCert, err := tls.X509KeyPair(rawClientCert, rawPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("unable to build client certificate: %w", err)
		}
		clientCerts = append(clientCerts, clientCert)

		if cred.ExistsProperty(CA_CERTS) {
			caCertPool = x509.NewCertPool()
			rawCaCerts := []byte(cred.GetProperty(CA_CERTS))
			if ok := caCertPool.AppendCertsFromPEM(rawCaCerts); !ok {
				return nil, fmt.Errorf("unable to append ca certificates to cert pool")
			}
		}
	}

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:   tls.VersionTLS13,
				RootCAs:      caCertPool,
				Certificates: clientCerts,
			},
		},
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to send http request: %w", err)
	}
	defer res.Body.Close()

	responseBodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request returned with status code %d: %s", res.StatusCode, string(responseBodyBytes))
	}

	signaturePemBlocks, err := rsa.GetSignaturePEMBlocks(responseBodyBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to get signature pem block from response: %w", err)
	}

	if len(signaturePemBlocks) != 1 {
		return nil, fmt.Errorf("expected 1 signature pem block, found %d", len(signaturePemBlocks))
	}
	signatureBlock := signaturePemBlocks[0]

	signature := signatureBlock.Bytes
	if len(signature) == 0 {
		return nil, errors.New("invalid response: signature block doesn't contain signature")
	}

	algorithm := signatureBlock.Headers[rsa.SignaturePEMBlockAlgorithmHeader]
	if algorithm == "" {
		return nil, fmt.Errorf("invalid response: %s header is empty: %s", rsa.SignaturePEMBlockAlgorithmHeader, string(responseBodyBytes))
	}

	encodedSignature := pem.EncodeToMemory(signatureBlock)

	return &signing.Signature{
		Value:     string(encodedSignature),
		MediaType: MediaTypePEM,
		Algorithm: algorithm,
		Issuer:    issuer,
	}, nil
}
