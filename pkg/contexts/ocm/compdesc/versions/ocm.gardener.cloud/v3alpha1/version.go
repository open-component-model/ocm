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

package v3alpha1

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/versions/ocm.gardener.cloud/v3alpha1/jsonscheme"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const SchemaVersion = "v3alpha1"
const GroupVersion = metav1.GROUP + "/" + SchemaVersion
const Kind = metav1.KIND

func init() {
	compdesc.RegisterScheme(&DescriptorVersion{})
}

type DescriptorVersion struct{}

var _ compdesc.Scheme = (*DescriptorVersion)(nil)

func (v *DescriptorVersion) GetVersion() string {
	return GroupVersion
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
	if in.Kind != Kind {
		return nil, errors.ErrInvalid("kind", in.Kind)
	}

	defer compdesc.CatchConversionError(&err)
	out = &compdesc.ComponentDescriptor{
		Metadata: compdesc.Metadata{
			ConfiguredVersion: in.APIVersion,
		},
		ComponentSpec: compdesc.ComponentSpec{
			ObjectMeta:         *in.ObjectMeta.Copy(),
			RepositoryContexts: in.RepositoryContexts.Copy(),
			Sources:            convertSourcesTo(in.Spec.Sources),
			Resources:          convertResourcesTo(in.Spec.Resources),
			References:         convertReferencesTo(in.Spec.References),
		},
		Signatures: in.Signatures.Copy(),
	}
	return out, nil
}

func convertReferenceTo(in *Reference) *compdesc.ComponentReference {
	if in == nil {
		return nil
	}
	out := &compdesc.ComponentReference{
		ElementMeta:   *convertElementMetaTo(&in.ElementMeta),
		ComponentName: in.ComponentName,
		Digest:        in.Digest.Copy(),
	}
	return out
}

func convertReferencesTo(in []Reference) compdesc.References {
	if in == nil {
		return nil
	}
	out := make(compdesc.References, len(in))
	for i, v := range in {
		out[i] = *convertReferenceTo(&v)
	}
	return out
}

func convertSourceTo(in *Source) *compdesc.Source {
	if in == nil {
		return nil
	}
	out := &compdesc.Source{
		SourceMeta: compdesc.SourceMeta{
			ElementMeta: *convertElementMetaTo(&in.ElementMeta),
			Type:        in.Type,
		},
		Access: compdesc.GenericAccessSpec(in.Access.DeepCopy()),
	}
	return out
}

func convertSourcesTo(in Sources) compdesc.Sources {
	if in == nil {
		return nil
	}
	out := make(compdesc.Sources, len(in))
	for i, v := range in {
		out[i] = *convertSourceTo(&v)
	}
	return out
}

func convertElementMetaTo(in *ElementMeta) *compdesc.ElementMeta {
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

func convertResourceTo(in *Resource) *compdesc.Resource {
	if in == nil {
		return nil
	}
	out := &compdesc.Resource{
		ResourceMeta: compdesc.ResourceMeta{
			ElementMeta: *convertElementMetaTo(&in.ElementMeta),
			Type:        in.Type,
			Relation:    in.Relation,
			SourceRef:   ConvertSourcerefsTo(in.SourceRef),
			Digest:      in.Digest.Copy(),
		},
		Access: compdesc.GenericAccessSpec(in.Access),
	}
	return out
}

func convertResourcesTo(in Resources) compdesc.Resources {
	if in == nil {
		return nil
	}
	out := make(compdesc.Resources, len(in))
	for i, v := range in {
		out[i] = *convertResourceTo(&v)
	}
	return out
}

func convertSourcerefTo(in *SourceRef) *compdesc.SourceRef {
	if in == nil {
		return nil
	}
	out := &compdesc.SourceRef{
		IdentitySelector: in.IdentitySelector.Copy(),
		Labels:           in.Labels.Copy(),
	}
	return out
}

func ConvertSourcerefsTo(in []SourceRef) []compdesc.SourceRef {
	if in == nil {
		return nil
	}
	out := make([]compdesc.SourceRef, len(in))
	for i, v := range in {
		out[i] = *convertSourcerefTo(&v)
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
	sourceFrom, err := convertSourcesFrom(in.Sources)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sources from: %w", err)
	}
	resourcesFrom, err := convertResourcesFrom(in.Resources)
	if err != nil {
		return nil, fmt.Errorf("failed to convert resources from: %w", err)
	}
	out := &ComponentDescriptor{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion,
			Kind:       Kind,
		},
		ObjectMeta:         *in.ObjectMeta.Copy(),
		RepositoryContexts: in.RepositoryContexts.Copy(),
		Spec: ComponentVersionSpec{
			Sources:    sourceFrom,
			Resources:  resourcesFrom,
			References: convertReferencesFrom(in.References),
		},
		Signatures: in.Signatures.Copy(),
	}
	if err := out.Default(); err != nil {
		return nil, err
	}
	return out, nil
}

