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

package v1

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	SystemIdentityName    = "name"
	SystemIdentityVersion = "version"
)

// Metadata defines the metadata of the component descriptor.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type Metadata struct {
	// Version is the schema version of the component descriptor.
	Version string `json:"schemaVersion"`
}

// ProviderName describes the provider type of component in the origin's context.
// Defines whether the component is created by a third party or internally.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type ProviderName string

// ResourceRelation describes the type of a resource.
// Defines whether the component is created by a third party or internally.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type ResourceRelation string

const (
	// LocalRelation defines a internal relation
	// which describes a internally maintained resource in the origin's context.
	LocalRelation ResourceRelation = "local"
	// ExternalRelation defines a external relation
	// which describes a resource maintained by a third party vendor in the origin's context.
	ExternalRelation ResourceRelation = "external"
)

func ValidateRelation(fldPath *field.Path, relation ResourceRelation) *field.Error {
	if len(relation) == 0 {
		return field.Required(fldPath, "relation must be set")
	}
	if relation != LocalRelation && relation != ExternalRelation {
		return field.NotSupported(fldPath, relation, []string{string(LocalRelation), string(ExternalRelation)})
	}
	return nil
}

const (
	GROUP = "ocm.gardener.cloud"
	KIND  = "ComponentVersion"
)

// TypeMeta describes the schema of a descriptor.
type TypeMeta struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

// ObjectMeta defines the metadata of the component descriptor.
type ObjectMeta struct {
	// Name is the name of the component.
	Name string `json:"name"`
	// Version is the version of the component.
	Version string `json:"version"`
	// Labels describe additional properties of the component version
	Labels Labels `json:"labels,omitempty"`
	// Provider described the component provider
	Provider Provider `json:"provider"`
}

// GetName returns the name of the object.
func (o ObjectMeta) GetName() string {
	return o.Name
}

// SetName sets the name of the object.
func (o *ObjectMeta) SetName(name string) {
	o.Name = name
}

// GetVersion returns the version of the object.
func (o ObjectMeta) GetVersion() string {
	return o.Version
}

// SetVersion sets the version of the object.
func (o *ObjectMeta) SetVersion(version string) {
	o.Version = version
}

// GetLabels returns the label of the object.
func (o ObjectMeta) GetLabels() Labels {
	return o.Labels
}

// SetLabels sets the labels of the object.
func (o *ObjectMeta) SetLabels(labels []Label) {
	o.Labels = labels
}

// GetName returns the name of the object.
func (o *ObjectMeta) Copy() *ObjectMeta {
	return &ObjectMeta{
		Name:     o.Name,
		Version:  o.Version,
		Labels:   o.Labels.Copy(),
		Provider: *o.Provider.Copy(),
	}
}

////////////////////////////////////////////////////////////////////////////////

// Provider describes the provider information of a component version.
type Provider struct {
	Name ProviderName `json:"name"`
	// Labels describe additional properties of provider
	Labels Labels `json:"labels,omitempty"`
}

// GetName returns the name of the provider.
func (o Provider) GetName() ProviderName {
	return o.Name
}

// SetName sets the name of the provider.
func (o *Provider) SetName(name ProviderName) {
	o.Name = name
}

// GetLabels returns the label of the provider.
func (o Provider) GetLabels() Labels {
	return o.Labels
}

// SetLabels sets the labels of the provider.
func (o *Provider) SetLabels(labels []Label) {
	o.Labels = labels
}

// Copy copies the provider info.
func (o *Provider) Copy() *Provider {
	return &Provider{
		Name:   o.Name,
		Labels: o.Labels.Copy(),
	}
}
