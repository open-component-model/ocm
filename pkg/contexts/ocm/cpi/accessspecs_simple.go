// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"strings"

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
	description string
	handler     flagsets.ConfigOptionTypeSetHandler
}

func NewAccessSpecType(name string, proto core.AccessSpec, desc string, handler ...flagsets.ConfigOptionTypeSetHandler) core.AccessType {
	return &accessType{
		ObjectVersionedType: runtime.NewVersionedObjectType(name),
		TypedObjectDecoder:  runtime.MustNewDirectDecoder(proto),
		description:         desc,
		handler:             _handler(handler),
	}
}

func (t *accessType) ConfigOptionTypeSetHandler() flagsets.ConfigOptionTypeSetHandler {
	return t.handler
}

func (t *accessType) Description(cli bool) string {
	group := ""
	if t.handler != nil && cli {
		opts := t.handler.OptionTypeNames()
		var names []string
		if len(opts) > 0 {
			for _, o := range opts {
				names = append(names, "<code>--"+o+"</code>")
			}
			group = "\nOptions used to configure fields: " + strings.Join(names, ", ")
		}
	}
	return t.description + group
}
