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

package config

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/common"
	cfgcpi "github.com/open-component-model/ocm/pkg/contexts/config/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/memory"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const (
	ConfigType   = "memory.credentials.config" + common.TypeGroupSuffix
	ConfigTypeV1 = ConfigType + runtime.VersionSeparator + "v1"
)

func init() {
	cfgcpi.RegisterConfigType(ConfigType, cfgcpi.NewConfigType(ConfigType, &Config{}, usage))
	cfgcpi.RegisterConfigType(ConfigTypeV1, cfgcpi.NewConfigType(ConfigTypeV1, &Config{}, usage))
}

// Config describes a configuration for the config context
type Config struct {
	runtime.ObjectVersionedType `json:",inline"`
	RepoName                    string            `json:"repoName"`
	Credentials                 []CredentialsSpec `json:"credentials,omitempty"`
}

type CredentialsSpec struct {
	CredentialsName string `json:"credentialsName"`
	// Reference refers to credentials store in some othe repo
	Reference *cpi.GenericCredentialsSpec `json:"reference,omitempty"`
	// Credentials are direct credentials (one of Reference or Credentails must be set)
	Credentails common.Properties `json:"credentials"`
}

// New creates a new memory ConfigSpec
func New(repo string, credentials ...CredentialsSpec) *Config {
	return &Config{
		ObjectVersionedType: runtime.NewVersionedObjectType(ConfigType),
		RepoName:            repo,
		Credentials:         credentials,
	}
}

func (a *Config) GetType() string {
	return ConfigType
}

func (a *Config) AddCredentials(name string, props common.Properties) error {
	a.Credentials = append(a.Credentials, CredentialsSpec{CredentialsName: name, Credentails: props})
	return nil
}

func (a *Config) AddCredentialsRef(name string, refname string, spec cpi.RepositorySpec) error {
	repo, err := cpi.ToGenericRepositorySpec(spec)
	if err != nil {
		return fmt.Errorf("unable to convert cpi repository spec to generic: %w", err)
	}

	ref := cpi.NewGenericCredentialsSpec(refname, repo)
	a.Credentials = append(a.Credentials, CredentialsSpec{CredentialsName: name, Reference: ref})

	return nil
}

func (a *Config) ApplyTo(ctx cfgcpi.Context, target interface{}) error {
	list := errors.ErrListf("applying config")

	t, ok := target.(cpi.Context)
	if !ok {
		return cfgcpi.ErrNoContext(ConfigType)
	}

	repo, err := t.RepositoryForSpec(memory.NewRepositorySpec(a.RepoName))
	if err != nil {
		return fmt.Errorf("unabel to get repository for spec: %w", err)
	}

	mem := repo.(*memory.Repository)

	for i, e := range a.Credentials {
		var creds cpi.Credentials
		if e.Reference != nil {
			if len(e.Credentails) != 0 {
				err = fmt.Errorf("credentials and reference set")
			} else {
				creds, err = e.Reference.Credentials(t)

			}
		} else {
			creds = cpi.NewCredentials(e.Credentails)
		}
		if err != nil {
			list.Add(errors.Wrapf(err, "config entry %d[%s]", i, e.CredentialsName))
		}
		if creds != nil {
			_, err = mem.WriteCredentials(e.CredentialsName, creds)
			if err != nil {
				list.Add(errors.Wrapf(err, "config entry %d", i))
			}
		}
	}
	return list.Result()
}

const usage = `
The config type <code>` + ConfigType + `</code> can be used to define a list
of arbitrary credentials stored in a memory based credentials repository:

<pre>
    type: ` + ConfigType + `
    repoName: default
    credentials:
      - credentialsName: ref
        reference:  # refer to a credential set stored in some other credential repository
          type: Credentials # this is a repo providing just one explicit credential set
          properties:
            username: mandelsoft
            password: specialsecret
      - credentialsName: direct
        credentials: # direct credential specification
            username: mandelsoft2
            password: specialsecret2
</pre>
`
