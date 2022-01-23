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

package memory

import (
	"sync"

	"github.com/gardener/ocm/pkg/credentials/cpi"
)

type Repository struct {
	lock        sync.RWMutex
	name        string
	credentials map[string]cpi.Credentials
}

func NewRepository(name string) *Repository {
	return &Repository{
		name:        name,
		credentials: map[string]cpi.Credentials{},
	}
}

var _ cpi.Repository = &Repository{}

func (r *Repository) ExistsCredentials(name string) (bool, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	_, ok := r.credentials[name]
	return ok, nil
}

func (r Repository) LookupCredentials(name string) (cpi.Credentials, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	c, ok := r.credentials[name]
	if ok {
		return cpi.NewCredentials(c.Properties()), nil
	}
	return nil, cpi.ErrUnknownCredentials(name)
}

func (r Repository) WriteCredentials(name string, creds cpi.Credentials) (cpi.Credentials, error) {
	c := cpi.NewCredentials(creds.Properties())
	r.lock.Lock()
	defer r.lock.Unlock()
	r.credentials[name] = c
	return c, nil
}
