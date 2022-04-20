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
	"sync"

	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
)

const ATTR_REPOS = "github.com/open-component-model/ocm/pkg/credentials/repositories/dockerconfig"

type Repositories struct {
	lock  sync.Mutex
	repos map[string]*Repository
}

func newRepositories(datacontext.Context) interface{} {
	return &Repositories{
		repos: map[string]*Repository{},
	}
}

func (r *Repositories) GetRepository(ctx cpi.Context, name string, propagate bool) (*Repository, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	var err error = nil
	repo := r.repos[name]
	if repo == nil {
		repo, err = NewRepository(ctx, name, propagate)
		r.repos[name] = repo
	}
	return repo, err
}
