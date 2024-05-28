package iotools

import (
	"encoding/base64"
	"encoding/hex"
	"regexp"
)

// DecodeBase64ToHex decodes a base64 encoded string and returns the hex representation.
// Any prefix like 'sha512-' or 'SHA-256:' or 'Sha1-' is removed.
func DecodeBase64ToHex(b64encoded string) (string, error) {
	re := regexp.MustCompile(`(?i)^sha(\d+-|\-\d+:)`) // sha512- or SHA-512:
	b64encoded = re.ReplaceAllString(b64encoded, "")
	digest, err := base64.StdEncoding.DecodeString(b64encoded)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(digest), nil
}
