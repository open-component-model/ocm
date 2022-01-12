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
	"encoding/json"
	"errors"

	"github.com/gardener/ocm/pkg/ocm/runtime"
)

const SchemaVersion = "v2"

var (
	NotFound = errors.New("NotFound")
)

// Metadata defines the metadata of the component descriptor.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type Metadata struct {
	// Version is the schema version of the component descriptor.
	Version string `json:"schemaVersion"`
}

// ProviderType describes the provider type of component in the origin's context.
// Defines whether the component is created by a third party or internally.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type ProviderType string

// ResourceRelation describes the type of a resource.
// Defines whether the component is created by a third party or internally.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type ResourceRelation string

// ComponentDescriptor defines a versioned component with a source and dependencies.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type ComponentDescriptor struct {
	// Metadata specifies the schema version of the component.
	Metadata Metadata `json:"meta"`
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
	RepositoryContexts []*runtime.UnstructuredTypedObject `json:"repositoryContexts"`
	// Provider defines the provider type of a component.
	// It can be external or internal.
	Provider ProviderType `json:"provider"`
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
	Labels Labels `json:"labels,omitempty"`
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

const (
	SystemIdentityName    = "name"
	SystemIdentityVersion = "version"
)

// Identity describes the identity of an object.
// Only ascii characters are allowed
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type Identity map[string]string

// Digest returns the object digest of an identity
func (i Identity) Digest() []byte {
	data, _ := json.Marshal(i)
	return data
}

// Match implements the selector interface.
func (i Identity) Match(obj map[string]string) (bool, error) {
	for k, v := range i {
		if obj[k] != v {
			return false, nil
		}
	}
	return true, nil
}

// ElementMeta defines a object that is uniquely identified by its identity.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type ElementMeta struct {
	// Name is the context unique name of the object.
	Name string `json:"name"`
	// Version is the semver version of the object.
	Version string `json:"version"`
	// Type describes the type of the object.
	Type string `json:"type"`
	// ExtraIdentity is the identity of an object.
	// An additional label with key "name" ist not allowed
	ExtraIdentity Identity `json:"extraIdentity,omitempty"`
	// Labels defines an optional set of additional labels
	// describing the object.
	// +optional
	Labels Labels `json:"labels,omitempty"`
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

// GetType returns the type of the object.
func (o ElementMeta) GetType() string {
	return o.Type
}

// SetType sets the type of the object.
func (o *ElementMeta) SetType(ttype string) {
	o.Type = ttype
}

// GetLabels returns the label of the object.
func (o ElementMeta) GetLabels() Labels {
	return o.Labels
}

// SetLabels sets the labels of the object.
func (o *ElementMeta) SetLabels(labels []Label) {
	o.Labels = labels
}

// SetExtraIdentity sets the identity of the object.
func (o *ElementMeta) SetExtraIdentity(identity Identity) {
	o.ExtraIdentity = identity
}

// GetIdentity returns the identity of the object.
func (o *ElementMeta) GetIdentity(accessor ElementMetaAccessor) Identity {
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

// Label is a label that can be set on objects.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type Label struct {
	// Name is the unique name of the label.
	Name string `json:"name"`
	// Value is the json/yaml data of the label
	Value json.RawMessage `json:"value"`
}

// Labels describe a list of labels
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type Labels []Label

// Get returns the label witht the given name
func (l Labels) Get(name string) ([]byte, bool) {
	for _, label := range l {
		if label.Name == name {
			return label.Value, true
		}
	}
	return nil, false
}

// NameAccessor describes a accessor for a named object.
type NameAccessor interface {
	// GetName returns the name of the object.
	GetName() string
	// SetName sets the name of the object.
	SetName(name string)
}

// VersionAccessor describes a accessor for a versioned object.
type VersionAccessor interface {
	// GetVersion returns the version of the object.
	GetVersion() string
	// SetVersion sets the version of the object.
	SetVersion(version string)
}

// LabelsAccessor describes a accessor for a labeled object.
type LabelsAccessor interface {
	// GetLabels returns the labels of the object.
	GetLabels() Labels
	// SetLabels sets the labels of the object.
	SetLabels(labels []Label)
}

// ObjectMetaAccessor describes a accessor for named and versioned object.
type ObjectMetaAccessor interface {
	NameAccessor
	VersionAccessor
	LabelsAccessor
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
	ElementMeta `json:",inline"`
	Access      *runtime.UnstructuredTypedObject `json:"access"`
}

// SourceRef defines a reference to a source
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type SourceRef struct {
	// IdentitySelector defines the identity that is used to match a source.
	IdentitySelector map[string]string `json:"identitySelector,omitempty"`
	// Labels defines an optional set of additional labels
	// describing the object.
	// +optional
	Labels Labels `json:"labels,omitempty"`
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

	// Relation describes the relation of the resource to the component.
	// Can be a local or external resource
	Relation ResourceRelation `json:"relation,omitempty"`

	// SourceRef defines a list of source names.
	// These names reference the sources defines in `component.sources`.
	SourceRef []SourceRef `json:"srcRef,omitempty"`

	// Access describes the type specific method to
	// access the defined resource.
	Access *runtime.UnstructuredTypedObject `json:"access"`
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
	ExtraIdentity Identity `json:"extraIdentity,omitempty"`
	// Labels defines an optional set of additional labels
	// describing the object.
	// +optional
	Labels Labels `json:"labels,omitempty"`
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
func (o ComponentReference) GetLabels() Labels {
	return o.Labels
}

// SetLabels sets the labels of the object.
func (o *ComponentReference) SetLabels(labels []Label) {
	o.Labels = labels
}

// GetIdentity returns the identity of the object.
func (o *ComponentReference) GetIdentity() Identity {
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
