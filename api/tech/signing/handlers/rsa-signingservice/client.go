package rsa_signingservice

import (
	"bytes"
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/credentials/identity/hostpath"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/signutils"
)

const (
	AcceptHeader = "Accept"

	// MediaTypePEM defines the media type for PEM formatted data.
	MediaTypePEM = "application/x-pem-file"
)

type SigningServerSigner struct {
	ServerURL *url.URL
}

func NewSigningClient(serverURL string) (*SigningServerSigner, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid signing server URL (%q)", serverURL)
	}
	signer := SigningServerSigner{
		ServerURL: u,
	}
	return &signer, nil
}

func (signer *SigningServerSigner) Sign(cctx credentials.Context, signatureAlgo string, hashAlgo crypto.Hash, digest string, sctx signing.SigningContext) (*signing.Signature, error) {
	decodedHash, err := hex.DecodeString(digest)
	if err != nil {
		return nil, fmt.Errorf("unable to hex decode hash: %w", err)
	}

	u := *signer.ServerURL
	if !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	u.Path += "sign/" + strings.ToLower(signatureAlgo)
	q := u.Query()
	q.Set("hashAlgorithm", hashAlgo.String())
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		u.String(),
		bytes.NewBuffer(decodedHash),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to build http request: %w", err)
	}
	req.Header.Add(AcceptHeader, MediaTypePEM)

	// TODO: split up signing server url into protocol, host, and port for matching?
	consumerId := credentials.ConsumerIdentity{
		credentials.ID_TYPE: CONSUMER_TYPE,
		ID_HOSTNAME:         signer.ServerURL.Hostname(),
		ID_PORT:             signer.ServerURL.Port(),
		ID_SCHEME:           signer.ServerURL.Scheme,
		ID_PATHPREFIX:       signer.ServerURL.Path,
	}
	credSource, err := cctx.GetCredentialsForConsumer(consumerId, hostpath.Matcher)
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

		if !cred.ExistsProperty(ATTR_CLIENT_CERT) {
			return nil, fmt.Errorf("credential for consumer %+v is missing property %q", consumerId, ATTR_CLIENT_CERT)
		}
		if !cred.ExistsProperty(ATTR_PRIVATE_KEY) {
			return nil, fmt.Errorf("credential for consumer %+v is missing property %q", consumerId, ATTR_PRIVATE_KEY)
		}

		rawClientCert := []byte(cred.GetProperty(ATTR_CLIENT_CERT))
		rawPrivateKey := []byte(cred.GetProperty(ATTR_PRIVATE_KEY))
		clientCert, err := tls.X509KeyPair(rawClientCert, rawPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("unable to build client certificate: %w", err)
		}
		clientCerts = append(clientCerts, clientCert)

		if cred.ExistsProperty(ATTR_CA_CERTS) {
			caCertPool = x509.NewCertPool()
			rawCaCerts := []byte(cred.GetProperty(ATTR_CA_CERTS))
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

	signature, algorithm, certs, err := signutils.GetSignatureFromPem(responseBodyBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to get signature pem block from response: %w", err)
	}
	if len(signature) == 0 {
		return nil, errors.New("invalid response: signature block doesn't contain signature")
	}

	if algorithm == "" {
		return nil, fmt.Errorf("invalid response: %s header is empty: %s", signutils.SignaturePEMBlockAlgorithmHeader, string(responseBodyBytes))
	}

	encodedSignature := responseBodyBytes

	issuer := sctx.GetIssuer()
	var iss string
	if issuer != nil {
		if len(certs) == 0 {
			return nil, errors.Newf("certificates missing in signing response")
		}
		if err := signutils.MatchDN(certs[0].Subject, *issuer); err != nil {
			return nil, errors.Wrapf(err, "unexpected issuer in signing response")
		}
		iss = issuer.String()
	}
	if len(certs) > 0 {
		err = signutils.VerifyCertificate(certs[0], certs, sctx.GetRootCerts(), issuer)
		if err != nil {
			return nil, err
		}
	}

	return &signing.Signature{
		Value:     string(encodedSignature),
		MediaType: MediaTypePEM,
		Algorithm: algorithm,
		Issuer:    iss,
	}, nil
}
