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

package oci

import (
	"context"
	"reflect"

	"github.com/gardener/ocm/pkg/runtime"
)

type Context interface {
	context.Context
	RepositoryTypes() RepositoryTypeScheme

	RepositoryForSpec(spec RepositorySpec) (Repository, error)
	RepositoryForConfig(data []byte, unmarshaler runtime.Unmarshaler) (Repository, error)
}

type _context struct {
	context.Context
	knownRepositoryTypes RepositoryTypeScheme
}

var key = reflect.TypeOf(_context{})

func NewDefaultContext(ctx context.Context) Context {
	c := &_context{
		knownRepositoryTypes: DefaultRepositoryTypeScheme,
	}
	c.Context = context.WithValue(ctx, key, c)
	return c
}

func RepositoryContext(ctx context.Context) Context {
	c := ctx.Value(key)
	if c == nil {
		return nil
	}
	return c.(Context)
}

func (c *_context) RepositoryTypes() RepositoryTypeScheme {
	return c.knownRepositoryTypes
}

func (c *_context) RepositoryForSpec(spec RepositorySpec) (Repository, error) {
	return spec.Repository(c)
}

func (c *_context) RepositoryForConfig(data []byte, unmarshaler runtime.Unmarshaler) (Repository, error) {
	spec, err := c.knownRepositoryTypes.DecodeRepositorySpec(data, unmarshaler)
	if err != nil {
		return nil, err
	}
	return spec.Repository(c)
}
