// Copyright 2020 Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package docker

import (
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	cpi2 "github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	DockerDeamonRepositoryType   = "DockerDaemon"
	DockerDaemonRepositoryTypeV1 = DockerDeamonRepositoryType + runtime.VersionSeparator + "v1"
)

func init() {
	cpi2.RegisterRepositoryType(DockerDeamonRepositoryType, cpi2.NewRepositoryType(DockerDeamonRepositoryType, &RepositorySpec{}))
	cpi2.RegisterRepositoryType(DockerDaemonRepositoryTypeV1, cpi2.NewRepositoryType(DockerDaemonRepositoryTypeV1, &RepositorySpec{}))
}

// RepositorySpec describes an OCI registry interface backed by an oci registry.
type RepositorySpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	DockerHost                  string `json:dockerHost,omitempty`
}

// NewRepositorySpec creates a new RepositorySpec for an optional host
func NewRepositorySpec(host ...string) *RepositorySpec {
	h := ""
	if len(host) > 0 {
		h = host[0]
	}
	return &RepositorySpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(DockerDeamonRepositoryType),
		DockerHost:          h,
	}
}

func (a *RepositorySpec) GetType() string {
	return DockerDeamonRepositoryType
}

func (a *RepositorySpec) Name() string {
	return DockerDeamonRepositoryType
}

func (a *RepositorySpec) Repository(ctx cpi2.Context, creds credentials.Credentials) (cpi2.Repository, error) {
	return NewRepository(ctx, a)
}
