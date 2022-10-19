// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/core"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type accessType struct {
	runtime.ObjectVersionedType
	runtime.TypedObjectDecoder
}

func NewAccessSpecType(name string, proto core.AccessSpec) core.AccessType {
	return &accessType{
		ObjectVersionedType: runtime.NewVersionedObjectType(name),
		TypedObjectDecoder:  runtime.MustNewDirectDecoder(proto),
	}
}
