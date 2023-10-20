// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package composition

import (
	"sync"

	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

const ATTR_REPOS = "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/composition"

type Repositories struct {
	lock  sync.Mutex
	repos map[string]cpi.Repository
}

func newRepositories(datacontext.Context) interface{} {
	return &Repositories{
		repos: map[string]cpi.Repository{},
	}
}

func (r *Repositories) GetRepository(name string) cpi.Repository {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.repos[name]
}

func (r *Repositories) SetRepository(name string, repo cpi.Repository) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.repos[name] = repo
}
