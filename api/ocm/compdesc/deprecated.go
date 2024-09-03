package compdesc

import (
	"bytes"
	"fmt"

	"github.com/mandelsoft/goutils/sliceutils"

	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/utils/selector"
)

// GetResourceAccessByIdentity returns a pointer to the resource that matches the given identity.
//
// Deprecated: use GetResourceByIdentity.
func (cd *ComponentDescriptor) GetResourceAccessByIdentity(id v1.Identity) *Resource {
	dig := id.Digest()
	for i, res := range cd.Resources {
		if bytes.Equal(res.GetIdentityDigest(cd.Resources), dig) {
			return &cd.Resources[i]
		}
	}
	return nil
}

// GetResourceByRegexSelector returns resources that match the given selectors.
//
// Deprecated: use GetResources with appropriate selectors.
func (cd *ComponentDescriptor) GetResourceByRegexSelector(sel interface{}) (Resources, error) {
	identitySelector, err := selector.ParseRegexSelector(sel)
	if err != nil {
		return nil, fmt.Errorf("unable to parse selector: %w", err)
	}
	return cd.GetResourcesByIdentitySelectors(identitySelector)
}

// GetResourcesByIdentitySelectors returns resources that match the given identity selectors.
// Deprecated: use GetResources with appropriate selectors.
func (cd *ComponentDescriptor) GetResourcesByIdentitySelectors(selectors ...IdentitySelector) (Resources, error) {
	return cd.GetResourcesBySelectors(selectors, nil)
}

// GetResourcesByResourceSelectors returns resources that match the given resource selectors.
//
// Deprecated: use GetResources with appropriate selectors.
func (cd *ComponentDescriptor) GetResourcesByResourceSelectors(selectors ...ResourceSelector) (Resources, error) {
	return cd.GetResourcesBySelectors(nil, selectors)
}

// GetResourcesBySelectors returns resources that match the given selector.
//
// Deprecated: use GetResources with appropriate selectors.
func (cd *ComponentDescriptor) GetResourcesBySelectors(selectors []IdentitySelector, resourceSelectors []ResourceSelector) (Resources, error) {
	resources := make(Resources, 0)
	for i := range cd.Resources {
		selctx := NewResourceSelectionContext(i, cd.Resources)
		if len(selectors) > 0 {
			ok, err := selector.MatchSelectors(selctx.Identity(), selectors...)
			if err != nil {
				return nil, fmt.Errorf("unable to match selector for resource %s: %w", selctx.Name, err)
			}
			if !ok {
				continue
			}
		}
		ok, err := MatchResourceByResourceSelector(selctx, resourceSelectors...)
		if err != nil {
			return nil, fmt.Errorf("unable to match selector for resource %s: %w", selctx.Name, err)
		}
		if !ok {
			continue
		}
		resources = append(resources, *selctx.Resource)
	}
	if len(resources) == 0 {
		return resources, NotFound
	}
	return resources, nil
}

// GetExternalResources returns external resource with the given type, name and version.
//
// Deprecated: use GetResources with appropriate selectors.
func (cd *ComponentDescriptor) GetExternalResources(rtype, name, version string) (Resources, error) {
	return cd.GetResourcesBySelectors(
		[]selector.Interface{
			ByName(name),
			ByVersion(version),
		},
		[]ResourceSelector{
			ByResourceType(rtype),
			ByRelation(v1.ExternalRelation),
		})
}

// GetExternalResource returns external resource with the given type, name and version.
//
// If multiple resources match, the first one is returned.
// Deprecated: use GetResources with appropriate selectors.
func (cd *ComponentDescriptor) GetExternalResource(rtype, name, version string) (Resource, error) {
	resources, err := cd.GetExternalResources(rtype, name, version)
	if err != nil {
		return Resource{}, err
	}
	// at least one resource must be defined, otherwise the getResourceBySelectors functions returns a NotFound err.
	return resources[0], nil
}

// GetLocalResources returns all local resources with the given type, name and version.
func (cd *ComponentDescriptor) GetLocalResources(rtype, name, version string) (Resources, error) {
	return cd.GetResourcesBySelectors(
		[]selector.Interface{
			ByName(name),
			ByVersion(version),
		},
		[]ResourceSelector{
			ByResourceType(rtype),
			ByRelation(v1.LocalRelation),
		})
}

// GetLocalResource returns a local resource with the given type, name and version.
//
// If multiple resources match, the first one is returned.
// Deprecated: use GetResources with appropriate selectors.
func (cd *ComponentDescriptor) GetLocalResource(rtype, name, version string) (Resource, error) {
	resources, err := cd.GetLocalResources(rtype, name, version)
	if err != nil {
		return Resource{}, err
	}
	// at least one resource must be defined, otherwise the getResourceBySelectors functions returns a NotFound err.
	return resources[0], nil
}

// GetResourcesByType returns all resources that match the given type and selectors.
//
// Deprecated: use GetResources with appropriate selectors.
func (cd *ComponentDescriptor) GetResourcesByType(rtype string, selectors ...IdentitySelector) (Resources, error) {
	return cd.GetResourcesBySelectors(
		selectors,
		[]ResourceSelector{
			ByResourceType(rtype),
		})
}

