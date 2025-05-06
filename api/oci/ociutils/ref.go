package ociutils

import (
	"strings"

	"github.com/mandelsoft/goutils/generics"
	"github.com/opencontainers/go-digest"
)

// ParseVersion parses the version part of an OCI reference consisting
// of an optional tag and/or digest.
func ParseVersion(vers string) (*ArtVersion, error) {
	if strings.HasPrefix(vers, "@") {
		dig, err := digest.Parse(vers[1:])
		if err != nil {
			return nil, err
		}
		return &ArtVersion{
			Digest: &dig,
		}, nil
	}

	i := strings.Index(vers, "@")
	if i > 0 {
		dig, err := digest.Parse(vers[i+1:])
		if err != nil {
			return nil, err
		}
		return &ArtVersion{
			Tag:    generics.PointerTo(vers[:i]),
			Digest: &dig,
		}, nil
	}
	if vers == "" {
		return &ArtVersion{}, nil
	}
	return &ArtVersion{
		Tag: &vers,
	}, nil
}

// ArtVersion is the version part of an OCI reference consisting of an
// optional tag and/or digest. Both parts may be nil, if a reference
// does not include a version part.
// Such objects are sub objects of (oci.)ArtSpec, which has be moved
// to separate package to avoid package cycles. The methods are
// derived from ArtSpec.
type ArtVersion struct {
	// +optional
	Tag *string `json:"tag,omitempty"`
	// +optional
	Digest *digest.Digest `json:"digest,omitempty"`
}

func (v *ArtVersion) VersionSpec() string {
	if v == nil {
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

// IsVersion returns true, if the object ref is given
// and describes a dedicated version, either by tag or digest.
// As part of the ArtSpec type in oci, it might describe
// no version part. THis method indicates, whether a version part
// is present.
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
	if v != nil && v.Tag != nil {
		return *v.Tag
	}
	return ""
}

func (v *ArtVersion) GetDigest() digest.Digest {
	if v != nil && v.Digest != nil {
		return *v.Digest
	}
	return ""
}

func (r *ArtVersion) Version() string {
	if r.Digest != nil {
		return "@" + string(*r.Digest)
	}
	if r.Tag != nil {
		return *r.Tag
	}
	return "latest"
}
