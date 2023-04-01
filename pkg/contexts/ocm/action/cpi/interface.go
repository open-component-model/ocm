// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"reflect"

	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/runtime/scheme"
)

type Action interface {
	Name() string
	Description() string
	SpecificationProto() reflect.Type
	ResultProto() reflect.Type
}

////////////////////////////////////////////////////////////////////////////////
// Action Specification

type Selector string

type ActionSpec interface {
	runtime.VersionedTypedObject
	Selector() Selector
}

type ActionSpecType scheme.Type[ActionSpec]

////////////////////////////////////////////////////////////////////////////////
// Action Result

type ActionResult interface {
	runtime.VersionedTypedObject
	Error() string
}

// CommonResult is the minimal action result.
type CommonResult struct { //nolint: errname // general return status
	runtime.ObjectVersionedType `json:",inline"`
	ErrorMessage                string `json:"error,omitempty"`
}

func (r *CommonResult) Error() string {
	return r.ErrorMessage
}

////////////////////////////////////////////////////////////////////////////////
// Action Type

type ActionResultType scheme.Type[ActionResult]

type ActionType interface {
	SpecificationType() ActionSpecType
	ResultType() ActionResultType
}
