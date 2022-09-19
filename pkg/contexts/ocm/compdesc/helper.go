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
	"bytes"
	"fmt"

	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/utils/selector"
)

type IdentitySelector = selector.Interface

// ResourceSelectorFunc defines a function to filter a resource.
type ResourceSelectorFunc = func(obj Resource) (bool, error)

// MatchResourceSelectorFuncs applies all resource selector against the given resource object.
func MatchResourceSelectorFuncs(obj Resource, resourceSelectors ...ResourceSelectorFunc) (bool, error) {
	for _, sel := range resourceSelectors {
		ok, err := sel(obj)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// NewTypeResourceSelector creates a new resource selector that
// selects a resource based on its type.
func NewTypeResourceSelector(ttype string) ResourceSelectorFunc {
	return func(obj Resource) (bool, error) {
		return obj.GetType() == ttype, nil
	}
}

// NewVersionResourceSelector creates a new resource selector that
// selects a resource based on its version.
func NewVersionResourceSelector(version string) ResourceSelectorFunc {
	return func(obj Resource) (bool, error) {
		return obj.GetVersion() == version, nil
	}
}

// NewRelationResourceSelector creates a new resource selector that
// selects a resource based on its relation type.
func NewRelationResourceSelector(relation v1.ResourceRelation) ResourceSelectorFunc {
	return func(obj Resource) (bool, error) {
		return obj.Relation == relation, nil
	}
}

// NewNameSelector creates a new selector that matches a resource name.
func NewNameSelector(name string) selector.Interface {
	return selector.DefaultSelector{
		SystemIdentityName: name,
	}
}

// GetEffectiveRepositoryContext returns the currently active repository context.
func (c *ComponentDescriptor) GetEffectiveRepositoryContext() *runtime.UnstructuredTypedObject {
	if len(c.RepositoryContexts) == 0 {
		return nil
	}
	return &c.RepositoryContexts[len(c.RepositoryContexts)-1]
}

// AddRepositoryContext appends the given repository context to components descriptor repository history.
// The context is not appended if the effective repository context already matches the current context.
func (c *ComponentDescriptor) AddRepositoryContext(repoCtx runtime.TypedObject) error {
	effective, err := runtime.ToUnstructuredTypedObject(c.GetEffectiveRepositoryContext())
	if err != nil {
		return err
	}
	uRepoCtx, err := runtime.ToUnstructuredTypedObject(repoCtx)
	if err != nil {
		return err
	}
	if !runtime.UnstructuredTypesEqual(effective, uRepoCtx) {
		c.RepositoryContexts = append(c.RepositoryContexts, *uRepoCtx)
	}
	return nil
}

// GetComponentReferences returns all component references that matches the given selectors.
func (c *ComponentDescriptor) GetComponentReferences(selectors ...IdentitySelector) ([]ComponentReference, error) {
	refs := make([]ComponentReference, 0)
	for _, ref := range c.References {
		ok, err := selector.MatchSelectors(ref.GetIdentity(c.References), selectors...)
		if err != nil {
			return nil, fmt.Errorf("unable to match selector for resource %s: %w", ref.Name, err)
		}
		if ok {
			refs = append(refs, ref)
		}
	}
	if len(refs) == 0 {
		return refs, NotFound
	}
	return refs, nil
}

// GetResourceByIdentity returns resource that match the given identity.
func (c *ComponentDescriptor) GetResourceByIdentity(id v1.Identity) (Resource, error) {
	dig := id.Digest()
	for _, res := range c.Resources {
		if bytes.Equal(res.GetIdentityDigest(c.Resources), dig) {
			return res, nil
		}
	}
	return Resource{}, NotFound
}

// GetComponentReferencesByName returns all component references with a given name.
func (c *ComponentDescriptor) GetComponentReferencesByName(name string) ([]ComponentReference, error) {
	return c.GetComponentReferences(NewNameSelector(name))
}

// GetResourceByJSONScheme returns resources that match the given selectors.
func (c *ComponentDescriptor) GetResourceByJSONScheme(src interface{}) ([]Resource, error) {
	sel, err := selector.NewJSONSchemaSelectorFromGoStruct(src)
	if err != nil {
		return nil, err
	}
	return c.GetResourcesBySelector(sel)
}

// GetResourceByDefaultSelector returns resources that match the given selectors.
func (c *ComponentDescriptor) GetResourceByDefaultSelector(sel interface{}) ([]Resource, error) {
	identitySelector, err := selector.ParseDefaultSelector(sel)
	if err != nil {
		return nil, fmt.Errorf("unable to parse selector: %w", err)
	}
	return c.GetResourcesBySelector(identitySelector)
}

// GetResourceByRegexSelector returns resources that match the given selectors.
func (c ComponentDescriptor) GetResourceByRegexSelector(sel interface{}) ([]Resource, error) {
	identitySelector, err := selector.ParseRegexSelector(sel)
	if err != nil {
		return nil, fmt.Errorf("unable to parse selector: %w", err)
	}
	return c.GetResourcesBySelector(identitySelector)
}

// GetResourcesBySelector returns resources that match the given selector.
func (c *ComponentDescriptor) GetResourcesBySelector(selectors ...IdentitySelector) ([]Resource, error) {
	resources := make([]Resource, 0)
	for _, res := range c.Resources {
		ok, err := selector.MatchSelectors(res.GetIdentity(c.Resources), selectors...)
		if err != nil {
			return nil, fmt.Errorf("unable to match selector for resource %s: %w", res.Name, err)
		}
		if ok {
			resources = append(resources, res)
		}
	}
	if len(resources) == 0 {
		return resources, NotFound
	}
	return resources, nil
}

// GetResourcesBySelector returns resources that match the given selector.
func (c *ComponentDescriptor) getResourceBySelectors(selectors []IdentitySelector, resourceSelectors []ResourceSelectorFunc) ([]Resource, error) {
	resources := make([]Resource, 0)
	for _, res := range c.Resources {
		ok, err := selector.MatchSelectors(res.GetIdentity(c.Resources), selectors...)
		if err != nil {
			return nil, fmt.Errorf("unable to match selector for resource %s: %w", res.Name, err)
		}
		if !ok {
			continue
		}
		ok, err = MatchResourceSelectorFuncs(res, resourceSelectors...)
		if err != nil {
			return nil, fmt.Errorf("unable to match selector for resource %s: %w", res.Name, err)
		}
		if !ok {
			continue
		}
		resources = append(resources, res)
	}
	if len(resources) == 0 {
		return resources, NotFound
	}
	return resources, nil
}

// GetExternalResources returns a external resource with the given type, name and version.
func (c *ComponentDescriptor) GetExternalResources(rtype, name, version string) ([]Resource, error) {
	return c.getResourceBySelectors(
		[]selector.Interface{NewNameSelector(name)},
		[]ResourceSelectorFunc{
			NewTypeResourceSelector(rtype),
			NewVersionResourceSelector(version),
			NewRelationResourceSelector(v1.ExternalRelation),
		})
}

// GetExternalResource returns a external resource with the given type, name and version.
// If multiple resources match, the first one is returned.
func (c *ComponentDescriptor) GetExternalResource(rtype, name, version string) (Resource, error) {
	resources, err := c.GetExternalResources(rtype, name, version)
	if err != nil {
		return Resource{}, err
	}
	// at least one resource must be defined, otherwise the getResourceBySelectors functions returns a NotFound err.
	return resources[0], nil
}

// GetLocalResources returns all local resources with the given type, name and version.
func (c *ComponentDescriptor) GetLocalResources(rtype, name, version string) ([]Resource, error) {
	return c.getResourceBySelectors(
		[]selector.Interface{NewNameSelector(name)},
		[]ResourceSelectorFunc{
			NewTypeResourceSelector(rtype),
			NewVersionResourceSelector(version),
			NewRelationResourceSelector(v1.LocalRelation),
		})
}

// GetLocalResource returns a local resource with the given type, name and version.
// If multiple resources match, the first one is returned.
func (c *ComponentDescriptor) GetLocalResource(rtype, name, version string) (Resource, error) {
	resources, err := c.GetLocalResources(rtype, name, version)
	if err != nil {
		return Resource{}, err
	}
	// at least one resource must be defined, otherwise the getResourceBySelectors functions returns a NotFound err.
	return resources[0], nil
}

// GetResourcesByType returns all resources that match the given type and selectors.
func (c *ComponentDescriptor) GetResourcesByType(rtype string, selectors ...IdentitySelector) ([]Resource, error) {
	return c.getResourceBySelectors(
		selectors,
		[]ResourceSelectorFunc{
			NewTypeResourceSelector(rtype),
		})
}

// GetResourcesByName returns all local and external resources with a name.
func (c *ComponentDescriptor) GetResourcesByName(name string, selectors ...IdentitySelector) ([]Resource, error) {
	return c.getResourceBySelectors(
		append(selectors, NewNameSelector(name)),
		nil)
}

// GetResourceIndex returns the index of a given resource.
// If the index is not found -1 is returned.
func (c *ComponentDescriptor) GetResourceIndex(res *ResourceMeta) int {
	id := res.GetIdentity(c.Resources)
	for i, cur := range c.Resources {
		if cur.GetIdentity(c.Resources).Equals(id) {
			return i
		}
	}
	return -1
}

// GetComponentReferenceIndex returns the index of a given component reference.
// If the index is not found -1 is returned.
func (c *ComponentDescriptor) GetComponentReferenceIndex(ref ComponentReference) int {
	id := ref.GetIdentityDigest(c.References)
	for i, cur := range c.References {
		if bytes.Equal(cur.GetIdentityDigest(c.References), id) {
			return i
		}
	}
	return -1
}

// GetComponentReferenceByIdentity returns reference that match the given identity.
func (c *ComponentDescriptor) GetComponentReferenceByIdentity(id v1.Identity) (ComponentReference, error) {
	dig := id.Digest()
	for _, ref := range c.References {
		if bytes.Equal(ref.GetIdentityDigest(c.References), dig) {
			return ref, nil
		}
	}
	return ComponentReference{}, NotFound
}

// GetSourceByIdentity returns source that match the given identity.
func (c *ComponentDescriptor) GetSourceByIdentity(id v1.Identity) (Source, error) {
	dig := id.Digest()
	for _, res := range c.Sources {
		if bytes.Equal(res.GetIdentityDigest(c.Resources), dig) {
			return res, nil
		}
	}
	return Source{}, NotFound
}

// GetSourceIndex returns the index of a given source.
// If the index is not found -1 is returned.
func (c *ComponentDescriptor) GetSourceIndex(src *SourceMeta) int {
	id := src.GetIdentityDigest(c.Sources)
	for i, cur := range c.Sources {
		if bytes.Equal(cur.GetIdentityDigest(c.Sources), id) {
			return i
		}
	}
	return -1
}

// GetReferenceByIdentity returns reference that match the given identity.
func (c *ComponentDescriptor) GetReferenceByIdentity(id v1.Identity) (ComponentReference, error) {
	dig := id.Digest()
	for _, ref := range c.References {
		if bytes.Equal(ref.GetIdentityDigest(c.Resources), dig) {
			return ref, nil
		}
	}
	return ComponentReference{}, NotFound
}

// GetReferenceIndex returns the index of a given source.
// If the index is not found -1 is returned.
func (c *ComponentDescriptor) GetReferenceIndex(src *ElementMeta) int {
	id := src.GetIdentityDigest(c.References)
	for i, cur := range c.References {
		if bytes.Equal(cur.GetIdentityDigest(c.References), id) {
			return i
		}
	}
	return -1
}

// GetSignatureIndex returns the index of the signature with the given name
// If the index is not found -1 is returned.
func (c *ComponentDescriptor) GetSignatureIndex(name string) int {
	for i, cur := range c.Signatures {
		if cur.Name == name {
			return i
		}
	}
	return -1
}
