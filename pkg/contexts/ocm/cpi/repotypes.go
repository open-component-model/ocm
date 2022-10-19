// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"reflect"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/core"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type DefaultRepositoryType struct {
	runtime.ObjectVersionedType
	runtime.TypedObjectDecoder
	checker RepositoryAccessMethodChecker
}

type RepositoryAccessMethodChecker func(core.Context, compdesc.AccessSpec) bool

func NewRepositoryType(name string, proto core.RepositorySpec, checker RepositoryAccessMethodChecker) core.RepositoryType {
	t := reflect.TypeOf(proto)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return &DefaultRepositoryType{
		ObjectVersionedType: runtime.NewVersionedObjectType(name),
		TypedObjectDecoder:  runtime.MustNewDirectDecoder(proto),
		checker:             checker,
	}
}

func (t *DefaultRepositoryType) LocalSupportForAccessSpec(ctx core.Context, a compdesc.AccessSpec) bool {
	if t.checker != nil {
		return t.checker(ctx, a)
	}
	return false
}
