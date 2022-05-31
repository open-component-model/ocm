package handlers

import (
	"bytes"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

const (
	AcceptHeader = "Accept"

	// MediaTypePEM defines the media type for PEM formatted data.
	MediaTypePEM = "application/x-pem-file"
)

type SigningServerSigner struct {
	Url      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewSigningClient(url string, username, password string) (*SigningServerSigner, error) {
	return &SigningServerSigner{
		url, username, password,
	}, nil
}

func (signer *SigningServerSigner) Sign(digest string, key interface{}) (*signing.Signature, error) {
	decodedHash, err := hex.DecodeString(digest)
	if err != nil {
		return nil, fmt.Errorf("failed decoding hash: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/sign", signer.Url), bytes.NewBuffer(decodedHash))
	if err != nil {
		return nil, fmt.Errorf("failed building http request: %w", err)
	}
	req.Header.Add(AcceptHeader, MediaTypePEM)
	req.SetBasicAuth(signer.Username, signer.Password)

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed sending request: %w", err)
	}
	defer res.Body.Close()

	responseBodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request returned with response code %d: %s", res.StatusCode, string(responseBodyBytes))
	}

	signaturePemBlocks, err := rsa.GetSignaturePEMBlocks(responseBodyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed getting signature pem block from response: %w", err)
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
		return nil, fmt.Errorf("invalid response: %s header is empty", rsa.SignaturePEMBlockAlgorithmHeader)
	}

	encodedSignature := pem.EncodeToMemory(signatureBlock)

	return &signing.Signature{
		Value:     string(encodedSignature),
		MediaType: MediaTypePEM,
		Algorithm: algorithm,
	}, nil
}
