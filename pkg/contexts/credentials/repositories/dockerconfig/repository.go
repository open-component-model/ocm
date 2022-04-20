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
	"os"
	"strings"
	"sync"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	dockercred "github.com/docker/cli/cli/config/credentials"
	"github.com/docker/cli/cli/config/types"
	"github.com/open-component-model/ocm/pkg/common"
	cpi2 "github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/oci/identity"
	"github.com/open-component-model/ocm/pkg/errors"
)

type Repository struct {
	lock      sync.RWMutex
	ctx       cpi2.Context
	propagate bool
	path      string
	config    *configfile.ConfigFile
}

func NewRepository(ctx cpi2.Context, path string, propagate bool) (*Repository, error) {
	r := &Repository{
		ctx:       ctx,
		propagate: propagate,
		path:      path,
	}
	err := r.Read(true)
	return r, err
}

var _ cpi2.Repository = &Repository{}

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

func (r Repository) LookupCredentials(name string) (cpi2.Credentials, error) {
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

func (r Repository) WriteCredentials(name string, creds cpi2.Credentials) (cpi2.Credentials, error) {
	return nil, errors.ErrNotSupported("write", "credentials", DockerConfigRepositoryType)
}

func (r *Repository) Read(force bool) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if !force && r.config != nil {
		return nil
	}
	path := r.path
	if strings.HasPrefix(path, "~/") {
		home := os.Getenv("HOME")
		path = home + path[1:]
	}
	data, err := ioutil.ReadFile(path)
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
			hostname := dockercred.ConvertToHostname(h)
			if hostname == "index.docker.io" {
				hostname = "docker.io"
			}
			id := cpi2.ConsumerIdentity{
				cpi2.ATTR_TYPE:       identity.VALUE_TYPE,
				identity.ID_HOSTNAME: hostname,
			}
			//fmt.Printf("propgate id %s\n", id)
			r.ctx.SetCredentialsForConsumer(id, newCredentials(a))
		}
	}
	r.config = cfg
	return nil
}

func newCredentials(auth types.AuthConfig) cpi2.Credentials {
	props := common.Properties{
		cpi2.ATTR_USERNAME: auth.Username,
		cpi2.ATTR_PASSWORD: auth.Password,
	}
	props.SetNonEmptyValue("auth", auth.Auth)
	props.SetNonEmptyValue(cpi2.ATTR_SERVER_ADDRESS, auth.ServerAddress)
	props.SetNonEmptyValue(cpi2.ATTR_IDENTITY_TOKEN, auth.IdentityToken)
	props.SetNonEmptyValue(cpi2.ATTR_REGISTRY_TOKEN, auth.RegistryToken)
	return cpi2.NewCredentials(props)
}
