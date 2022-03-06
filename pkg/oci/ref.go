// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"fmt"
	"path"
	"strings"

	"github.com/containerd/containerd/reference/docker"
	"github.com/opencontainers/go-digest"
)

// to find a suitable secret for images on Docker Hub, we need its two domains to do matching
const (
	dockerHubDomain       = "docker.io"
	dockerHubLegacyDomain = "index.docker.io"
)

// ParseRef parses a oci reference into a internal representation.
func ParseRef(resourceURL string) (RefSpec, error) {
	scheme := ""
	if strings.Contains(resourceURL, "://") {
		// remove protocol if exists
		i := strings.Index(resourceURL, "://")
		scheme = resourceURL[:i]
		resourceURL = resourceURL[i+3:]
	}

	a, err := docker.ParseAnyReference(resourceURL)
	if err == nil {
		spec := RefSpec{Scheme: scheme}
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
	return RefSpec{}, err
}

// RefSpec is a go internal representation of a oci reference.
type RefSpec struct {
	// Scheme
	Scheme string
	// Host is the hostname of a oci ref.
	Host string
	// Repository is the part of a reference without its hostname
	Repository string
	// +optional
	Tag *string
	// +optional
	Digest *digest.Digest
}

func (r *RefSpec) Name() string {
	return path.Join(r.Host, r.Repository)
}

func (r *RefSpec) HostPort() (string, string) {
	i := strings.Index(r.Host, ":")
	if i < 0 {
		return r.Host, ""
	}
	return r.Host[:i], r.Host[i+1:]
}

func (r *RefSpec) Reference() string {
	if r.Tag != nil {
		return *r.Tag
	}
	if r.Digest != nil {
		return string(*r.Digest)
	}
	return "latest"
}

func (r RefSpec) String() string {
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
	refspec := RefSpec{
		Host:       r.Host,
		Repository: r.Repository,
	}
	if r.Tag != nil {
		tag := *r.Tag
		refspec.Tag = &tag
	}
	if r.Digest != nil {
		dig := r.Digest.String()
		d := digest.FromString(dig)
		refspec.Digest = &d
	}
	return refspec
}
