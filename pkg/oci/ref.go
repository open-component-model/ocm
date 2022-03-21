// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"fmt"
	"path"
	"strings"

	"github.com/containerd/containerd/reference/docker"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/grammar"
	"github.com/opencontainers/go-digest"
)

// to find a suitable secret for images on Docker Hub, we need its two domains to do matching
const (
	dockerHubDomain       = "docker.io"
	dockerHubLegacyDomain = "index.docker.io"

	KIND_OCI_REFERENCE       = "oci reference"
	KIND_ARETEFACT_REFERENCE = "artefact reference"
)

// ParseRepo parses a standard ocm repository reference into a internal representation.
func ParseRepo(ref string) (UniformRepositorySpec, error) {
	match := grammar.AnchoredRepositoryRegexp.FindSubmatch([]byte(ref))
	if match == nil {
		match = grammar.AnchoredGenericRepositoryRegexp.FindSubmatch([]byte(ref))
		if match == nil {
			return UniformRepositorySpec{}, errors.ErrInvalid(KIND_OCI_REFERENCE, ref)
		}
		return UniformRepositorySpec{
			Type: string(match[1]),
			Info: string(match[2]),
		}, nil

	}
	return UniformRepositorySpec{
		Type:   string(match[1]),
		Scheme: string(match[2]),
		Host:   string(match[3]),
	}, nil
}

// RefSpec is a go internal representation of a oci reference.
type RefSpec struct {
	UniformRepositorySpec
	// Repository is the part of a reference without its hostname
	Repository string `json:"respository"`
	// +optional
	Tag *string `json:"tag,omitempty"`
	// +optional
	Digest *digest.Digest `json:"digest,omitempty"`
}

// ParseRef parses a oci reference into a internal representation.
func ParseRef(ref string) (RefSpec, error) {
	match := grammar.AnchoredGenericRepositoryRegexp.FindSubmatch([]byte(ref))
	if match == nil {
		return RefSpec{}, errors.ErrInvalid(KIND_OCI_REFERENCE)
	}
	typ := string(match[1])
	ref = string(match[2])
	match = grammar.AnchoredSchemedRegexp.FindSubmatch([]byte(ref))
	if match == nil {
		return RefSpec{}, errors.ErrInvalid(KIND_OCI_REFERENCE)
	}
	scheme := string(match[1])
	ref = string(match[2])

	a, err := docker.ParseAnyReference(ref)
	if err == nil {
		spec := RefSpec{UniformRepositorySpec: UniformRepositorySpec{Type: typ, Scheme: scheme}}
		if t, ok := a.(docker.Named); ok {
			spec.Host = docker.Domain(t)
			spec.Repository = docker.Path(t)
		}
		if t, ok := a.(docker.Tagged); ok {
			tag := t.Tag()
			spec.Tag = &tag
		}
		if t, ok := a.(docker.Digested); ok {
			digest := t.Digest()
			spec.Digest = &digest
		}
		return spec, nil
	}
	return RefSpec{}, errors.ErrInvalidWrap(err, KIND_OCI_REFERENCE, ref)
}

func (r *RefSpec) Name() string {
	return path.Join(r.Host, r.Repository)
}

func (r *RefSpec) Base() string {
	if r.Scheme == "" {
		return r.Host
	}
	return r.Scheme + "://" + r.Host
}

func (r *RefSpec) HostPort() (string, string) {
	i := strings.Index(r.Host, ":")
	if i < 0 {
		return r.Host, ""
	}
	return r.Host[:i], r.Host[i+1:]
}

func (r *RefSpec) Version() string {
	if r.Tag != nil {
		return *r.Tag
	}
	if r.Digest != nil {
		return "@" + string(*r.Digest)
	}
	return "latest"
}

func (r *RefSpec) IsVersion() bool {
	return r.Tag != nil || r.Digest != nil
}

func (r *RefSpec) String() string {
	if r.Tag != nil {
		return fmt.Sprintf("%s:%s", r.Name(), *r.Tag)
	}
	if r.Digest != nil {
		return fmt.Sprintf("%s@%s", r.Name(), r.Digest.String())
	}
	return ""
}

// CredHost fallback to legacy docker domain if applicable
// this is how containerd translates the old domain for DockerHub to the new one, taken from containerd/reference/docker/reference.go:674
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

func ParseArt(art string) (ArtSpec, error) {
	match := grammar.AnchoredArtefactVersionRegexp.FindSubmatch([]byte(art))

	if match == nil {
		return ArtSpec{}, errors.ErrInvalid(KIND_ARETEFACT_REFERENCE, art)
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
			return ArtSpec{}, errors.ErrInvalidWrap(err, KIND_ARETEFACT_REFERENCE, art)
		}
		dig = &d
	}
	return ArtSpec{
		Repository: string(match[1]),
		Tag:        tag,
		Digest:     dig,
	}, nil
}

// ArtSpec is a go internal representation of a oci reference.
type ArtSpec struct {
	// Repository is the part of a reference without its hostname
	Repository string
	// +optional
	Tag *string
	// +optional
	Digest *digest.Digest
}

func (r *ArtSpec) IsVersion() bool {
	return r.Tag != nil || r.Digest != nil
}

func (r *ArtSpec) Reference() string {
	if r.Tag != nil {
		return *r.Tag
	}
	if r.Digest != nil {
		return "@" + string(*r.Digest)
	}
	return "latest"
}

func (r *ArtSpec) String() string {
	s := r.Repository
	if r.Tag != nil {
		s += fmt.Sprintf(":%s", *r.Tag)
	}
	if r.Digest != nil {
		s += fmt.Sprintf("@%s", r.Digest.String())
	}
	return s
}
