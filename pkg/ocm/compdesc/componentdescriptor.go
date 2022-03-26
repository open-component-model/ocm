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
	"github.com/gardener/ocm/pkg/errors"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
	"github.com/gardener/ocm/pkg/runtime"
)

var (
	NotFound = errors.ErrNotFound()
)

// Metadata defines the configured metadata of the component descriptor.
// It is taken from the original serialization format. It can be set
// to define a default serialization version.
type Metadata struct {
	ConfiguredVersion string `json:"configuredSchemaVersion"`
}

// ComponentDescriptor defines a versioned component with a source and dependencies.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type ComponentDescriptor struct {
	// Metadata specifies the schema version of the component.
	Metadata Metadata `json:"meta"`
	// Spec contains the specification of the component.
	ComponentSpec `json:"component"`
}

func New(name, version string) *ComponentDescriptor {
	return DefaultComponent(&ComponentDescriptor{
		Metadata: Metadata{
			ConfiguredVersion: "v2",
		},
		ComponentSpec: ComponentSpec{
			ObjectMeta: ObjectMeta{
				Name:    name,
				Version: version,
			},
			Provider: "acme",
		},
	})
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
	ComponentReferences ComponentReferences `json:"componentReferences"`
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

// Copy copies the ObjectMeta value
func (o ObjectMeta) Copy() ObjectMeta {
	o.Labels = o.Labels.Copy()
	return o
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

// SetLabel sets a single label to an effective value.
// If the value is no byte slice, it is marshaled.
func (o *ElementMeta) SetLabel(name string, value interface{}) error {
	return o.Labels.Set(name, value)
}

// RemoveLabel removes a single label.
func (o *ElementMeta) RemoveLabel(name string) bool {
	return o.Labels.Remove(name)
}

// SetExtraIdentity sets the identity of the object.
func (o *ElementMeta) SetExtraIdentity(identity metav1.Identity) {
	o.ExtraIdentity = identity
}

// GetIdentity returns the identity of the object.
func (o *ElementMeta) GetIdentity(accessor ElementAccessor) metav1.Identity {
	identity := o.ExtraIdentity.Copy()
	if identity == nil {
		identity = metav1.Identity{}
	}
	identity[SystemIdentityName] = o.Name
	if accessor != nil {
		found := false
		l := accessor.Len()
		for i := 0; i < l; i++ {
			m := accessor.Get(i).GetMeta()
			if m.Name == o.Name && m.ExtraIdentity.Equals(o.ExtraIdentity) {
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

// GetMatchBaseIdentity returns all possible identity attributes for resource matching
func (o *ElementMeta) GetMatchBaseIdentity() metav1.Identity {
	identity := o.ExtraIdentity.Copy()
	if identity == nil {
		identity = metav1.Identity{}
	}
	identity[SystemIdentityName] = o.Name
	identity[SystemIdentityVersion] = o.Version

	return identity
}

// GetIdentityDigest returns the digest of the object's identity.
func (o *ElementMeta) GetIdentityDigest(accessor ElementAccessor) []byte {
	return o.GetIdentity(accessor).Digest()
}

func (o *ElementMeta) Copy() *ElementMeta {
	if o == nil {
		return nil
	}
	return &ElementMeta{
		Name:          o.Name,
		Version:       o.Version,
		ExtraIdentity: o.ExtraIdentity.Copy(),
		Labels:        o.Labels.Copy(),
	}
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
	GetLabels() metav1.Labels
	// SetLabels sets the labels of the object.
	SetLabels(labels []metav1.Label)
}

// ObjectMetaAccessor describes a accessor for named and versioned object.
type ObjectMetaAccessor interface {
	NameAccessor
	VersionAccessor
	LabelsAccessor
}

// ElementMetaAccessor provides generic access an elements meta information
type ElementMetaAccessor interface {
	GetMeta() *ElementMeta
}

// ElementAccessor provides generic access to list of elements
type ElementAccessor interface {
	Len() int
	Get(i int) ElementMetaAccessor
}

// AccessSpec is an abstract specification of an access method
// The outbound object is typicall a runtime.UnstructuredTypedObject.
// Inbound any serializable AccessSpec implementation is possible.
type AccessSpec interface {
	runtime.VersionedTypedObject
}

// GenericAccessSpec returns a generic AccessSpec implementation for an unstructured object.
// It can always be used instead of a dedicated access spec implementation. The core
// methods will map these spec into effective ones before an access is returned to the caller.
func GenericAccessSpec(un *runtime.UnstructuredTypedObject) AccessSpec {
	return &runtime.UnstructuredVersionedTypedObject{
		*un.DeepCopy(),
	}
}

// Sources describes a set of source specifications
type Sources []Source

var _ ElementAccessor = Sources{}

func (s Sources) Len() int {
	return len(s)
}

func (s Sources) Get(i int) ElementMetaAccessor {
	return &s[i]
}

func (s Sources) Copy() Sources {
	if s == nil {
		return nil
	}
	out := make(Sources, len(s))
	for i, v := range s {
		out[i] = *v.Copy()
	}
	return out
}

// Source is the definition of a component's source.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type Source struct {
	SourceMeta `json:",inline"`
	Access     AccessSpec `json:"access"`
}

func (s *Source) GetMeta() *ElementMeta {
	return &s.ElementMeta
}

func (s *Source) Copy() *Source {
	return &Source{
		SourceMeta: *s.SourceMeta.Copy(),
		Access:     s.Access,
	}
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

// Copy copies a source meta
func (o *SourceMeta) Copy() *SourceMeta {
	if o == nil {
		return nil
	}
	return &SourceMeta{
		ElementMeta: *o.ElementMeta.Copy(),
		Type:        o.Type,
	}
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

// Copy copy a source ref
func (r *SourceRef) Copy() *SourceRef {
	if r == nil {
		return nil
	}
	return &SourceRef{
		IdentitySelector: r.IdentitySelector.Copy(),
		Labels:           r.Labels.Copy(),
	}
}

type SourceRefs []SourceRef

// Copy copies a list of source refs
func (r SourceRefs) Copy() SourceRefs {
	if r == nil {
		return nil
	}

	result := make(SourceRefs, len(r))
	for i, v := range r {
		result[i] = *v.Copy()
	}
	return result
}

// Resources describes a set of resource specifications
type Resources []Resource

var _ ElementAccessor = Resources{}

func (r Resources) Len() int {
	return len(r)
}

func (r Resources) Get(i int) ElementMetaAccessor {
	return &r[i]
}

func (r Resources) Copy() Resources {
	if r == nil {
		return nil
	}
	out := make(Resources, len(r))
	for i, v := range r {
		out[i] = *v.Copy()
	}
	return out
}

// Resource describes a resource dependency of a component.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type Resource struct {
	ResourceMeta `json:",inline"`
	// Access describes the type specific method to
	// access the defined resource.
	Access AccessSpec `json:"access"`
}

func (r *Resource) GetMeta() *ElementMeta {
	return &r.ElementMeta
}

func (r *Resource) Copy() *Resource {
	return &Resource{
		ResourceMeta: *r.ResourceMeta.Copy(),
		Access:       r.Access,
	}
}

// ResourceMeta describes the meta data of a resource.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type ResourceMeta struct {
	ElementMeta `json:",inline"`

	// Type describes the type of the object.
	Type string `json:"type"`

	// Relation describes the relation of the resource to the component.
	// Can be a local or external resource
	Relation metav1.ResourceRelation `json:"relation,omitempty"`

	// SourceRef defines a list of source names.
	// These names reference the sources defines in `component.sources`.
	SourceRef SourceRefs `json:"srcRef,omitempty"`
}

// GetType returns the type of the object.
func (o ResourceMeta) GetType() string {
	return o.Type
}

// SetType sets the type of the object.
func (o *ResourceMeta) SetType(ttype string) {
	o.Type = ttype
}

// Copy copies a resource meta
func (o *ResourceMeta) Copy() *ResourceMeta {
	if o == nil {
		return nil
	}
	r := &ResourceMeta{
		ElementMeta: *o.ElementMeta.Copy(),
		Type:        o.Type,
		Relation:    o.Relation,
		SourceRef:   o.SourceRef.Copy(),
	}
	return r
}

type ComponentReferences []ComponentReference

func (r ComponentReferences) Len() int {
	return len(r)
}

func (r ComponentReferences) Get(i int) ElementMetaAccessor {
	return &r[i]
}

func (r ComponentReferences) Copy() ComponentReferences {
	if r == nil {
		return nil
	}
	out := make(ComponentReferences, len(r))
	for i, v := range r {
		out[i] = *v.Copy()
	}
	return out
}

// ComponentReference describes the reference to another component in the registry.
// +k8s:deepcopy-gen=true
// +k8s:openapi-gen=true
type ComponentReference struct {
	ElementMeta `json:",inline"`
	// ComponentName describes the remote name of the referenced object
	ComponentName string `json:"componentName"`
}

func (r *ComponentReference) GetMeta() *ElementMeta {
	return &r.ElementMeta
}

func (r *ComponentReference) Copy() *ComponentReference {
	return &ComponentReference{
		ElementMeta:   *r.ElementMeta.Copy(),
		ComponentName: r.ComponentName,
	}
}
