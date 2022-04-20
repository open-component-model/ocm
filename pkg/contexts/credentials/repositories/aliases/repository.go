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

package aliases

import (
	"sync"

	"github.com/open-component-model/ocm/pkg/contexts/credentials/core"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
)

type Repository struct {
	sync.Mutex
	name  string
	spec  cpi.RepositorySpec
	creds cpi.CredentialsSource
	repo  cpi.Repository
}

func (a *Repository) GetRepository(ctx cpi.Context, creds cpi.Credentials) (cpi.Repository, error) {
	a.Lock()
	defer a.Unlock()
	if a.repo != nil {
		return a.repo, nil
	}

	src := core.CredentialsChain{}
	if a.creds != nil {
		src = append(src, a.creds)
	}
	if creds != nil {
		src = append(src, creds)
	}
	repo, err := ctx.RepositoryForSpec(a.spec, src...)
	if err != nil {
		return nil, err
	}
	a.repo = repo
	return repo, nil
}

func NewRepository(name string, spec cpi.RepositorySpec, creds cpi.Credentials) *Repository {
	return &Repository{
		name:  name,
		spec:  spec,
		creds: creds,
	}
}
