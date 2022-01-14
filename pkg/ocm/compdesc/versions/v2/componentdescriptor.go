// Copyright 2022 Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
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
	"errors"

	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
	"github.com/gardener/ocm/pkg/ocm/runtime"
)

var (
	NotFound = errors.New("NotFound")
)

// ComponentDescriptor defines a versioned component with a source and dependencies.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type ComponentDescriptor struct {
	// Metadata specifies the schema version of the component.
	Metadata metav1.Metadata `json:"meta"`
	// Spec contains the specification of the component.
	ComponentSpec `json:"component"`
}

// ComponentSpec defines a virtual component with
// a repository context, source and dependencies.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type ComponentSpec struct {
	ObjectMeta `json:",inline"`
	// RepositoryContexts defines the previous repositories of the component
	RepositoryContexts runtime.UnstructuredTypedObjectList `json:"repositoryContexts"`
	// Provider defines the provider type of a component.
	// It can be external or internal.
	Provider metav1.ProviderType `json:"provider"`
	// Sources defines sources that produced the component
	Sources Sources `json:"sources"`
	// ComponentReferences references component dependencies that can be resolved in the current context.
	ComponentReferences []ComponentReference `json:"componentReferences"`
	// Resources defines all resources that are created by the component and by a third party.
	Resources Resources `json:"resources"`
}

// ObjectMeta defines a object that is uniquely identified by its name and version.
// +k8s:deepcopy-gen=true
type ObjectMeta struct {
	// Name is the context unique name of the object.
	Name string `json:"name"`
	// Version is the semver version of the object.
	Version string `json:"version"`
	// Labels defines an optional set of additional labels
	// describing the object.
	// +optional
	Labels metav1.Labels `json:"labels,omitempty"`
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
func (o ObjectMeta) GetLabels() metav1.Labels {
	return o.Labels
}

// SetLabels sets the labels of the object.
func (o *ObjectMeta) SetLabels(labels []metav1.Label) {
	o.Labels = labels
}

const (
	SystemIdentityName    = "name"
	SystemIdentityVersion = "version"
)

// ElementMeta defines a object that is uniquely identified by its identity.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type ElementMeta struct {
	// Name is the context unique name of the object.
	Name string `json:"name"`
	// Version is the semver version of the object.
	Version string `json:"version"`
	// ExtraIdentity is the identity of an object.
	// An additional label with key "name" ist not allowed
	ExtraIdentity metav1.Identity `json:"extraIdentity,omitempty"`
	// Labels defines an optional set of additional labels
	// describing the object.
	// +optional
	Labels metav1.Labels `json:"labels,omitempty"`
}

// GetName returns the name of the object.
func (o ElementMeta) GetName() string {
	return o.Name
}

// SetName sets the name of the object.
func (o *ElementMeta) SetName(name string) {
	o.Name = name
}

// GetVersion returns the version of the object.
func (o ElementMeta) GetVersion() string {
	return o.Version
}

// SetVersion sets the version of the object.
func (o *ElementMeta) SetVersion(version string) {
	o.Version = version
}

// GetLabels returns the label of the object.
func (o ElementMeta) GetLabels() metav1.Labels {
	return o.Labels
}

// SetLabels sets the labels of the object.
func (o *ElementMeta) SetLabels(labels []metav1.Label) {
	o.Labels = labels
}

// SetExtraIdentity sets the identity of the object.
func (o *ElementMeta) SetExtraIdentity(identity metav1.Identity) {
	o.ExtraIdentity = identity
}

// GetIdentity returns the identity of the object.
func (o *ElementMeta) GetIdentity(accessor ElementMetaAccessor) metav1.Identity {
	identity := map[string]string{}
	for k, v := range o.ExtraIdentity {
		identity[k] = v
	}
	identity[SystemIdentityName] = o.Name
	if accessor != nil {
		found := false
		for _, m := range accessor.GetMetas() {
			if m.Name == o.Name {
				if found {
					identity[SystemIdentityVersion] = o.Version
					break
				}
				found = true
			}
		}
	}
	return identity
}

// GetIdentityDigest returns the digest of the object's identity.
func (o *ElementMeta) GetIdentityDigest(accessor ElementMetaAccessor) []byte {
	return o.GetIdentity(accessor).Digest()
}

