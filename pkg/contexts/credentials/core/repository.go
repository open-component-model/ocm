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

package core

import (
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/open-component-model/ocm/pkg/common"
)

const DirectCredentialsType = "Credentials"

type Repository interface {
	ExistsCredentials(name string) (bool, error)
	LookupCredentials(name string) (Credentials, error)
	WriteCredentials(name string, creds Credentials) (Credentials, error)
}

type Credentials interface {
	CredentialsSource
	ExistsProperty(name string) bool
	GetProperty(name string) string
	PropertyNames() sets.String
	Properties() common.Properties
}

type DirectCredentials struct {
	Values common.Properties `json: "properties"`
}

var _ Credentials = &DirectCredentials{}

func NewCredentials(props common.Properties) *DirectCredentials {
	if props == nil {
		props = common.Properties{}
	} else {
		props = props.Copy()
	}
	return &DirectCredentials{
		Values: props,
	}
}

func (c *DirectCredentials) GetType() string {
	return DirectCredentialsType
}

func (c *DirectCredentials) ExistsProperty(name string) bool {
	_, ok := c.Values[name]
	return ok
}

func (c *DirectCredentials) GetProperty(name string) string {
	return c.Values[name]
}

func (c *DirectCredentials) PropertyNames() sets.String {
	return c.Values.Names()
}

func (c *DirectCredentials) Properties() common.Properties {
	return c.Values.Copy()
}

func (c *DirectCredentials) Credentials(Context, ...CredentialsSource) (Credentials, error) {
	return c, nil
}

func (c *DirectCredentials) Copy() *DirectCredentials {
	return &DirectCredentials{
		Values: c.Values.Copy(),
	}
}
