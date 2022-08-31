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

package ocm

import (
	"sync"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/core"
	"github.com/open-component-model/ocm/pkg/errors"
)

type CompoundResolver struct {
	lock      sync.RWMutex
	resolvers []ComponentVersionResolver
}

var _ ComponentVersionResolver = (*CompoundResolver)(nil)

func NewCompoundResolver(res ...ComponentVersionResolver) ComponentVersionResolver {
	for i := 0; i < len(res); i++ {
		if res[i] == nil {
			res = append(res[:i], res[i+1:]...)
			i--
		}
	}
	if len(res) == 1 {
		return res[0]
	}
	return &CompoundResolver{resolvers: res}
}

func (c *CompoundResolver) LookupComponentVersion(name string, version string) (core.ComponentVersionAccess, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	for _, r := range c.resolvers {
		if r == nil {
			continue
		}
		cv, err := r.LookupComponentVersion(name, version)
		if err == nil && cv != nil {
			return cv, nil
		}
		if !errors.IsErrNotFoundKind(err, KIND_COMPONENTVERSION) {
			return nil, err
		}
	}
	return nil, errors.ErrNotFound(KIND_OCM_REFERENCE, common.NewNameVersion(name, version).String())
}

func (c *CompoundResolver) AddResolver(r ComponentVersionResolver) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.resolvers = append(c.resolvers, r)
}

type sessionBasedResolver struct {
	session    Session
	repository Repository
}

func NewSessionBasedResolver(session Session, repo Repository) ComponentVersionResolver {
	return &sessionBasedResolver{session, repo}
}

func (c *sessionBasedResolver) LookupComponentVersion(name string, version string) (core.ComponentVersionAccess, error) {
	return c.session.LookupComponentVersion(c.repository, name, version)
}
