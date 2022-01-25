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

package config

import (
	"github.com/gardener/ocm/pkg/common"
	"github.com/gardener/ocm/pkg/config/cpi"
	"github.com/gardener/ocm/pkg/credentials"
	"github.com/gardener/ocm/pkg/runtime"
)

const (
	CredentialsConfigType   = "credentials.config" + common.TypeGroupSuffix
	CredentialsConfigTypeV1 = CredentialsConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cpi.RegisterConfigType(CredentialsConfigType, cpi.NewConfigType(CredentialsConfigType, &ConfigSpec{}))
	cpi.RegisterConfigType(CredentialsConfigTypeV1, cpi.NewConfigType(CredentialsConfigTypeV1, &ConfigSpec{}))
}

// ConfigSpec describes a memory based repository interface.
type ConfigSpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	Consumers                   []ConsumerSpec                                 `json:"consumers,omitempty"`
	Repositories                []RepositorySpec                               `json:"repositories,omitempty"`
	Aliases                     map[string]*credentials.GenericCredentialsSpec `json:"aliases,omitempty"`
}

type ConsumerSpec struct {
	Identity    credentials.ConsumerIdentity        `json:"identity"`
	Credentials *credentials.GenericCredentialsSpec `json:"credentials"`
}

type RepositorySpec struct {
	Repository  credentials.GenericRepositorySpec   `json:"repository"`
	Credentials *credentials.GenericCredentialsSpec `json:"credentials,omitempty"`
}

// NewConfigSpec creates a new memory ConfigSpec
func NewConfigSpec() *ConfigSpec {
	return &ConfigSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(CredentialsConfigType),
	}
}

func (a *ConfigSpec) GetType() string {
	return CredentialsConfigType
}

func (a *ConfigSpec) AddConsumer(id credentials.ConsumerIdentity, creds credentials.CredentialsSpec) error {
	gen, err := credentials.ToGenericCredentialsSpec(creds)
	if err != nil {
		return err
	}
	spec := &ConsumerSpec{
		Identity:    id,
		Credentials: gen,
	}
	a.Consumers = append(a.Consumers, *spec)
	return nil
}

func (a *ConfigSpec) AddRepository(repo credentials.RepositorySpec, creds credentials.CredentialsSpec) error {
	rgen, err := credentials.ToGenericRepositorySpec(repo)
	if err != nil {
		return err
	}
	cgen, err := credentials.ToGenericCredentialsSpec(creds)
	if err != nil {
		return err
	}
	spec := &RepositorySpec{
		Repository:  *rgen,
		Credentials: cgen,
	}
	a.Repositories = append(a.Repositories, *spec)
	return nil
}

func (a *ConfigSpec) ApplyTo(ctx cpi.Context, target interface{}) error {
	return nil
}
