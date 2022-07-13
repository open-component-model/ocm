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
	"os"
	"strings"
	"sync"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	dockercred "github.com/docker/cli/cli/config/credentials"
	"github.com/docker/cli/cli/config/types"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/oci/identity"
	"github.com/open-component-model/ocm/pkg/errors"
)

var log = false

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

func (r *Repository) LookupCredentials(name string) (cpi.Credentials, error) {
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

func (r *Repository) WriteCredentials(name string, creds cpi.Credentials) (cpi.Credentials, error) {
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
	defaultStore := dockercred.DetectDefaultStore(cfg.CredentialsStore)
	store := dockercred.NewNativeStore(cfg, defaultStore)
	// get default native credential store
	if r.propagate {
		all := cfg.GetAuthConfigs()
		for h, a := range all {
			hostname := dockercred.ConvertToHostname(h)
			if hostname == "index.docker.io" {
				hostname = "docker.io"
			}
			id := cpi.ConsumerIdentity{
				cpi.ATTR_TYPE:        identity.CONSUMER_TYPE,
				identity.ID_HOSTNAME: hostname,
			}

			var creds cpi.Credentials
			if IsEmptyAuthConfig(a) {
				if log {
					fmt.Printf("propagate id %q with default store %q\n", id, defaultStore)
				}
				creds = NewCredentials(r, h, store)
			} else {
				if log {
					fmt.Printf("propagate id %q\n", id)
				}
				creds = newCredentials(a)
			}
			r.ctx.SetCredentialsForConsumer(id, creds)
		}
		for h, helper := range cfg.CredentialHelpers {
			hostname := dockercred.ConvertToHostname(h)
			if hostname == "index.docker.io" {
				hostname = "docker.io"
			}
			id := cpi.ConsumerIdentity{
				cpi.ATTR_TYPE:        identity.CONSUMER_TYPE,
				identity.ID_HOSTNAME: hostname,
			}
			if log {
				fmt.Printf("propagate id %s with helper %s\n", id, helper)
			}
			r.ctx.SetCredentialsForConsumer(id, NewCredentials(r, h, dockercred.NewNativeStore(cfg, helper)))
		}
	}
	r.config = cfg
	return nil
}

func newCredentials(auth types.AuthConfig) cpi.Credentials {
	props := common.Properties{
		cpi.ATTR_USERNAME: auth.Username,
		cpi.ATTR_PASSWORD: auth.Password,
	}
	props.SetNonEmptyValue("auth", auth.Auth)
	props.SetNonEmptyValue(cpi.ATTR_SERVER_ADDRESS, auth.ServerAddress)
	props.SetNonEmptyValue(cpi.ATTR_IDENTITY_TOKEN, auth.IdentityToken)
	props.SetNonEmptyValue(cpi.ATTR_REGISTRY_TOKEN, auth.RegistryToken)
	return cpi.NewCredentials(props)
}

// IsEmptyAuthConfig validates if the resulting auth config contains credentails
func IsEmptyAuthConfig(auth types.AuthConfig) bool {
	if len(auth.Auth) != 0 {
		return false
	}
	if len(auth.Username) != 0 {
		return false
	}
	return true
}
