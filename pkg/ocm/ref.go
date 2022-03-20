// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ocm

import (
	"fmt"
	"strings"

	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/ocm/grammar"
)

const (
	dockerHubDomain       = "docker.io"
	dockerHubLegacyDomain = "index.docker.io"

	KIND_OCM_REFERENCE = "ocm reference"
)

// RefSpec is a go internal representation of a oci reference.
type RefSpec struct {
	UniformRepositorySpec
	CompSpec
}

// ParseRef parses a standard ocm reference into a internal representation.
func ParseRef(ref string) (RefSpec, error) {
	match := grammar.AnchoredComponentVersionRegexp.FindSubmatch([]byte(ref))
	if match == nil {
		return RefSpec{}, errors.ErrInvalid(KIND_OCM_REFERENCE, ref)
	}
	v := string(match[5])
	r := RefSpec{
		UniformRepositorySpec{
			Type:    string(match[1]),
			Host:    string(match[2]),
			SubPath: string(match[3]),
		},
		CompSpec{
			Component: string(match[4]),
			Version:   nil,
		},
	}
	if v != "" {
		r.Version = &v
	}
	return r, nil
}

func (r *RefSpec) Name() string {
	if r.SubPath == "" {
		return fmt.Sprintf("%s//%s", r.Host, r.Component)
	}
	return fmt.Sprintf("%s/%s//%s", r.Host, r.SubPath, r.Component)
}

func (r *RefSpec) HostPort() (string, string) {
	i := strings.Index(r.Host, ":")
	if i < 0 {
		return r.Host, ""
	}
	return r.Host[:i], r.Host[i+1:]
}

func (r *RefSpec) Reference() string {
	t := r.Type
	if t != "" {
		t = t + "::"
	}
	s := r.SubPath
	if s != "" {
		s = "/" + s
	}
	v := ""
	if r.Version != nil && *r.Version != "" {
		v = ":" + *r.Version
	}
	return fmt.Sprintf("%s%s%s//%s%s", t, r.Host, s, r.Component, v)
}

func (r *RefSpec) IsVersion() bool {
	return r.Version != nil
}

func (r *RefSpec) String() string {
	return r.Reference()
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
	if r.Info != nil {
		info := map[string]string{}
		for k, v := range r.Info {
			info[k] = v
		}
		r.Info = info
	}
	return r
}

////////////////////////////////////////////////////////////////////////////////

func ParseComp(ref string) (CompSpec, error) {
	match := grammar.AnchoredComponentVersionRegexp.FindSubmatch([]byte(ref))

	if match == nil {
		return CompSpec{}, errors.ErrInvalid(KIND_COMPONENTVERSION, ref)
	}

	v := string(match[2])
	r := CompSpec{
		Component: string(match[1]),
		Version:   nil,
	}
	if v != "" {
		r.Version = &v
	}
	return r, nil
}

// CompSpec is a go internal representation of a ocm component version name.
type CompSpec struct {
	// Component is the component name part of a component version
	Component string
	// +optional
	Version *string
}

func (r *CompSpec) IsVersion() bool {
	return r.Version != nil
}

func (r *CompSpec) Reference() string {
	v := ""
	if r.Version != nil && *r.Version != "" {
		v = ":" + *r.Version
	}
	return fmt.Sprintf("%s%s", r.Component, v)
}

func (r *CompSpec) String() string {
	return r.Reference()
}
