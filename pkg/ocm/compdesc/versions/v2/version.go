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
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
	"github.com/gardener/ocm/pkg/ocm/compdesc/versions/v2/jsonscheme"
	"github.com/gardener/ocm/pkg/runtime"
)

const SchemaVersion = "v2"

func init() {
	compdesc.RegisterScheme(&DescriptorVersion{})
}

type DescriptorVersion struct{}

func (v *DescriptorVersion) GetVersion() string {
	return SchemaVersion
}

func (v *DescriptorVersion) Decode(data []byte, opts *compdesc.DecodeOptions) (interface{}, error) {
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

func (v *DescriptorVersion) ConvertTo(obj interface{}) (out *compdesc.ComponentDescriptor, err error) {
	if obj == nil {
		return nil, nil
	}
	in, ok := obj.(*ComponentDescriptor)
	if !ok {
		return nil, errors.Newf("%T is no version v2 descriptor", obj)
	}

	defer compdesc.CatchConversionError(&err)
	out = &compdesc.ComponentDescriptor{
		Metadata: compdesc.Metadata{in.Metadata.Version},
		ComponentSpec: compdesc.ComponentSpec{
			ObjectMeta: compdesc.ObjectMeta{
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

func convert_ComponentReference_to(in *ComponentReference) *compdesc.ComponentReference {
	if in == nil {
		return nil
	}
	out := &compdesc.ComponentReference{
		Name:          in.Name,
		ComponentName: in.ComponentName,
		Version:       in.Version,
		ExtraIdentity: in.ExtraIdentity.Copy(),
		Labels:        in.Labels.Copy(),
	}
	return out
}

func convert_ComponentReferences_to(in []ComponentReference) compdesc.ComponentReferences {
	if in == nil {
		return nil
	}
	out := make(compdesc.ComponentReferences, len(in))
	for i, v := range in {
		out[i] = *convert_ComponentReference_to(&v)
	}
	return out
}

func convert_Source_to(in *Source) *compdesc.Source {
	if in == nil {
		return nil
	}
	out := &compdesc.Source{
		SourceMeta: compdesc.SourceMeta{
			ElementMeta: *convert_ElementMeta_to(&in.ElementMeta),
			Type:        in.Type,
		},
		Access: compdesc.GenericAccessSpec(in.Access.DeepCopy()),
	}
	return out
}

func convert_Sources_to(in Sources) compdesc.Sources {
	if in == nil {
		return nil
	}
	out := make(compdesc.Sources, len(in))
	for i, v := range in {
		out[i] = *convert_Source_to(&v)
	}
	return out
}

func convert_ElementMeta_to(in *ElementMeta) *compdesc.ElementMeta {
	if in == nil {
		return nil
	}
	out := &compdesc.ElementMeta{
		Name:          in.Name,
		Version:       in.Version,
		ExtraIdentity: in.ExtraIdentity.Copy(),
		Labels:        in.Labels.Copy(),
	}
	return out
}

func convert_Resource_to(in *Resource) *compdesc.Resource {
	if in == nil {
		return nil
	}
	out := &compdesc.Resource{
		ResourceMeta: compdesc.ResourceMeta{
			ElementMeta: *convert_ElementMeta_to(&in.ElementMeta),
			Type:        in.Type,
			Relation:    in.Relation,
			SourceRef:   convert_SourceRefs_to(in.SourceRef),
		},
		Access: compdesc.GenericAccessSpec(in.Access),
	}
	return out
}

func convert_Resources_to(in Resources) compdesc.Resources {
	if in == nil {
		return nil
	}
	out := make(compdesc.Resources, len(in))
	for i, v := range in {
		out[i] = *convert_Resource_to(&v)
	}
	return out
}

func convert_SourceRef_to(in *SourceRef) *compdesc.SourceRef {
	if in == nil {
		return nil
	}
	out := &compdesc.SourceRef{
		IdentitySelector: in.IdentitySelector.Copy(),
		Labels:           in.Labels.Copy(),
	}
	return out
}

func convert_SourceRefs_to(in []SourceRef) []compdesc.SourceRef {
	if in == nil {
		return nil
	}
	out := make([]compdesc.SourceRef, len(in))
	for i, v := range in {
		out[i] = *convert_SourceRef_to(&v)
	}
	return out
}

////////////////////////////////////////////////////////////////////////////////
// convert from internal version
////////////////////////////////////////////////////////////////////////////////

func (v *DescriptorVersion) ConvertFrom(in *compdesc.ComponentDescriptor) (interface{}, error) {
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
	return out, nil
}

func convert_ComponentReference_from(in *compdesc.ComponentReference) *ComponentReference {
	if in == nil {
		return nil
	}
	out := &ComponentReference{
		Name:          in.Name,
		ComponentName: in.ComponentName,
		Version:       in.Version,
		ExtraIdentity: in.ExtraIdentity.Copy(),
		Labels:        in.Labels.Copy(),
	}
	return out
}

func convert_ComponentReferences_from(in []compdesc.ComponentReference) []ComponentReference {
	if in == nil {
		return nil
	}
	out := make([]ComponentReference, len(in))
	for i, v := range in {
		out[i] = *convert_ComponentReference_from(&v)
	}
	return out
}

func convert_Source_from(in *compdesc.Source) *Source {
	if in == nil {
		return nil
	}
	acc, err := runtime.ToUnstructuredTypedObject(in.Access)
	if err != nil {
		compdesc.ThrowConversionError(err)
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

func convert_Sources_from(in compdesc.Sources) Sources {
	if in == nil {
		return nil
	}
	out := make(Sources, len(in))
	for i, v := range in {
		out[i] = *convert_Source_from(&v)
	}
	return out
}

func convert_ElementMeta_from(in *compdesc.ElementMeta) *ElementMeta {
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

func convert_Resource_from(in *compdesc.Resource) *Resource {
	if in == nil {
		return nil
	}
	acc, err := runtime.ToUnstructuredTypedObject(in.Access)
	if err != nil {
		compdesc.ThrowConversionError(err)
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

func convert_Resources_from(in compdesc.Resources) Resources {
	if in == nil {
		return nil
	}
	out := make(Resources, len(in))
	for i, v := range in {
		out[i] = *convert_Resource_from(&v)
	}
	return out
}

func convert_SourceRef_from(in *compdesc.SourceRef) *SourceRef {
	if in == nil {
		return nil
	}
	out := &SourceRef{
		IdentitySelector: in.IdentitySelector.Copy(),
		Labels:           in.Labels.Copy(),
	}
	return out
}

func convert_SourceRefs_from(in []compdesc.SourceRef) []SourceRef {
	if in == nil {
		return nil
	}
	out := make([]SourceRef, len(in))
	for i, v := range in {
		out[i] = *convert_SourceRef_from(&v)
	}
	return out
}
