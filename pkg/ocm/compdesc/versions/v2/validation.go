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
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// Validate validates a parsed v2 component descriptor
func Validate(component *ComponentDescriptor) error {
	if err := validate(nil, component); err != nil {
		return err.ToAggregate()
	}
	return nil
}

func validate(fldPath *field.Path, component *ComponentDescriptor) field.ErrorList {
	if component == nil {
		return nil
	}
	allErrs := field.ErrorList{}

	if len(component.Metadata.Version) == 0 {
		metaPath := field.NewPath("meta").Child("schemaVersion")
		if fldPath != nil {
			metaPath = fldPath.Child("meta").Child("schemaVersion")
		}
		allErrs = append(allErrs, field.Required(metaPath, "must specify a version"))
	}

	compPath := field.NewPath("component")
	if fldPath != nil {
		compPath = fldPath.Child("component")
	}

	if err := validateProvider(compPath.Child("provider"), component.Provider); err != nil {
		allErrs = append(allErrs, err)
	}

	allErrs = append(allErrs, ValidateObjectMeta(compPath, component)...)

	srcPath := compPath.Child("sources")
	allErrs = append(allErrs, ValidateSources(srcPath, component.Sources)...)

	refPath := compPath.Child("componentReferences")
	allErrs = append(allErrs, ValidateComponentReferences(refPath, component.ComponentReferences)...)

	resourcePath := compPath.Child("resources")
	allErrs = append(allErrs, ValidateResources(resourcePath, component.Resources, component.GetVersion())...)

	return allErrs
}

// ValidateObjectMeta validate the metadata of an object.
func ValidateObjectMeta(fldPath *field.Path, om compdesc.ObjectMetaAccessor) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(om.GetName()) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("name"), "must specify a name"))
	}
	if len(om.GetVersion()) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("version"), "must specify a version"))
	}
	if len(om.GetLabels()) != 0 {
		allErrs = append(allErrs, metav1.ValidateLabels(fldPath.Child("labels"), om.GetLabels())...)
	}
	return allErrs
}

// ValidateSources validates a list of sources.
// It makes sure that no duplicate sources are present.
func ValidateSources(fldPath *field.Path, sources Sources) field.ErrorList {
	allErrs := field.ErrorList{}
	sourceIDs := make(map[string]struct{})
	for i, src := range sources {
		srcPath := fldPath.Index(i)
		allErrs = append(allErrs, ValidateSource(srcPath, src)...)

		id := string(src.GetIdentityDigest(sources))
		if _, ok := sourceIDs[id]; ok {
			allErrs = append(allErrs, field.Duplicate(srcPath, "duplicate source"))
			continue
		}
		sourceIDs[id] = struct{}{}
	}
	return allErrs
}

// ValidateSource validates the a component's source object.
func ValidateSource(fldPath *field.Path, src Source) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(src.GetName()) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("name"), "must specify a name"))
	}
	if len(src.GetType()) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("type"), "must specify a type"))
	}
	allErrs = append(allErrs, metav1.ValidateIdentity(fldPath.Child("extraIdentity"), src.ExtraIdentity)...)
	return allErrs
}

// ValidateResource validates a components resource
func ValidateResource(fldPath *field.Path, res Resource, access bool) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, ValidateObjectMeta(fldPath, &res)...)

	if err := metav1.ValidateRelation(fldPath.Child("relation"), res.Relation); err != nil {
		allErrs = append(allErrs, err)
	}

	if !metav1.IsIdentity(res.Name) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), res.Name, metav1.IdentityKeyValidationErrMsg))
	}

	if len(res.GetType()) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("type"), "must specify a type"))
	}

	if res.Access == nil && access {
		allErrs = append(allErrs, field.Required(fldPath.Child("access"), "must specify a access"))
	}
	allErrs = append(allErrs, metav1.ValidateIdentity(fldPath.Child("extraIdentity"), res.ExtraIdentity)...)

	return allErrs
}

func validateProvider(fldPath *field.Path, provider metav1.ProviderType) *field.Error {
	if len(provider) == 0 {
		return field.Required(fldPath, "provider must be set")
	}
	return nil
}

// ValidateComponentReference validates a component reference.
func ValidateComponentReference(fldPath *field.Path, cr ComponentReference) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(cr.ComponentName) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("componentName"), "must specify a component name"))
	}
	allErrs = append(allErrs, ValidateObjectMeta(fldPath, &cr)...)
	return allErrs
}

// ValidateComponentReferences validates a list of component references.
// It makes sure that no duplicate sources are present.
func ValidateComponentReferences(fldPath *field.Path, refs []ComponentReference) field.ErrorList {
	allErrs := field.ErrorList{}
	refIDs := make(map[string]struct{})
	for i, ref := range refs {
		refPath := fldPath.Index(i)
		allErrs = append(allErrs, ValidateComponentReference(refPath, ref)...)

		id := string(ref.GetIdentityDigest())
		if _, ok := refIDs[id]; ok {
			allErrs = append(allErrs, field.Duplicate(refPath, "duplicate component reference name"))
			continue
		}
		refIDs[id] = struct{}{}
	}
	return allErrs
}

// ValidateResources validates a list of resources.
// It makes sure that no duplicate sources are present.
func ValidateResources(fldPath *field.Path, resources Resources, componentVersion string) field.ErrorList {
	allErrs := field.ErrorList{}
	resourceIDs := make(map[string]struct{})
	for i, res := range resources {
		localPath := fldPath.Index(i)
		allErrs = append(allErrs, ValidateResource(localPath, res, true)...)

		// only validate the component version if it is defined
		if res.Relation == metav1.LocalRelation && len(componentVersion) != 0 {
			if res.GetVersion() != componentVersion {
				allErrs = append(allErrs, field.Invalid(localPath.Child("version"), "invalid version",
					"version of local resources must match the component version"))
			}
		}

		if err := ValidateSourceRefs(localPath.Child("sourceRef"), res.SourceRef); err != nil {
			allErrs = append(allErrs, err...)
		}

		id := string(res.GetIdentityDigest(resources))
		if _, ok := resourceIDs[id]; ok {
			allErrs = append(allErrs, field.Duplicate(localPath, "duplicated resource"))
			continue
		}
		resourceIDs[id] = struct{}{}
	}
	return allErrs
}

func ValidateSourceRefs(fldPath *field.Path, srcs []SourceRef) field.ErrorList {
	allErrs := field.ErrorList{}
	for i, src := range srcs {
		localPath := fldPath.Index(i)
		if err := metav1.ValidateLabels(localPath.Child("labels"), src.Labels); err != nil {
			allErrs = append(allErrs, err...)
		}
	}
	return allErrs
}
