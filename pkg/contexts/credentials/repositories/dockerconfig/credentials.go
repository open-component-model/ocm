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
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
)

type Credentials struct {
	repo *Repository
	name string
}

var _ cpi.Credentials = (*Credentials)(nil)

func (c *Credentials) get() common.Properties {
	auth, err := c.repo.config.GetAuthConfig(c.name)
	if err != nil {
		return common.Properties{}
	}
	return newCredentials(auth).Properties()
}

func (c *Credentials) Credentials(context cpi.Context, source ...cpi.CredentialsSource) (cpi.Credentials, error) {
	auth, err := c.repo.config.GetAuthConfig(c.name)
	if err != nil {
		return nil, err
	}
	return newCredentials(auth), nil
}

func (c *Credentials) ExistsProperty(name string) bool {
	_, ok := c.get()[name]
	return ok
}

func (c *Credentials) GetProperty(name string) string {
	return c.get()[name]
}

func (c *Credentials) PropertyNames() sets.String {
	return c.get().Names()
}

func (c *Credentials) Properties() common.Properties {
	return c.get()
}
