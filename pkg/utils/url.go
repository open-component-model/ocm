package utils

import (
	"net/url"
	"strings"
)

func ParseURL(urlToParse string) (*url.URL, error) {
	if !strings.Contains(urlToParse, "://") {
		urlToParse = "dummy://" + urlToParse
	}
	parsedURL, err := url.Parse(urlToParse)
	if err != nil {
		return nil, err
	}
	return parsedURL, nil
}
