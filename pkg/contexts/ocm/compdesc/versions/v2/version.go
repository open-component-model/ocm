// Copyright 2020 Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compdesc

import (
	compdesc2 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/v2/jsonscheme"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const SchemaVersion = "v2"

func init() {
	compdesc2.RegisterScheme(&DescriptorVersion{})
}

type DescriptorVersion struct{}

func (v *DescriptorVersion) GetVersion() string {
	return SchemaVersion
}

func (v *DescriptorVersion) Decode(data []byte, opts *compdesc2.DecodeOptions) (interface{}, error) {
	var cd ComponentDescriptor
	if !opts.DisableValidation {
		if err := jsonscheme.Validate(data); err != nil {
			return nil, err
		}
	}
	var err error
	if opts.StrictMode {
		err = opts.Codec.DecodeStrict(data, &cd)
	} else {
		err = opts.Codec.Decode(data, &cd)
	}
	if err != nil {
		return nil, err
	}

	if err := DefaultComponent(&cd); err != nil {
		return nil, err
	}

	if !opts.DisableValidation {
		err = Validate(&cd)
		if err != nil {
			return nil, err
		}
	}
	return &cd, err
}

////////////////////////////////////////////////////////////////////////////////
// convert to internal version
////////////////////////////////////////////////////////////////////////////////

func (v *DescriptorVersion) ConvertTo(obj interface{}) (out *compdesc2.ComponentDescriptor, err error) {
	if obj == nil {
		return nil, nil
	}
	in, ok := obj.(*ComponentDescriptor)
	if !ok {
		return nil, errors.Newf("%T is no version v2 descriptor", obj)
	}

	defer compdesc2.CatchConversionError(&err)
	out = &compdesc2.ComponentDescriptor{
		Metadata: compdesc2.Metadata{in.Metadata.Version},
		ComponentSpec: compdesc2.ComponentSpec{
			ObjectMeta: compdesc2.ObjectMeta{
				Name:    in.Name,
				Version: in.Version,
				Labels:  in.Labels.Copy(),
			},
			RepositoryContexts:  in.RepositoryContexts.Copy(),
			Provider:            in.Provider,
			Sources:             convert_Sources_to(in.Sources),
			Resources:           convert_Resources_to(in.Resources),
			ComponentReferences: convert_ComponentReferences_to(in.ComponentReferences),
		},
	}
	return out, nil
}

func convert_ComponentReference_to(in *ComponentReference) *compdesc2.ComponentReference {
	if in == nil {
		return nil
	}
	out := &compdesc2.ComponentReference{
		ElementMeta:   *convert_ElementMeta_to(&in.ElementMeta),
		ComponentName: in.ComponentName,
	}
	return out
}

func convert_ComponentReferences_to(in []ComponentReference) compdesc2.ComponentReferences {
	if in == nil {
		return nil
	}
	out := make(compdesc2.ComponentReferences, len(in))
	for i, v := range in {
		out[i] = *convert_ComponentReference_to(&v)
	}
	return out
}

func convert_Source_to(in *Source) *compdesc2.Source {
	if in == nil {
		return nil
	}
	out := &compdesc2.Source{
		SourceMeta: compdesc2.SourceMeta{
			ElementMeta: *convert_ElementMeta_to(&in.ElementMeta),
			Type:        in.Type,
		},
		Access: compdesc2.GenericAccessSpec(in.Access.DeepCopy()),
	}
	return out
}

func convert_Sources_to(in Sources) compdesc2.Sources {
	if in == nil {
		return nil
	}
	out := make(compdesc2.Sources, len(in))
	for i, v := range in {
		out[i] = *convert_Source_to(&v)
	}
	return out
}

func convert_ElementMeta_to(in *ElementMeta) *compdesc2.ElementMeta {
	if in == nil {
		return nil
	}
	out := &compdesc2.ElementMeta{
		Name:          in.Name,
		Version:       in.Version,
		ExtraIdentity: in.ExtraIdentity.Copy(),
		Labels:        in.Labels.Copy(),
	}
	return out
}

func convert_Resource_to(in *Resource) *compdesc2.Resource {
	if in == nil {
		return nil
	}
	out := &compdesc2.Resource{
		ResourceMeta: compdesc2.ResourceMeta{
			ElementMeta: *convert_ElementMeta_to(&in.ElementMeta),
			Type:        in.Type,
			Relation:    in.Relation,
			SourceRef:   Convert_SourceRefs_to(in.SourceRef),
		},
		Access: compdesc2.GenericAccessSpec(in.Access),
	}
	return out
}

func convert_Resources_to(in Resources) compdesc2.Resources {
	if in == nil {
		return nil
	}
	out := make(compdesc2.Resources, len(in))
	for i, v := range in {
		out[i] = *convert_Resource_to(&v)
	}
	return out
}

func convert_SourceRef_to(in *SourceRef) *compdesc2.SourceRef {
	if in == nil {
		return nil
	}
	out := &compdesc2.SourceRef{
		IdentitySelector: in.IdentitySelector.Copy(),
		Labels:           in.Labels.Copy(),
	}
	return out
}

func Convert_SourceRefs_to(in []SourceRef) []compdesc2.SourceRef {
	if in == nil {
		return nil
	}
	out := make([]compdesc2.SourceRef, len(in))
	for i, v := range in {
		out[i] = *convert_SourceRef_to(&v)
	}
	return out
}

////////////////////////////////////////////////////////////////////////////////
// convert from internal version
////////////////////////////////////////////////////////////////////////////////

func (v *DescriptorVersion) ConvertFrom(in *compdesc2.ComponentDescriptor) (interface{}, error) {
	if in == nil {
		return nil, nil
	}
	out := &ComponentDescriptor{
		Metadata: metav1.Metadata{
			SchemaVersion,
		},
		ComponentSpec: ComponentSpec{
			ObjectMeta: ObjectMeta{
				Name:    in.Name,
				Version: in.Version,
				Labels:  in.Labels.Copy(),
			},
			RepositoryContexts:  in.RepositoryContexts.Copy(),
			Provider:            in.Provider,
			Sources:             convert_Sources_from(in.Sources),
			Resources:           convert_Resources_from(in.Resources),
			ComponentReferences: convert_ComponentReferences_from(in.ComponentReferences),
		},
	}
	if err := DefaultComponent(out); err != nil {
		return nil, err
	}
	return out, nil
}

func convert_ComponentReference_from(in *compdesc2.ComponentReference) *ComponentReference {
	if in == nil {
		return nil
	}
	out := &ComponentReference{
		ElementMeta:   *convert_ElementMeta_from(&in.ElementMeta),
		ComponentName: in.ComponentName,
	}
	return out
}

func convert_ComponentReferences_from(in []compdesc2.ComponentReference) []ComponentReference {
	if in == nil {
		return nil
	}
	out := make([]ComponentReference, len(in))
	for i, v := range in {
		out[i] = *convert_ComponentReference_from(&v)
	}
	return out
}

func convert_Source_from(in *compdesc2.Source) *Source {
	if in == nil {
		return nil
	}
	acc, err := runtime.ToUnstructuredTypedObject(in.Access)
	if err != nil {
		compdesc2.ThrowConversionError(err)
	}
	out := &Source{
		SourceMeta: SourceMeta{
			ElementMeta: *convert_ElementMeta_from(&in.ElementMeta),
			Type:        in.Type,
		},
		Access: acc,
	}
	return out
}

func convert_Sources_from(in compdesc2.Sources) Sources {
	if in == nil {
		return nil
	}
	out := make(Sources, len(in))
	for i, v := range in {
		out[i] = *convert_Source_from(&v)
	}
	return out
}

func convert_ElementMeta_from(in *compdesc2.ElementMeta) *ElementMeta {
	if in == nil {
		return nil
	}
	out := &ElementMeta{
		Name:          in.Name,
		Version:       in.Version,
		ExtraIdentity: in.ExtraIdentity.Copy(),
		Labels:        in.Labels.Copy(),
	}
	return out
}

func convert_Resource_from(in *compdesc2.Resource) *Resource {
	if in == nil {
		return nil
	}
	acc, err := runtime.ToUnstructuredTypedObject(in.Access)
	if err != nil {
		compdesc2.ThrowConversionError(err)
	}
	out := &Resource{
		ElementMeta: *convert_ElementMeta_from(&in.ElementMeta),
		Type:        in.Type,
		Relation:    in.Relation,
		SourceRef:   convert_SourceRefs_from(in.SourceRef),
		Access:      acc,
	}
	return out
}

func convert_Resources_from(in compdesc2.Resources) Resources {
	if in == nil {
		return nil
	}
	out := make(Resources, len(in))
	for i, v := range in {
		out[i] = *convert_Resource_from(&v)
	}
	return out
}

func convert_SourceRef_from(in *compdesc2.SourceRef) *SourceRef {
	if in == nil {
		return nil
	}
	out := &SourceRef{
		IdentitySelector: in.IdentitySelector.Copy(),
		Labels:           in.Labels.Copy(),
	}
	return out
}

func convert_SourceRefs_from(in []compdesc2.SourceRef) []SourceRef {
	if in == nil {
		return nil
	}
	out := make([]SourceRef, len(in))
	for i, v := range in {
		out[i] = *convert_SourceRef_from(&v)
	}
	return out
}
