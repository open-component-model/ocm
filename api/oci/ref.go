package oci

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/opencontainers/go-digest"
	"ocm.software/ocm/api/oci/grammar"
	"ocm.software/ocm/api/oci/ociutils"
)

// to find a suitable secret for images on Docker Hub, we need its two domains to do matching.
const (
	dockerHubDomain       = "docker.io"
	dockerHubLegacyDomain = "index.docker.io"

	KIND_OCI_REFERENCE       = "oci reference"
	KIND_ARETEFACT_REFERENCE = "artifact reference"
)

// ParseRepo parses a standard oci repository reference into an internal representation.
func ParseRepo(ref string) (UniformRepositorySpec, error) {
	create := false
	if strings.HasPrefix(ref, "+") {
		create = true
		ref = ref[1:]
	}
	uspec := UniformRepositorySpec{}
	match := grammar.AnchoredRegistryRegexp.FindSubmatch([]byte(ref))
	if match == nil {
		match = grammar.AnchoredGenericRegistryRegexp.FindSubmatch([]byte(ref))
		if match == nil {
			return uspec, errors.ErrInvalid(KIND_OCI_REFERENCE, ref)
		}
		uspec.SetType(string(match[1]))
		uspec.Info = string(match[2])
		uspec.CreateIfMissing = create
		return uspec, nil
	}
	uspec.SetType(string(match[1]))
	uspec.Scheme = string(match[2])
	uspec.Host = string(match[3])
	uspec.CreateIfMissing = create
	return uspec, nil
}

// RefSpec is a go internal representation of an oci reference.
type RefSpec struct {
	UniformRepositorySpec `json:",inline"`
	ArtSpec               `json:",inline"`
}

func pointer(b []byte) *string {
	if len(b) == 0 {
		return nil
	}
	s := string(b)
	return &s
}

func dig(b []byte) *digest.Digest {
	if len(b) == 0 {
		return nil
	}
	s := digest.Digest(b)
	return &s
}

// ParseRef parses a oci reference into a internal representation.
func ParseRef(ref string) (RefSpec, error) {
	create := false
	if strings.HasPrefix(ref, "+") {
		create = true
		ref = ref[1:]
	}

	spec := RefSpec{UniformRepositorySpec: UniformRepositorySpec{CreateIfMissing: create}}
	match := grammar.AnchoredTypedSchemedHostPortArtifactRegexp.FindSubmatch([]byte(ref))
	if match != nil {
		spec.SetType(string(match[1]))
		spec.Scheme = string(match[2])
		spec.Host = string(match[3])
		spec.Repository = string(match[4])
		spec.Tag = pointer(match[5])
		spec.Digest = dig(match[6])
		return spec, nil
	}

	match = grammar.AnchoredTypedOptSchemedReqHostReqPortArtifactRegexp.FindSubmatch([]byte(ref))
	if match != nil {
		spec.SetType(string(match[1]))
		spec.Scheme = string(match[2])
		spec.Host = string(match[3])
		spec.Repository = string(match[4])
		spec.Tag = pointer(match[5])
		spec.Digest = dig(match[6])
		return spec, nil
	}
	match = grammar.FileReferenceRegexp.FindSubmatch([]byte(ref))
	if match != nil {
		spec.SetType(string(match[1]))
		spec.Info = string(match[2])
		spec.Repository = string(match[3])
		spec.Tag = pointer(match[4])
		spec.Digest = dig(match[5])
		return spec, nil
	}
	match = grammar.DockerLibraryReferenceRegexp.FindSubmatch([]byte(ref))
	if match != nil {
		spec.Host = dockerHubDomain
		spec.Repository = "library" + grammar.RepositorySeparator + string(match[1])
		spec.Tag = pointer(match[2])
		spec.Digest = dig(match[3])
		return spec, nil
	}
	match = grammar.DockerReferenceRegexp.FindSubmatch([]byte(ref))
	if match != nil {
		spec.Host = dockerHubDomain
		spec.Repository = string(match[1])
		spec.Tag = pointer(match[2])
		spec.Digest = dig(match[3])
		return spec, nil
	}
	match = grammar.ReferenceRegexp.FindSubmatch([]byte(ref))
	if match != nil {
		spec.Scheme = string(match[1])
		spec.Host = string(match[2])
		spec.Repository = string(match[3])
		spec.Tag = pointer(match[4])
		spec.Digest = dig(match[5])
		return spec, nil
	}
	match = grammar.TypedReferenceRegexp.FindSubmatch([]byte(ref))
	if match != nil {
		spec.SetType(string(match[1]))
		spec.Scheme = string(match[2])
		spec.Host = string(match[3])
		spec.Repository = string(match[4])
		spec.Tag = pointer(match[5])
		spec.Digest = dig(match[6])
		return spec, nil
	}
	match = grammar.TypedURIRegexp.FindSubmatch([]byte(ref))
	if match != nil {
		spec.SetType(string(match[1]))
		spec.Scheme = string(match[2])
		spec.Host = string(match[3])
		spec.Repository = string(match[4])
		spec.Tag = pointer(match[5])
		spec.Digest = dig(match[6])
		return spec, nil
	}
	match = grammar.TypedGenericReferenceRegexp.FindSubmatch([]byte(ref))
	if match != nil {
		spec.SetType(string(match[1]))
		spec.Info = string(match[2])
		spec.Repository = string(match[3])
		spec.Tag = pointer(match[4])
		spec.Digest = dig(match[5])
		return spec, nil
	}
	match = grammar.AnchoredRegistryRegexp.FindSubmatch([]byte(ref))
	if match != nil {
		spec.SetType(string(match[1]))
		spec.Info = string(match[2])
		spec.Repository = string(match[3])
		spec.Tag = pointer(match[4])
		spec.Digest = dig(match[5])
		return spec, nil
	}

	match = grammar.AnchoredGenericRegistryRegexp.FindSubmatch([]byte(ref))
	if match != nil {
		spec.SetType(string(match[1]))
		spec.Info = string(match[2])

		match = grammar.ErrorCheckRegexp.FindSubmatch([]byte(ref))
		if match != nil {
			if len(match[3]) != 0 || len(match[4]) != 0 {
				return RefSpec{}, errors.ErrInvalid(KIND_OCI_REFERENCE, ref)
			}
		}
		return spec, nil
	}
	return RefSpec{}, errors.ErrInvalid(KIND_OCI_REFERENCE, ref)
}