// ElementMetaAccessor describes an accessor for the set of meta information for a set of elements.
type ElementMetaAccessor interface {
	// Get all the meta information for a set of elements
	GetMetas() []ElementMeta
}

// Sources describes a set of source specifications
type Sources []Source

func (s Sources) GetMetas() []ElementMeta {
	metas := make([]ElementMeta, len(s), len(s))
	for i, src := range s {
		metas[i] = src.ElementMeta
	}
	return metas
}

// Source is the definition of a component's source.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type Source struct {
	SourceMeta `json:",inline"`

	Access *runtime.UnstructuredTypedObject `json:"access"`
}

// SourceMeta is the definition of the meta data of a source.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type SourceMeta struct {
	ElementMeta
	// Type describes the type of the object.
	Type string `json:"type"`
}

// GetType returns the type of the object.
func (o SourceMeta) GetType() string {
	return o.Type
}

// SetType sets the type of the object.
func (o *SourceMeta) SetType(ttype string) {
	o.Type = ttype
}

// SourceRef defines a reference to a source
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type SourceRef struct {
	// IdentitySelector defines the identity that is used to match a source.
	IdentitySelector metav1.StringMap `json:"identitySelector,omitempty"`
	// Labels defines an optional set of additional labels
	// describing the object.
	// +optional
	Labels metav1.Labels `json:"labels,omitempty"`
}

// Resources describes a set of resource specifications
type Resources []Resource

func (r Resources) GetMetas() []ElementMeta {
	metas := make([]ElementMeta, len(r), len(r))
	for i, rsc := range r {
		metas[i] = rsc.ElementMeta
	}
	return metas
}

// Resource describes a resource dependency of a component.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type Resource struct {
	ElementMeta `json:",inline"`

	// Type describes the type of the object.
	Type string `json:"type"`

	// Relation describes the relation of the resource to the component.
	// Can be a local or external resource
	Relation metav1.ResourceRelation `json:"relation,omitempty"`

	// SourceRef defines a list of source names.
	// These names reference the sources defines in `component.sources`.
	SourceRef []SourceRef `json:"srcRef,omitempty"`

	// Access describes the type specific method to
	// access the defined resource.
	Access *runtime.UnstructuredTypedObject `json:"access"`
}

// GetType returns the type of the object.
func (o Resource) GetType() string {
	return o.Type
}

// SetType sets the type of the object.
func (o *Resource) SetType(ttype string) {
	o.Type = ttype
}

// ComponentReference describes the reference to another component in the registry.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type ComponentReference struct {
	// Name is the context unique name of the object.
	Name string `json:"name"`
	// ComponentName describes the remote name of the referenced object
	ComponentName string `json:"componentName"`
	// Version is the semver version of the object.
	Version string `json:"version"`
	// ExtraIdentity is the identity of an object.
	// An additional label with key "name" ist not allowed
	ExtraIdentity metav1.Identity `json:"extraIdentity,omitempty"`
	// Labels defines an optional set of additional labels
	// describing the object.
	// +optional
	Labels metav1.Labels `json:"labels,omitempty"`
}

// GetName returns the name of the object.
func (o ComponentReference) GetName() string {
	return o.Name
}

// SetName sets the name of the object.
func (o *ComponentReference) SetName(name string) {
	o.Name = name
}

// GetVersion returns the version of the object.
func (o ComponentReference) GetVersion() string {
	return o.Version
}

// SetVersion sets the version of the object.
func (o *ComponentReference) SetVersion(version string) {
	o.Version = version
}

// GetLabels returns the label of the object.
func (o ComponentReference) GetLabels() metav1.Labels {
	return o.Labels
}

// SetLabels sets the labels of the object.
func (o *ComponentReference) SetLabels(labels []metav1.Label) {
	o.Labels = labels
}

// GetIdentity returns the identity of the object.
func (o *ComponentReference) GetIdentity() metav1.Identity {
	identity := map[string]string{}
	for k, v := range o.ExtraIdentity {
		identity[k] = v
	}
	identity[SystemIdentityName] = o.Name
	return identity
}

// GetIdentityDigest returns the digest of the object's identity.
func (o *ComponentReference) GetIdentityDigest() []byte {
	return o.GetIdentity().Digest()
}
