package utils

import (
	"net/url"
	"path"
	"strings"

	"github.com/mandelsoft/goutils/errors"
)

func ParseURL(urlToParse string) (*url.URL, error) {
	const dummyScheme = "dummy://"
	if !strings.Contains(urlToParse, "://") {
		urlToParse = dummyScheme + urlToParse
	}
	parsedURL, err := url.Parse(urlToParse)
	if err != nil {
		return nil, err
	}
	if parsedURL.Scheme == dummyScheme {
		parsedURL.Scheme = ""
	}
	return parsedURL, nil
}

func GetFileExtensionFromUrl(url string) (string, error) {
	u, err := ParseURL(url)
	if err != nil {
		return "", err
	}
	s := path.Base(u.Path)
	pos := strings.LastIndex(s, ".")
	if pos == -1 {
		return "", errors.New("failed to deduct file extension from url")
	}
	return s[pos:], nil
}
