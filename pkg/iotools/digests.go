package iotools

import (
	"encoding/base64"
	"encoding/hex"
	"regexp"
)

func DecodeBase64ToHex(b64encoded string) (string, error) {
	re := regexp.MustCompile(`(?i)^sha\d+-`)
	b64encoded = re.ReplaceAllString(b64encoded, "")
	digest, err := base64.StdEncoding.DecodeString(b64encoded)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(digest), nil
}
