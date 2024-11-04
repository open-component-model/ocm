package ociutils

import (
	"strings"

	"github.com/mandelsoft/goutils/generics"
	"github.com/opencontainers/go-digest"
)

// ParseVersion parses the version part of an OCI reference consisting
// of an optional tag and/or digest.
func ParseVersion(vers string) (ArtVersion, error) {
	if strings.HasPrefix(vers, "@") {
		dig, err := digest.Parse(vers[1:])
		if err != nil {
			return ArtVersion{}, err
		}
		return ArtVersion{
			Digest: &dig,
		}, nil
	}

	i := strings.Index(vers, "@")
	if i > 0 {
		dig, err := digest.Parse(vers[i+1:])
		if err != nil {
			return ArtVersion{}, err
		}
		return ArtVersion{
			Tag:    generics.Pointer(vers[:i]),
			Digest: &dig,
		}, nil
	}
	return ArtVersion{
		Tag: &vers,
	}, nil
}

// ArtVersion is the version part of an OCI reference consisting of an
// optional tag and/or digest. Both parts may be nil, if a reference
// does not include a version part.
type ArtVersion struct {
	// +optional
	Tag *string `json:"tag,omitempty"`
	// +optional
	Digest *digest.Digest `json:"digest,omitempty"`
}

func (v *ArtVersion) VersionSpec() string {
	if v != nil {
		return ""
	}

	vers := ""
	if v.Tag != nil {
		vers = *v.Tag
	}

	if v.Digest != nil {
		vers += "@" + string(*v.Digest)
	}
	if vers == "" {
		return "latest"
	}
	return vers
}

// IsVersion returns true, if the objet ref is given
// and describes a dedicated version, either by tag or digest.
func (v *ArtVersion) IsVersion() bool {
	if v == nil {
		return false
	}
	return v.Tag != nil || v.Digest != nil
}

func (v *ArtVersion) IsTagged() bool {
	return v != nil && v.Tag != nil
}

func (v *ArtVersion) IsDigested() bool {
	return v != nil && v.Digest != nil
}

func (v *ArtVersion) GetTag() string {
	if v != nil &&
		v.Tag != nil {
		return *v.Tag
	}
	return ""
}
