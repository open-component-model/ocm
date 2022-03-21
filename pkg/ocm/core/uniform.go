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

package core

import (
	"fmt"
	"sync"

	"github.com/gardener/ocm/pkg/datacontext/vfsattr"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/repositories/ocireg"
	"github.com/gardener/ocm/pkg/runtime"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

const (
	dockerHubDomain       = "docker.io"
	dockerHubLegacyDomain = "index.docker.io"
)

// UniformRepositorySpec is is generic specification of the repository
// for handling as part of standard references
type UniformRepositorySpec struct {
	// Type
	Type string `json:"type,omitempty"`
	// Host is the hostname of an ocm ref.
	Host string `json:"host,omitempty"`
	// SubPath is the sub path spec used to host component versions
	SubPath string `json:"subPath,omitempty"`
	// Info is the file path used to host ctf component versions
	Info string `json:"filePath,omitempty"`
}

// CredHost fallback to legacy docker domain if applicable
// this is how containerd translates the old domain for DockerHub to the new one, taken from containerd/reference/docker/reference.go:674
func (r *UniformRepositorySpec) CredHost() string {
	if r.Host == dockerHubDomain {
		return dockerHubLegacyDomain
	}
	return r.Host
}

func (u *UniformRepositorySpec) String() string {
	t := u.Type
	if t != "" && t != ocireg.OCIRegistryRepositoryType {
		t = t + "::"
	}
	if u.Info != "" {
		return fmt.Sprintf("%s%s", t, u.Info)

	} else {
		s := u.SubPath
		if s != "" {
			s = "/" + s
		}
		return fmt.Sprintf("%s%s%s", t, u.Host, s)
	}
}

type RepositorySpecHandler interface {
	MapReference(ctx Context, u *UniformRepositorySpec) (RepositorySpec, error)
}

type RepositorySpecHandlers interface {
	Register(hdlr RepositorySpecHandler, types ...string)
	MapUniformRepositorySpec(ctx Context, u *UniformRepositorySpec, aliases map[string]RepositorySpec) (RepositorySpec, error)
}

var DefaultRepositorySpecHandlers = NewRepositorySpecHandlers()

func RegisterRepositorySpecHandler(hdlr RepositorySpecHandler, types ...string) {
	DefaultRepositorySpecHandlers.Register(hdlr, types...)
}

type specHandlers struct {
	lock     sync.RWMutex
	handlers map[string][]RepositorySpecHandler
}

func NewRepositorySpecHandlers() RepositorySpecHandlers {
	return &specHandlers{handlers: map[string][]RepositorySpecHandler{}}
}

func (s *specHandlers) Register(hdlr RepositorySpecHandler, types ...string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if hdlr != nil {
		for _, typ := range types {
			s.handlers[typ] = append(s.handlers[typ], hdlr)
		}
	}
}

func (s *specHandlers) MapUniformRepositorySpec(ctx Context, u *UniformRepositorySpec, aliases map[string]RepositorySpec) (RepositorySpec, error) {
	var err error
	s.lock.RLock()
	defer s.lock.RUnlock()

	if len(aliases) > 0 && u.Type == "" {
		if u.Info != "" {
			spec := aliases[u.Info]
			if spec != nil {
				return spec, nil
			}
		}
		if u.Host != "" {
			spec := aliases[u.Host]
			if spec != nil {
				return spec, nil
			}
		}
	}

	for _, h := range s.handlers[u.Type] {
		spec, err := h.MapReference(ctx, u)
		if err != nil || spec != nil {
			return spec, err
		}
	}
	if u.Info != "" {
		spec := &runtime.UnstructuredVersionedTypedObject{}
		err = runtime.DefaultJSONEncoding.Unmarshal([]byte(u.Info), spec)
		if err == nil {
			if spec.GetType() == spec.GetKind() && spec.GetVersion() == "v1" { // only type set, use it as version
				spec.SetType(u.Type + runtime.VersionSeparator + spec.GetType())
			}
			if spec.GetKind() != u.Type {
				return nil, errors.ErrInvalid()
			}
			return ctx.RepositoryTypes().CreateRepositorySpec(spec)
		}
		// generic info set, no json, but existing file
		if ok, err := vfs.Exists(vfsattr.Get(ctx), u.Info); ok && err == nil {
			list := s.handlers[CommonTransportFormat]
			for _, h := range list {
				spec, err := h.MapReference(ctx, u)
				if err != nil {
					return nil, err
				}
				if spec != nil {
					return spec, nil
				}
			}
		}
	}
	for _, h := range s.handlers["*"] {
		spec, err := h.MapReference(ctx, u)
		if err != nil || spec != nil {
			return spec, err
		}
	}

	return nil, errors.ErrNotSupported("uniform repository ref %q", u.String())
}
