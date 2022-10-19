// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	compdescv2 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/v2"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type ResourceSpecHandler struct{}

var _ common.ResourceSpecHandler = (*ResourceSpecHandler)(nil)

func (ResourceSpecHandler) RequireInputs() bool {
	return false
}

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
	meta := &compdesc.ComponentReference{
		ElementMeta: compdesc.ElementMeta{
			Name:          spec.Name,
			Version:       vers,
			ExtraIdentity: spec.ExtraIdentity,
			Labels:        spec.Labels,
		},
		ComponentName: spec.ComponentName,
	}
	return v.SetReference(meta)
}

////////////////////////////////////////////////////////////////////////////////

type ResourceSpec struct {
	compdescv2.ComponentReference `json:",inline"`
}

var _ common.ResourceSpec = (*ResourceSpec)(nil)

func (r *ResourceSpec) Info() string {
	return fmt.Sprintf("reference %s: %s", r.ComponentName, r.GetRawIdentity())
}

func (r *ResourceSpec) Validate(ctx clictx.Context, input *common.ResourceInput) error {
	allErrs := field.ErrorList{}
	var fldPath *field.Path

	if err := compdescv2.ValidateComponentReference(fldPath, r.ComponentReference); err != nil {
		allErrs = append(allErrs, err...)
	}
	return allErrs.ToAggregate()
}
