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

package add

import (
	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	compdesc2 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/v2"
	"github.com/open-component-model/ocm/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type ResourceSpecHandler struct{}

var _ common.ResourceSpecHandler = (*ResourceSpecHandler)(nil)

func (ResourceSpecHandler) Decode(data []byte) (common.ResourceSpec, error) {
	var desc ResourceSpec
	err := runtime.DefaultYAMLEncoding.Unmarshal(data, &desc)
	if err != nil {
		return nil, err
	}
	return &desc, nil
}

func (ResourceSpecHandler) Set(v ocm.ComponentVersionAccess, r common.Resource, acc compdesc.AccessSpec) error {
	spec := r.Spec().(*ResourceSpec)
	vers := spec.Version
	if vers == "" {
		vers = v.GetVersion()
	}
	meta := &compdesc.SourceMeta{
		ElementMeta: compdesc.ElementMeta{
			Name:          spec.Name,
			Version:       vers,
			ExtraIdentity: spec.ExtraIdentity,
			Labels:        spec.Labels,
		},
		Type: spec.Type,
	}
	return v.SetSource(meta, acc)
}

////////////////////////////////////////////////////////////////////////////////

type ResourceSpec struct {
	compdesc2.Source `json:",inline"`
}

var _ common.ResourceSpec = (*ResourceSpec)(nil)

func (r *ResourceSpec) Validate(ctx clictx.Context, input *common.ResourceInput) error {
	allErrs := field.ErrorList{}
	var fldPath *field.Path

	if err := compdesc2.ValidateSource(fldPath, r.Source, false); err != nil {
		allErrs = append(allErrs, err...)
	}
	if r.Access != nil {
		if r.Access.GetType() == "" {
			allErrs = append(allErrs, field.Required(fldPath.Child("access", "type"), "type of access required"))
		} else {
			acc, err := r.Access.Evaluate(ctx.OCMContext().AccessMethods())
			if err != nil {
				raw, _ := r.Access.GetRaw()
				allErrs = append(allErrs, field.Invalid(fldPath.Child("access"), string(raw), err.Error()))
			} else {
				if acc.(ocm.AccessSpec).IsLocal(ctx.OCMContext()) {
					kind := runtime.ObjectVersionedType(r.Access.ObjectType).GetKind()
					allErrs = append(allErrs, field.Invalid(fldPath.Child("access", "type"), kind, "local access no possible"))
				}
			}
		}
	}
	return allErrs.ToAggregate()
}
