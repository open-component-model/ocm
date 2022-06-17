// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package oci

import (
	"fmt"
	"path"
	"strings"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/contexts/oci/grammar"
	"github.com/open-component-model/ocm/pkg/errors"
)

// to find a suitable secret for images on Docker Hub, we need its two domains to do matching
const (
	dockerHubDomain       = "docker.io"
	dockerHubLegacyDomain = "index.docker.io"

	KIND_OCI_REFERENCE       = "oci reference"
	KIND_ARETEFACT_REFERENCE = "artefact reference"
)

// ParseRepo parses a standard oci repository reference into a internal representation.
func ParseRepo(ref string) (UniformRepositorySpec, error) {
	create := false
	if strings.HasPrefix(ref, "+") {
		create = true
		ref = ref[1:]
	}
	match := grammar.AnchoredRegistryRegexp.FindSubmatch([]byte(ref))
	if match == nil {
		match = grammar.AnchoredGenericRegistryRegexp.FindSubmatch([]byte(ref))
		if match == nil {
			return UniformRepositorySpec{}, errors.ErrInvalid(KIND_OCI_REFERENCE, ref)
		}
		return UniformRepositorySpec{
			Type:            string(match[1]),
			Info:            string(match[2]),
			CreateIfMissing: create,
		}, nil

	}
	return UniformRepositorySpec{
		Type:            string(match[1]),
		Scheme:          string(match[2]),
		Host:            string(match[3]),
		CreateIfMissing: create,
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

	match := grammar.FileReferenceRegexp.FindSubmatch([]byte(ref))
	if match != nil {
		spec.Type = string(match[1])
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
		spec.Type = string(match[1])
		spec.Scheme = string(match[2])
		spec.Host = string(match[3])
		spec.Repository = string(match[4])
		spec.Tag = pointer(match[5])
		spec.Digest = dig(match[6])
		return spec, nil
	}
	match = grammar.TypedGenericReferenceRegexp.FindSubmatch([]byte(ref))
	if match != nil {
		spec.Type = string(match[1])
		spec.Info = string(match[2])
		spec.Repository = string(match[3])
		spec.Tag = pointer(match[4])
		spec.Digest = dig(match[5])
		return spec, nil
	}
	match = grammar.AnchoredRegistryRegexp.FindSubmatch([]byte(ref))
	if match != nil {
		spec.Type = string(match[1])
		spec.Info = string(match[2])
		spec.Repository = string(match[3])
		spec.Tag = pointer(match[4])
		spec.Digest = dig(match[5])
		return spec, nil
	}

	match = grammar.AnchoredGenericRegistryRegexp.FindSubmatch([]byte(ref))
	if match != nil {
		spec.Type = string(match[1])
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

func (r *RefSpec) IsRegistry() bool {
	return r.Repository == ""
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
