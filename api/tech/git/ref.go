package git

import (
	"fmt"
	"hash/fnv"
	"net/url"
	"regexp"

	"github.com/go-git/go-git/v5/plumbing"

	"ocm.software/ocm/api/utils"
)

const urlToRefSeparator = "@"
const refToPathSeparator = "#"

// refRegexp is a regular expression that matches a git ref string.
// The ref string is expected to be in the format of:
// <url>@<ref>#<path>
// where:
// - url is the URL of the git repository
// - ref is the git reference to checkout, if not specified, defaults to "HEAD"
// - path is the path to the file or directory to use as the source, if not specified, defaults to the root of the repository.
var refRegexp = regexp.MustCompile(`^([^@#]+)(@[^#\n]+)?(#[^@\n]+)?`)

// gurl represents a git URL reference.
// It contains the URL of the git repository,
// the reference to check out,
// and the path to the file or directory to use as the source.
type gurl struct {
	url  *url.URL
	ref  plumbing.ReferenceName
	path string
}

// decodeGitURL decodes a git ref string into a gurl struct.
// The ref string is expected to be in the format of:
// <url>@<ref>#<path>
// see refRegexp for more details.
func decodeGitURL(rawRef string) (*gurl, error) {
	matches := refRegexp.FindStringSubmatch(rawRef)
	if matches == nil {
		return nil, fmt.Errorf("failed to match ref: %s via %s", rawRef, refRegexp)
	}
	rawURL := matches[1]
	matchedRef := matches[2]
	path := matches[3]

	parsedURL, err := utils.ParseURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	var ref plumbing.ReferenceName
	if matchedRef == "" {
		ref = plumbing.HEAD
	} else {
		ref = plumbing.ReferenceName(matchedRef)
		if err := ref.Validate(); err != nil {
			return nil, fmt.Errorf("failed to validate ref: %w", err)
		}
	}

	return &gurl{
		url:  parsedURL,
		ref:  ref,
		path: path,
	}, nil
}

func (ref *gurl) String() string {
	return fmt.Sprintf("%s%s%s%s%s", ref.url.String(), urlToRefSeparator, ref.ref, refToPathSeparator, ref.path)
}

func (ref *gurl) Hash() []byte {
	hash := fnv.New64()
	_, _ = hash.Write([]byte(ref.url.String()))
	return hash.Sum(nil)
}

func (ref *gurl) HashString() string {
	return fmt.Sprintf("%x", ref.Hash())
}
