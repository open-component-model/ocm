// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package srcs

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	compdescv2 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/v2"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type ResourceSpecHandler struct{}

var _ common.ResourceSpecHandler = (*ResourceSpecHandler)(nil)

func (ResourceSpecHandler) Key() string {
	return "source"
}

func (ResourceSpecHandler) RequireInputs() bool {
	return true
}

func (ResourceSpecHandler) Decode(data []byte) (addhdlrs.ElementSpec, error) {
	var desc ResourceSpec
	err := runtime.DefaultYAMLEncoding.Unmarshal(data, &desc)
	if err != nil {
		return nil, err
	}
	return &desc, nil
}

func (ResourceSpecHandler) Set(v ocm.ComponentVersionAccess, r addhdlrs.Element, acc compdesc.AccessSpec) error {
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
	compdescv2.SourceMeta `json:",inline"`

	addhdlrs.ResourceInput `json:",inline"`
}

var _ addhdlrs.ElementSpec = (*ResourceSpec)(nil)

func (r *ResourceSpec) Info() string {
	return fmt.Sprintf("source %s: %s", r.Type, r.GetRawIdentity())
}

func (r *ResourceSpec) Validate(ctx clictx.Context, input *addhdlrs.ResourceInput) error {
	allErrs := field.ErrorList{}
	var fldPath *field.Path

	src := compdescv2.Source{
		SourceMeta: r.SourceMeta,
	}
	if err := compdescv2.ValidateSource(fldPath, src, false); err != nil {
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
			} else if acc.(ocm.AccessSpec).IsLocal(ctx.OCMContext()) {
				kind := runtime.ObjectVersionedType(r.Access.ObjectType).GetKind()
				allErrs = append(allErrs, field.Invalid(fldPath.Child("access", "type"), kind, "local access no possible"))
			}
		}
	}
	return allErrs.ToAggregate()
}