// GetResourcesByName returns all local and external resources with a name.
//
// Deprecated: use GetResources with appropriate selectors.
func (cd *ComponentDescriptor) GetResourcesByName(name string, selectors ...IdentitySelector) (Resources, error) {
	return cd.GetResourcesBySelectors(
		sliceutils.CopyAppend[IdentitySelector](selectors, ByName(name)),
		nil)
}

////////////////////////////////////////////////////////////////////////////////

// GetSourceAccessByIdentity returns a pointer to the source that matches the given identity.
//
// Deprecated: use GetSourceByIdentity.
func (cd *ComponentDescriptor) GetSourceAccessByIdentity(id v1.Identity) *Source {
	dig := id.Digest()
	for i, res := range cd.Sources {
		if bytes.Equal(res.GetIdentityDigest(cd.Sources), dig) {
			return &cd.Sources[i]
		}
	}
	return nil
}

// GetSourcesByIdentitySelectors returns references that match the given selector.
//
// Deprecated: use GetSources with appropriate selectors.
func (cd *ComponentDescriptor) GetSourcesByIdentitySelectors(selectors ...IdentitySelector) (Sources, error) {
	srcs := make(Sources, 0)
	for _, src := range cd.Sources {
		ok, err := selector.MatchSelectors(src.GetIdentity(cd.Sources), selectors...)
		if err != nil {
			return nil, fmt.Errorf("unable to match selector for source %s: %w", src.Name, err)
		}
		if ok {
			srcs = append(srcs, src)
		}
	}
	if len(srcs) == 0 {
		return srcs, NotFound
	}
	return srcs, nil
}

// GetSourcesByName returns all sources with a name.
//
// Deprecated: use GetSources with appropriate selectors.
func (cd *ComponentDescriptor) GetSourcesByName(name string, selectors ...IdentitySelector) (Sources, error) {
	return cd.GetSourcesByIdentitySelectors(
		sliceutils.CopyAppend[IdentitySelector](selectors, ByName(name))...)
}

////////////////////////////////////////////////////////////////////////////////

// GetComponentReferences returns all component references that matches the given selectors.
//
// Deprectated: use GetReferences with appropriate selectors.
func (cd *ComponentDescriptor) GetComponentReferences(selectors ...IdentitySelector) ([]Reference, error) {
	refs := make([]Reference, 0)
	for _, ref := range cd.References {
		ok, err := selector.MatchSelectors(ref.GetIdentity(cd.References), selectors...)
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

// GetComponentReferenceIndex returns the index of a given component reference.
// If the index is not found -1 is returned.
// Deprecated: use GetReferenceIndex.
func (cd *ComponentDescriptor) GetComponentReferenceIndex(ref Reference) int {
	return cd.GetReferenceIndex(ref.GetMeta())
}

// GetReferenceAccessByIdentity returns a pointer to the reference that matches the given identity.
// Deprectated: use GetReferenceByIdentity.
func (cd *ComponentDescriptor) GetReferenceAccessByIdentity(id v1.Identity) *Reference {
	dig := id.Digest()
	for i, ref := range cd.References {
		if bytes.Equal(ref.GetIdentityDigest(cd.Resources), dig) {
			return &cd.References[i]
		}
	}
	return nil
}

// GetReferencesByIdentitySelectors returns resources that match the given identity selectors.
// Deprectated: use GetReferences with appropriate selectors.
func (cd *ComponentDescriptor) GetReferencesByIdentitySelectors(selectors ...IdentitySelector) (References, error) {
	return cd.GetReferencesBySelectors(selectors, nil)
}

// GetReferencesByReferenceSelectors returns resources that match the given resource selectors.
// Deprectated: use GetReferences with appropriate selectors.
func (cd *ComponentDescriptor) GetReferencesByReferenceSelectors(selectors ...ReferenceSelector) (References, error) {
	return cd.GetReferencesBySelectors(nil, selectors)
}

// GetReferencesBySelectors returns resources that match the given selector.
// Deprectated: use GetReferences with appropriate selectors.
func (cd *ComponentDescriptor) GetReferencesBySelectors(selectors []IdentitySelector, referenceSelectors []ReferenceSelector) (References, error) {
	references := make(References, 0)
	for i := range cd.References {
		selctx := NewReferenceSelectionContext(i, cd.References)
		if len(selectors) > 0 {
			ok, err := selector.MatchSelectors(selctx.Identity(), selectors...)
			if err != nil {
				return nil, fmt.Errorf("unable to match selector for resource %s: %w", selctx.Name, err)
			}
			if !ok {
				continue
			}
		}
		ok, err := MatchReferencesByReferenceSelector(selctx, referenceSelectors...)
		if err != nil {
			return nil, fmt.Errorf("unable to match selector for resource %s: %w", selctx.Name, err)
		}
		if !ok {
			continue
		}
		references = append(references, *selctx.Reference)
	}
	if len(references) == 0 {
		return references, NotFound
	}
	return references, nil
}

// GetReferencesByName returns references that match the given name.
// Deprectated: use GetReferences with appropriate selectors.
func (cd *ComponentDescriptor) GetReferencesByName(name string, selectors ...IdentitySelector) (References, error) {
	return cd.GetReferencesBySelectors(
		sliceutils.CopyAppend[IdentitySelector](selectors, ByName(name)),
		nil)
}
