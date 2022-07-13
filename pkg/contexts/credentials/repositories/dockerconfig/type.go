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

package dockerconfig

import (
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	DockerConfigRepositoryType   = "DockerConfig"
	DockerConfigRepositoryTypeV1 = DockerConfigRepositoryType + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterRepositoryType(DockerConfigRepositoryType, cpi.NewRepositoryType(DockerConfigRepositoryType, &RepositorySpec{}))
	cpi.RegisterRepositoryType(DockerConfigRepositoryTypeV1, cpi.NewRepositoryType(DockerConfigRepositoryTypeV1, &RepositorySpec{}))
}

// RepositorySpec describes a docker config based credential repository interface.
type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	DockerConfigFile            string `json:"dockerConfigFile"`
	PropgateConsumerIdentity    bool   `json:"propagateConsumerIdentity,omitempty"`
}

func (s RepositorySpec) WithConsumerPropagation(propagate bool) *RepositorySpec {
	s.PropgateConsumerIdentity = propagate
	return &s
}

// NewRepositorySpec creates a new memory RepositorySpec
func NewRepositorySpec(path string, prop ...bool) *RepositorySpec {
	p := false
	for _, e := range prop {
		p = p || e
	}
	return &RepositorySpec{
		ObjectVersionedType:      runtime.NewVersionedObjectType(DockerConfigRepositoryType),
		DockerConfigFile:         path,
		PropgateConsumerIdentity: p,
	}
}

func (a *RepositorySpec) GetType() string {
	return DockerConfigRepositoryType
}

func (a *RepositorySpec) Repository(ctx cpi.Context, creds cpi.Credentials) (cpi.Repository, error) {
	repos := ctx.GetAttributes().GetOrCreateAttribute(ATTR_REPOS, newRepositories).(*Repositories)
	return repos.GetRepository(ctx, a.DockerConfigFile, a.PropgateConsumerIdentity)
}
