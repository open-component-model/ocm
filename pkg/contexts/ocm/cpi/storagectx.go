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

package cpi

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/core"
)

type DefaultStorageContext struct {
	ComponentRepository          Repository
	ComponentVersion             ComponentVersionAccess
	ImplementationRepositoryType ImplementationRepositoryType
}

var _ StorageContext = (*DefaultStorageContext)(nil)

func NewDefaultStorageContext(repo Repository, vers ComponentVersionAccess, reptype ImplementationRepositoryType) *DefaultStorageContext {
	return &DefaultStorageContext{
		ComponentRepository:          repo,
		ComponentVersion:             vers,
		ImplementationRepositoryType: reptype,
	}
}

func (c *DefaultStorageContext) GetContext() core.Context {
	return c.ComponentRepository.GetContext()
}

func (c *DefaultStorageContext) TargetComponentVersion() core.ComponentVersionAccess {
	return c.ComponentVersion
}

func (c *DefaultStorageContext) TargetComponentRepository() core.Repository {
	return c.ComponentRepository
}

func (c *DefaultStorageContext) GetImplementationRepositoryType() ImplementationRepositoryType {
	return c.ImplementationRepositoryType
}
