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

package compdesc

import (
	"encoding/json"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/v2/jsonscheme"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const SchemaVersion = "v2"

func init() {
	compdesc.RegisterScheme(&DescriptorVersion{})
}

type DescriptorVersion struct{}

var _ compdesc.Scheme = (*DescriptorVersion)(nil)

func (v *DescriptorVersion) GetVersion() string {
	return SchemaVersion
}

func (v *DescriptorVersion) Decode(data []byte, opts *compdesc.DecodeOptions) (compdesc.ComponentDescriptorVersion, error) {
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

	if err := cd.Default(); err != nil {
		return nil, err
	}

	if !opts.DisableValidation {
		err = cd.Validate()
		if err != nil {
			return nil, err
		}
	}
	return &cd, err
}

////////////////////////////////////////////////////////////////////////////////
// convert to internal version
////////////////////////////////////////////////////////////////////////////////

func (v *DescriptorVersion) ConvertTo(obj compdesc.ComponentDescriptorVersion) (out *compdesc.ComponentDescriptor, err error) {
	if obj == nil {
		return nil, nil
	}
	in, ok := obj.(*ComponentDescriptor)
	if !ok {
		return nil, errors.Newf("%T is no version v2 descriptor", obj)
	}

	defer compdesc.CatchConversionError(&err)
	var provider metav1.Provider
	err = json.Unmarshal([]byte(in.Provider), &provider)
	if err != nil {
		provider.Name = in.Provider
		provider.Labels = nil
	}

	out = &compdesc.ComponentDescriptor{
		Metadata: compdesc.Metadata{in.Metadata.Version},
		ComponentSpec: compdesc.ComponentSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:     in.Name,
				Version:  in.Version,
				Labels:   in.Labels.Copy(),
				Provider: provider,
			},
			RepositoryContexts: in.RepositoryContexts.Copy(),
			Sources:            convert_Sources_to(in.Sources),
			Resources:          convert_Resources_to(in.Resources),
			References:         convert_ComponentReferences_to(in.ComponentReferences),
		},
		Signatures: in.Signatures.Copy(),
	}
	return out, nil
}

func convert_ComponentReference_to(in *ComponentReference) *compdesc.ComponentReference {
	if in == nil {
		return nil
	}
	out := &compdesc.ComponentReference{
		ElementMeta:   *convert_ElementMeta_to(&in.ElementMeta),
		ComponentName: in.ComponentName,
		Digest:        in.Digest.Copy(),
	}
	return out
}

func convert_ComponentReferences_to(in []ComponentReference) compdesc.References {
	if in == nil {
		return nil
	}
	out := make(compdesc.References, len(in))
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
			SourceRef:   Convert_SourceRefs_to(in.SourceRef),
			Digest:      in.Digest.Copy(),
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

func Convert_SourceRefs_to(in []SourceRef) []compdesc.SourceRef {
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

func (v *DescriptorVersion) ConvertFrom(in *compdesc.ComponentDescriptor) (compdesc.ComponentDescriptorVersion, error) {
	if in == nil {
		return nil, nil
	}
	provider := in.Provider.Name
	if len(in.Provider.Labels) != 0 {
		data, err := json.Marshal(in.Provider)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot marshal provider")
		}
		provider = metav1.ProviderName(data)
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
			Provider:            provider,
			Sources:             convert_Sources_from(in.Sources),
			Resources:           convert_Resources_from(in.Resources),
			ComponentReferences: convert_ComponentReferences_from(in.References),
		},
		Signatures: in.Signatures.Copy(),
	}
	if err := out.Default(); err != nil {
		return nil, err
	}
	return out, nil
}

func convert_ComponentReference_from(in *compdesc.ComponentReference) *ComponentReference {
	if in == nil {
		return nil
	}
	out := &ComponentReference{
		ElementMeta:   *convert_ElementMeta_from(&in.ElementMeta),
		ComponentName: in.ComponentName,
		Digest:        in.Digest.Copy(),
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
		Digest:      in.Digest.Copy(),
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
