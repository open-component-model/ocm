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
	"github.com/open-component-model/ocm/pkg/common"
	cfgcpi "github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	ConfigType   = "credentials.config" + common.TypeGroupSuffix
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cfgcpi.RegisterConfigType(ConfigType, cfgcpi.NewConfigType(ConfigType, &ConfigSpec{}))
	cfgcpi.RegisterConfigType(ConfigTypeV1, cfgcpi.NewConfigType(ConfigTypeV1, &ConfigSpec{}))
}

// ConfigSpec describes a memory based repository interface.
type ConfigSpec struct {
	runtime.ObjectVersionedType `json:",inline"`
	// Consumers describe predefine logical cosumer specs mapped to credentials
	// These will (potentially) be evaluated if access objects requiring crednetials
	// are provided by other modules (e.g. oci repo access) without
	// specifying crednentials. Then this module can request credentials here by passing
	// an appropriate consumer spec.
	Consumers []ConsumerSpec `json:"consumers,omitempty"`
	// Repositories describe preloaded credential repositories with potential credential chain
	Repositories []RepositorySpec `json:"repositories,omitempty"`
	// Aliases describe logical credential repositories mapped to implementig repositories
	Aliases map[string]RepositorySpec `json:"aliases,omitempty"`
}

type ConsumerSpec struct {
	Identity    cpi.ConsumerIdentity         `json:"identity"`
	Credentials []cpi.GenericCredentialsSpec `json:"credentials"`
}

type RepositorySpec struct {
	Repository  cpi.GenericRepositorySpec    `json:"repository"`
	Credentials []cpi.GenericCredentialsSpec `json:"credentials,omitempty"`
}

// NewConfigSpec creates a new memory ConfigSpec
func NewConfigSpec() *ConfigSpec {
	return &ConfigSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(ConfigType),
	}
}

func (a *ConfigSpec) GetType() string {
	return ConfigType
}

func (a *ConfigSpec) MapCredentialsChain(creds ...cpi.CredentialsSpec) ([]cpi.GenericCredentialsSpec, error) {
	var cgens []cpi.GenericCredentialsSpec
	for _, c := range creds {
		cgen, err := cpi.ToGenericCredentialsSpec(c)
		if err != nil {
			return nil, err
		}
		cgens = append(cgens, *cgen)
	}
	return cgens, nil
}

func (a *ConfigSpec) AddConsumer(id cpi.ConsumerIdentity, creds ...cpi.CredentialsSpec) error {
	cgens, err := a.MapCredentialsChain(creds...)
	if err != nil {
		return err
	}

	spec := &ConsumerSpec{
		Identity:    id,
		Credentials: cgens,
	}
	a.Consumers = append(a.Consumers, *spec)
	return nil
}

func (a *ConfigSpec) MapRepository(repo cpi.RepositorySpec, creds ...cpi.CredentialsSpec) (*RepositorySpec, error) {
	rgen, err := cpi.ToGenericRepositorySpec(repo)
	if err != nil {
		return nil, err
	}

	cgens, err := a.MapCredentialsChain(creds...)
	if err != nil {
		return nil, err
	}

	return &RepositorySpec{
		Repository:  *rgen,
		Credentials: cgens,
	}, nil
}

func (a *ConfigSpec) AddRepository(repo cpi.RepositorySpec, creds ...cpi.CredentialsSpec) error {
	spec, err := a.MapRepository(repo, creds...)
	if err != nil {
		return err
	}
	a.Repositories = append(a.Repositories, *spec)
	return nil
}

func (a *ConfigSpec) AddAlias(name string, repo cpi.RepositorySpec, creds ...cpi.CredentialsSpec) error {
	spec, err := a.MapRepository(repo, creds...)
	if err != nil {
		return err
	}

	if a.Aliases == nil {
		a.Aliases = map[string]RepositorySpec{}
	}
	a.Aliases[name] = *spec
	return nil
}

func (a *ConfigSpec) ApplyTo(ctx cfgcpi.Context, target interface{}) error {
	list := errors.ErrListf("applying config")
	t, ok := target.(cpi.Context)
	if !ok {
		return cfgcpi.ErrNoContext(ConfigType)
	}
	for _, e := range a.Consumers {
		t.SetCredentialsForConsumer(e.Identity, CredentialsChain(e.Credentials...))
	}
	sub := errors.ErrListf("applying aliases")
	for n, e := range a.Aliases {
		sub.Add(t.SetAlias(n, &e.Repository, CredentialsChain(e.Credentials...)))
	}
	list.Add(sub.Result())
	sub = errors.ErrListf("applying repositories")
	for i, e := range a.Repositories {
		_, err := t.RepositoryForSpec(&e.Repository, CredentialsChain(e.Credentials...))
		sub.Add(errors.Wrapf(err, "repository entry %d", i))
	}
	list.Add(sub.Result())

	return list.Result()
}

func CredentialsChain(creds ...cpi.GenericCredentialsSpec) cpi.CredentialsChain {
	r := make([]cpi.CredentialsSource, len(creds))
	for i := range creds {
		r[i] = &creds[i]
	}
	return r
}
