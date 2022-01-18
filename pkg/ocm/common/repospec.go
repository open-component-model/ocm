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

package common

import (
	"reflect"

	"github.com/gardener/ocm/pkg/ocm/compdesc"
	"github.com/gardener/ocm/pkg/ocm/core"
	"github.com/gardener/ocm/pkg/ocm/runtime"
)

type RepositoryType struct {
	runtime.ObjectTypeVersion
	runtime.TypedObjectDecoder
	checker RepositoryAccessMethodChecker
}

type RepositoryAccessMethodChecker func(core.Context, compdesc.AccessSpec) bool

func NewRepositoryType(name string, proto core.RepositorySpec, checker RepositoryAccessMethodChecker) core.RepositoryType {
	t := reflect.TypeOf(proto)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return &RepositoryType{
		ObjectTypeVersion:  runtime.NewObjectTypeVersion(name),
		TypedObjectDecoder: runtime.MustNewDirectDecoder(proto),
		checker:            checker,
	}
}

func (t *RepositoryType) LocalSupportForAccessSpec(ctx core.Context, a compdesc.AccessSpec) bool {
	if t.checker != nil {
		return t.checker(ctx, a)
	}
	return false
}
