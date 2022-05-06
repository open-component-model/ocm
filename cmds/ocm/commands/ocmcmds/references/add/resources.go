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
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
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

func (r *ResourceSpec) Validate(ctx clictx.Context, input *common.ResourceInput) error {
	allErrs := field.ErrorList{}
	var fldPath *field.Path

	if err := compdescv2.ValidateComponentReference(fldPath, r.ComponentReference); err != nil {
		allErrs = append(allErrs, err...)
	}
	return allErrs.ToAggregate()
}