func convertReferenceFrom(in *compdesc.ComponentReference) *Reference {
	if in == nil {
		return nil
	}
	out := &Reference{
		ElementMeta:   *convertElementMetaFrom(&in.ElementMeta),
		ComponentName: in.ComponentName,
		Digest:        in.Digest.Copy(),
	}
	return out
}

func convertReferencesFrom(in []compdesc.ComponentReference) []Reference {
	if in == nil {
		return nil
	}
	out := make([]Reference, len(in))
	for i, v := range in {
		out[i] = *convertReferenceFrom(&v)
	}
	return out
}

func convertSourceFrom(in *compdesc.Source) (*Source, error) {
	if in == nil {
		return nil, nil
	}
	acc, err := runtime.ToUnstructuredTypedObject(in.Access)
	if err != nil {
		return nil, fmt.Errorf("failed to convert source: %w", err)
	}
	out := &Source{
		SourceMeta: SourceMeta{
			ElementMeta: *convertElementMetaFrom(&in.ElementMeta),
			Type:        in.Type,
		},
		Access: acc,
	}
	return out, nil
}

func convertSourcesFrom(in compdesc.Sources) (Sources, error) {
	if in == nil {
		return nil, nil
	}
	out := make(Sources, len(in))
	for i, v := range in {
		from, err := convertSourceFrom(&v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert sources form: %w", err)
		}
		out[i] = *from
	}
	return out, nil
}

func convertElementMetaFrom(in *compdesc.ElementMeta) *ElementMeta {
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

func convertResourceFrom(in *compdesc.Resource) (*Resource, error) {
	if in == nil {
		return nil, nil
	}
	acc, err := runtime.ToUnstructuredTypedObject(in.Access)
	if err != nil {
		return nil, fmt.Errorf("failed to convert resource form: %w", err)
	}
	out := &Resource{
		ElementMeta: *convertElementMetaFrom(&in.ElementMeta),
		Type:        in.Type,
		Relation:    in.Relation,
		SourceRef:   convertSourceRefsFrom(in.SourceRef),
		Access:      acc,
		Digest:      in.Digest.Copy(),
	}
	return out, nil
}

func convertResourcesFrom(in compdesc.Resources) (Resources, error) {
	if in == nil {
		return nil, nil
	}
	out := make(Resources, len(in))
	for i, v := range in {
		from, err := convertResourceFrom(&v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert resource from: %w", err)
		}
		out[i] = *from
	}
	return out, nil
}

func convertSourceRefFrom(in *compdesc.SourceRef) *SourceRef {
	if in == nil {
		return nil
	}
	out := &SourceRef{
		IdentitySelector: in.IdentitySelector.Copy(),
		Labels:           in.Labels.Copy(),
	}
	return out
}

func convertSourceRefsFrom(in []compdesc.SourceRef) []SourceRef {
	if in == nil {
		return nil
	}
	out := make([]SourceRef, len(in))
	for i, v := range in {
		out[i] = *convertSourceRefFrom(&v)
	}
	return out
}
