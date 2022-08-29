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

package docker

import (
	"strings"

	"github.com/containers/image/v5/types"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
)

type Repository struct {
	ctx    cpi.Context
	spec   *RepositorySpec
	sysctx *types.SystemContext
	client *client.Client
}

var _ cpi.Repository = &Repository{}

func NewRepository(ctx cpi.Context, spec *RepositorySpec) (*Repository, error) {
	sysctx := &types.SystemContext{
		DockerDaemonHost: spec.DockerHost,
	}
	client, err := newDockerClient(spec.DockerHost)
	if err != nil {
		return nil, err
	}

	return &Repository{
		ctx:    ctx,
		spec:   spec,
		sysctx: sysctx,
		client: client,
	}, nil
}

func (r *Repository) NamespaceLister() cpi.NamespaceLister {
	return r
}

func (r *Repository) NumNamespaces(prefix string) (int, error) {
	repos, err := r.GetRepositories()
	if err != nil {
		return -1, err
	}
	return len(cpi.FilterByNamespacePrefix(prefix, repos)), nil
}

func (r *Repository) GetNamespaces(prefix string, closure bool) ([]string, error) {
	repos, err := r.GetRepositories()
	if err != nil {
		return nil, err
	}
	return cpi.FilterChildren(closure, cpi.FilterByNamespacePrefix(prefix, repos)), nil
}

func (r *Repository) GetRepositories() ([]string, error) {
	opts := dockertypes.ImageListOptions{}
	list, err := r.client.ImageList(dummyContext, opts)
	if err != nil {
		return nil, err
	}
	var result cpi.StringList
	for _, e := range list {
		if len(e.RepoTags) > 0 {
			for _, t := range e.RepoTags {
				i := strings.Index(t, ":")
				if i > 0 {
					if t[:i] != "<none>" {
						result.Add(t[:i])
					}
				}
			}
		} else {
			result.Add("")
		}
	}
	return result, nil
}

func (r *Repository) IsReadOnly() bool {
	return true
}

func (r *Repository) IsClosed() bool {
	return false
}

func (r *Repository) GetSpecification() cpi.RepositorySpec {
	return r.spec
}

func (r *Repository) ExistsArtefact(name string, version string) (bool, error) {
	ref, err := ParseRef(name, version)
	if err != nil {
		return false, err
	}
	opts := dockertypes.ImageListOptions{}
	opts.Filters.Add("reference", ref.StringWithinTransport())
	list, err := r.client.ImageList(dummyContext, opts)
	if err != nil {
		return false, err
	}
	return len(list) > 0, nil
}

func (r *Repository) LookupArtefact(name string, version string) (cpi.ArtefactAccess, error) {
	n, err := r.LookupNamespace(name)
	if err != nil {
		return nil, err
	}
	return n.GetArtefact(version)
}

func (r *Repository) LookupNamespace(name string) (cpi.NamespaceAccess, error) {
	return NewNamespace(r, name)
}

func (r *Repository) Close() error {
	return nil
}
