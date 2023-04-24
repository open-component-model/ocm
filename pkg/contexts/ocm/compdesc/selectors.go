// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package compdesc

import (
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/utils/selector"
)

type IdentitySelector = selector.Interface

// ResourceSelectorFunc defines a function to filter a resource.
type ResourceSelectorFunc func(obj Resource) (bool, error)

var _ ResourceSelector = ResourceSelectorFunc(nil)

func (s ResourceSelectorFunc) MatchResource(obj Resource) (bool, error) {
	return s(obj)
}

// ResourceSelector defines a selector bases on resource attributes.
type ResourceSelector interface {
	MatchResource(obj Resource) (bool, error)
}

// IdentityAndResourceSelector is selector, which can act as
// resource and/or identity selector.
type IdentityAndResourceSelector interface {
	IdentitySelector
	ResourceSelector
}

// MatchResourceByResourceSelector applies all resource selector against the given resource object.
func MatchResourceByResourceSelector(obj Resource, resourceSelectors ...ResourceSelector) (bool, error) {
	for _, sel := range resourceSelectors {
		ok, err := sel.MatchResource(obj)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// ByResourceType creates a new resource selector that
// selects a resource based on its type.
func ByResourceType(ttype string) ResourceSelector {
	return ResourceSelectorFunc(func(obj Resource) (bool, error) {
		return ttype == "" || obj.GetType() == ttype, nil
	})
}

// ByRelation creates a new resource selector that
// selects a resource based on its relation type.
func ByRelation(relation v1.ResourceRelation) ResourceSelectorFunc {
	return ResourceSelectorFunc(func(obj Resource) (bool, error) {
		return obj.Relation == relation, nil
	})
}

type byVersion struct {
	version string
}

var (
	_ ResourceSelector = (*byVersion)(nil)
	_ IdentitySelector = (*byVersion)(nil)
)

func (b *byVersion) Match(obj map[string]string) (bool, error) {
	return obj[SystemIdentityVersion] == b.version, nil
}

func (b *byVersion) MatchResource(obj Resource) (bool, error) {
	return obj.GetVersion() == b.version, nil
}

// ByVersion creates a new resource and identity selector that
// selects a resource based on its version.
func ByVersion(version string) IdentityAndResourceSelector {
	return &byVersion{version: version}
}

type byName struct {
	name string
}

var (
	_ ResourceSelector = (*byName)(nil)
	_ IdentitySelector = (*byName)(nil)
)

func (b *byName) Match(obj map[string]string) (bool, error) {
	return obj[SystemIdentityName] == b.name, nil
}

func (b *byName) MatchResource(obj Resource) (bool, error) {
	return obj.GetName() == b.name, nil
}

// ByName creates a new selector that matches a resource name.
func ByName(name string) IdentityAndResourceSelector {
	return &byName{name: name}
}
