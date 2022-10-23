// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/core"
	"github.com/open-component-model/ocm/pkg/runtime"
)

func _handler(handler []flagsets.ConfigOptionTypeSetHandler) flagsets.ConfigOptionTypeSetHandler {
	if len(handler) > 0 {
		return handler[0]
	}
	return nil
}

type accessType struct {
	runtime.ObjectVersionedType
	runtime.TypedObjectDecoder
	handler flagsets.ConfigOptionTypeSetHandler
}

func NewAccessSpecType(name string, proto core.AccessSpec, handler ...flagsets.ConfigOptionTypeSetHandler) core.AccessType {
	return &accessType{
		ObjectVersionedType: runtime.NewVersionedObjectType(name),
		TypedObjectDecoder:  runtime.MustNewDirectDecoder(proto),
		handler:             _handler(handler),
	}
}

func (t *accessType) ConfigOptionTypeSetHandler() flagsets.ConfigOptionTypeSetHandler {
	return t.handler
}
