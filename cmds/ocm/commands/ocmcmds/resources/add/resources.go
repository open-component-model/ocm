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
	"github.com/gardener/ocm/cmds/ocm/clictx"
	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds"
	"github.com/gardener/ocm/pkg/ocm"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
	compdescv2 "github.com/gardener/ocm/pkg/ocm/compdesc/versions/v2"
	"github.com/gardener/ocm/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type Resources struct {
	*ResourceOptionList `json:",inline"`
	*ResourceOptions    `json:",inline"`
}

// ResourceOptions contains options that are used to describe a resource
type ResourceOptions struct {
	compdescv2.Resource `json:",inline"`
	Input               *ocmcmds.BlobInput `json:"input,omitempty"`
}

// ResourceOptionList contains a list of options that are used to describe a resource.
type ResourceOptionList struct {
	Resources []*ResourceOptions `json:"resources"`
}

func Validate(r *ResourceOptions, ctx clictx.Context, inputFilePath string) error {
	allErrs := field.ErrorList{}
	var fldPath *field.Path

	if r.Relation == "" {
		if r.Input != nil {
			r.Relation = metav1.LocalRelation
		}
		if r.Access != nil {
			r.Relation = metav1.ExternalRelation
		}
	}
	if r.Version == "" && r.Relation == metav1.LocalRelation {
		r.Version = "<componentversion>"
	}
	if err := compdescv2.ValidateResource(fldPath, r.Resource, false); err != nil {
		allErrs = append(allErrs, err...)
	}
	if r.Input != nil && r.Access != nil {
		allErrs = append(allErrs, field.Forbidden(fldPath, "only either input or access might be specified"))
	} else {
		if r.Input == nil && r.Access == nil {
			allErrs = append(allErrs, field.Forbidden(fldPath, "either input or access must be specified"))
		}
		if r.Access != nil {
			if r.Relation == metav1.LocalRelation {
				allErrs = append(allErrs, field.Forbidden(fldPath.Child("relation"), "access required external relation"))
			}
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
		if r.Input != nil {
			if r.Relation != metav1.LocalRelation {
				allErrs = append(allErrs, field.Forbidden(fldPath.Child("relation"), "input requires local relation"))
			}
			if err := ocmcmds.ValidateBlobInput(fldPath.Child("input"), r.Input, ctx.FileSystem(), inputFilePath); err != nil {
				allErrs = append(allErrs, err...)
			}
		}
	}
	return allErrs.ToAggregate()
}
