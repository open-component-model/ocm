// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/runtime/scheme"
)

type actionType struct {
	spectype ActionSpecType
	restype  ActionResultType
}

var _ ActionType = (*actionType)(nil)

func NewActionTypeByProtoTypes(specproto ActionSpec, specconv scheme.Converter[ActionSpec], resultproto ActionResult, resconv scheme.Converter[ActionResult]) ActionType {
	if specconv == nil {
		specconv = runtime.IdentityConverter[ActionSpec]{}
	}
	if resconv == nil {
		resconv = runtime.IdentityConverter[ActionResult]{}
	}
	st := scheme.NewTypeByProtoType(specproto, specconv)
	rt := scheme.NewTypeByProtoType(resultproto, resconv)
	return &actionType{
		spectype: st,
		restype:  rt,
	}
}

func (a *actionType) SpecificationType() ActionSpecType {
	return a.spectype
}

func (a *actionType) ResultType() ActionResultType {
	return a.restype
}
