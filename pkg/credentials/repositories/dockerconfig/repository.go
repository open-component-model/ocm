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
	"io/ioutil"
	"sync"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/gardener/ocm/pkg/common"
	"github.com/gardener/ocm/pkg/errors"

	"github.com/gardener/ocm/pkg/credentials/cpi"
)

type Repository struct {
	lock   sync.RWMutex
	path   string
	config *configfile.ConfigFile
}

func NewRepository(path string) (*Repository, error) {
	r := &Repository{
		path: path,
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
	props := common.Properties{
		"username": auth.Username,
		"password": auth.Password,
	}
	props.SetNonEmptyValue("auth", auth.Auth)
	props.SetNonEmptyValue("serverAddress", auth.ServerAddress)
	props.SetNonEmptyValue("identityToken", auth.IdentityToken)
	props.SetNonEmptyValue("registryToken", auth.RegistryToken)
	return cpi.NewCredentials(props), nil

	return nil, cpi.ErrUnknownCredentials(name)
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

	r.config = cfg
	return nil
}