func (r *RefSpec) Name() string {
	return r.UniformRepositorySpec.ComposeRef(r.Repository)
}

func (r *RefSpec) String() string {
	art := r.Repository
	if r.Tag != nil {
		art = fmt.Sprintf("%s:%s", art, *r.Tag)
	}
	if r.Digest != nil {
		art = fmt.Sprintf("%s@%s", art, r.Digest.String())
	}
	return r.UniformRepositorySpec.ComposeRef(art)
}

// CredHost fallback to legacy docker domain if applicable
// this is how containerd translates the old domain for DockerHub to the new one, taken from containerd/reference/docker/reference.go:674.
func (r *RefSpec) CredHost() string {
	if r.Host == dockerHubDomain {
		return dockerHubLegacyDomain
	}
	return r.Host
}

func (r RefSpec) DeepCopy() RefSpec {
	if r.Tag != nil {
		tag := *r.Tag
		r.Tag = &tag
	}
	if r.Digest != nil {
		dig := *r.Digest
		r.Digest = &dig
	}
	return r
}

////////////////////////////////////////////////////////////////////////////////

// ParseVersion parses an OCI version part of an OCI reference.
// It has to be placed in a utils package to avoid package cycles
// for particular users.
func ParseVersion(vers string) (*ArtVersion, error) {
	return ociutils.ParseVersion(vers)
}

func ParseArt(art string) (*ArtSpec, error) {
	match := grammar.AnchoredArtifactVersionRegexp.FindSubmatch([]byte(art))

	if match == nil {
		return nil, errors.ErrInvalid(KIND_ARETEFACT_REFERENCE, art)
	}
	var tag *string
	var dig *digest.Digest

	if match[2] != nil {
		t := string(match[2])
		tag = &t
	}
	if match[3] != nil {
		t := string(match[3])
		d, err := digest.Parse(t)
		if err != nil {
			return nil, errors.ErrInvalidWrap(err, KIND_ARETEFACT_REFERENCE, art)
		}
		dig = &d
	}
	return &ArtSpec{
		Repository: string(match[1]),
		ArtVersion: ArtVersion{
			Tag:    tag,
			Digest: dig,
		},
	}, nil
}

type ArtVersion = ociutils.ArtVersion

// ArtSpec is a go internal representation of a oci reference.
type ArtSpec struct {
	// Repository is the part of a reference without its hostname
	Repository string `json:"repository"`
	// artifact version
	ArtVersion `json:",inline"`
}

func (r *ArtSpec) IsRegistry() bool {
	return r.Repository == ""
}

func (r *ArtSpec) String() string {
	if r == nil {
		return ""
	}
	s := r.Repository
	if r.Tag != nil {
		s += fmt.Sprintf(":%s", *r.Tag)
	}
	if r.Digest != nil {
		s += fmt.Sprintf("@%s", r.Digest.String())
	}
	return s
}
