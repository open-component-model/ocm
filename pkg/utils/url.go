package utils

import (
	"net/url"
	"strings"

	"github.com/open-component-model/ocm/pkg/errors"
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
	pos := strings.LastIndex(u.Path, ".")
	if pos == -1 {
		return "", errors.New("failed to deduct file extension from url")
	}
	return u.Path[pos:len(u.Path)], nil
}
