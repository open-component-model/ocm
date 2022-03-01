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
	"bytes"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	dockercred "github.com/docker/cli/cli/config/credentials"
	"github.com/docker/cli/cli/config/types"
	"github.com/gardener/ocm/pkg/common"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/identity"

	"github.com/gardener/ocm/pkg/credentials/cpi"
)

type Repository struct {
	lock      sync.RWMutex
	ctx       cpi.Context
	propagate bool
	path      string
	config    *configfile.ConfigFile
}

func NewRepository(ctx cpi.Context, path string, propagate bool) (*Repository, error) {
	r := &Repository{
		ctx:       ctx,
		propagate: propagate,
		path:      path,
	}
	err := r.Read(true)
	return r, err
}

var _ cpi.Repository = &Repository{}

func (r *Repository) ExistsCredentials(name string) (bool, error) {
	err := r.Read(false)
	if err != nil {
		return false, err
	}
	r.lock.RLock()
	defer r.lock.RUnlock()

	_, err = r.config.GetAuthConfig(name)
	return err != nil, err
}

func (r Repository) LookupCredentials(name string) (cpi.Credentials, error) {
	err := r.Read(false)
	if err != nil {
		return nil, err
	}
	r.lock.RLock()
	defer r.lock.RUnlock()

	auth, err := r.config.GetAuthConfig(name)
	if err != nil {
		return nil, err
	}
	return newCredentials(auth), nil
}

func (r Repository) WriteCredentials(name string, creds cpi.Credentials) (cpi.Credentials, error) {
	return nil, errors.ErrNotSupported("write", "credentials", DockerConfigRepositoryType)
}

func (r *Repository) Read(force bool) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if !force && r.config != nil {
		return nil
	}
	data, err := ioutil.ReadFile(r.path)
	if err != nil {
		return err
	}

	cfg, err := config.LoadFromReader(bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	if r.propagate {
		all := cfg.GetAuthConfigs()
		for h, a := range all {
			id := cpi.ConsumerIdentity{
				identity.ID_HOSTNAME: dockercred.ConvertToHostname(h),
			}
			fmt.Printf("propgate id %s\n", id)
			r.ctx.SetCredentialsForConsumer(id, newCredentials(a))
		}
	}
	r.config = cfg
	return nil
}

func newCredentials(auth types.AuthConfig) cpi.Credentials {
	props := common.Properties{
		"username": auth.Username,
		"password": auth.Password,
	}
	props.SetNonEmptyValue("auth", auth.Auth)
	props.SetNonEmptyValue("serverAddress", auth.ServerAddress)
	props.SetNonEmptyValue("identityToken", auth.IdentityToken)
	props.SetNonEmptyValue("registryToken", auth.RegistryToken)
	return cpi.NewCredentials(props)
}
