package compdesc

import (
	"bytes"
	"slices"

	"github.com/mandelsoft/goutils/errors"

	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/selectors"
	"ocm.software/ocm/api/ocm/selectors/refsel"
	"ocm.software/ocm/api/ocm/selectors/rscsel"
	"ocm.software/ocm/api/ocm/selectors/srcsel"
	"ocm.software/ocm/api/utils/runtime"
)

// GetEffectiveRepositoryContext returns the currently active repository context.
func (cd *ComponentDescriptor) GetEffectiveRepositoryContext() *runtime.UnstructuredTypedObject {
	if len(cd.RepositoryContexts) == 0 {
		return nil
	}
	return cd.RepositoryContexts[len(cd.RepositoryContexts)-1]
}

// AddRepositoryContext appends the given repository context to components descriptor repository history.
// The context is not appended if the effective repository context already matches the current context.
func (cd *ComponentDescriptor) AddRepositoryContext(repoCtx runtime.TypedObject) error {
	effective, err := runtime.ToUnstructuredTypedObject(cd.GetEffectiveRepositoryContext())
	if err != nil {
		return err
	}
	uRepoCtx, err := runtime.ToUnstructuredTypedObject(repoCtx)
	if err != nil {
		return err
	}
	if !runtime.UnstructuredTypesEqual(effective, uRepoCtx) {
		cd.RepositoryContexts = append(cd.RepositoryContexts, uRepoCtx)
	}
	return nil
}

func (cd *ComponentDescriptor) SelectResources(sel ...rscsel.Selector) (Resources, error) {
	err := selectors.ValidateSelectors(sel...)
	if err != nil {
		return nil, err
	}

	list := MapToSelectorElementList(cd.Resources)
	result := Resources{}
outer:
	for _, r := range cd.Resources {
		if len(sel) > 0 {
			mr := MapToSelectorResource(&r)
			for _, s := range sel {
				if !s.MatchResource(list, mr) {
					continue outer
				}
			}
		}
		result = append(result, r)
	}
	return result, nil
}

func (cd *ComponentDescriptor) GetResources() Resources {
	return slices.Clone(cd.Resources)
}

// GetResourceByIdentity returns resource that matches the given identity.
func (cd *ComponentDescriptor) GetResourceByIdentity(id v1.Identity) (Resource, error) {
	dig := id.Digest()
	for _, res := range cd.Resources {
		if bytes.Equal(res.GetIdentityDigest(cd.Resources), dig) {
			return res, nil
		}
	}
	return Resource{}, NotFound
}

// GetResourceIndexByIdentity returns the index of the resource that matches the given identity.
func (cd *ComponentDescriptor) GetResourceIndexByIdentity(id v1.Identity) int {
	dig := id.Digest()
	for i, res := range cd.Resources {
		if bytes.Equal(res.GetIdentityDigest(cd.Resources), dig) {
			return i
		}
	}
	return -1
}

// GetResourceIndex returns the index of a given resource.
// If the index is not found -1 is returned.
func (cd *ComponentDescriptor) GetResourceIndex(res *ResourceMeta) int {
	return ElementIndex(cd.Resources, res)
}

func (cd *ComponentDescriptor) SelectSources(sel ...srcsel.Selector) (Sources, error) {
	err := selectors.ValidateSelectors(sel...)
	if err != nil {
		return nil, err
	}

	list := MapToSelectorElementList(cd.Sources)
	result := Sources{}
outer:
	for _, r := range cd.Sources {
		if len(sel) > 0 {
			mr := MapToSelectorSource(&r)
			for _, s := range sel {
				if !s.MatchSource(list, mr) {
					continue outer
				}
			}
		}
		result = append(result, r)
	}
	return result, nil
}

func (cd *ComponentDescriptor) GetSources() Sources {
	return slices.Clone(cd.Sources)
}

// GetSourceByIdentity returns source that match the given identity.
func (cd *ComponentDescriptor) GetSourceByIdentity(id v1.Identity) (Source, error) {
	dig := id.Digest()
	for _, res := range cd.Sources {
		if bytes.Equal(res.GetIdentityDigest(cd.Sources), dig) {
			return res, nil
		}
	}
	return Source{}, NotFound
}

// GetSourceIndexByIdentity returns the index of the source that matches the given identity.
func (cd *ComponentDescriptor) GetSourceIndexByIdentity(id v1.Identity) int {
	dig := id.Digest()
	for i, res := range cd.Sources {
		if bytes.Equal(res.GetIdentityDigest(cd.Sources), dig) {
			return i
		}
	}
	return -1
}

// GetSourceIndex returns the index of a given source.
// If the index is not found -1 is returned.
func (cd *ComponentDescriptor) GetSourceIndex(src *SourceMeta) int {
	return ElementIndex(cd.Sources, src)
}

// GetReferenceByIdentity returns reference that matches the given identity.
func (cd *ComponentDescriptor) GetReferenceByIdentity(id v1.Identity) (Reference, error) {
	dig := id.Digest()
	for _, ref := range cd.References {
		if bytes.Equal(ref.GetIdentityDigest(cd.Resources), dig) {
			return ref, nil
		}
	}
	return Reference{}, errors.ErrNotFound(KIND_REFERENCE, id.String())
}

func (cd *ComponentDescriptor) SelectReferences(sel ...refsel.Selector) (References, error) {
	err := selectors.ValidateSelectors(sel...)
	if err != nil {
		return nil, err
	}

	list := MapToSelectorElementList(cd.References)
	result := References{}
outer:
	for _, r := range cd.References {
		if len(sel) > 0 {
			mr := MapToSelectorReference(&r)
			for _, s := range sel {
				if !s.MatchReference(list, mr) {
					continue outer
				}
			}
		}
		result = append(result, r)
	}
	return result, nil
}

func (cd *ComponentDescriptor) GetReferences() References {
	return slices.Clone(cd.References)
}

// GetReferenceIndexByIdentity returns the index of the reference that matches the given identity.
func (cd *ComponentDescriptor) GetReferenceIndexByIdentity(id v1.Identity) int {
	dig := id.Digest()
	for i, ref := range cd.References {
		if bytes.Equal(ref.GetIdentityDigest(cd.Resources), dig) {
			return i
		}
	}
	return -1
}

// GetReferenceIndex returns the index of a given source.
// If the index is not found -1 is returned.
func (cd *ComponentDescriptor) GetReferenceIndex(src ElementMetaProvider) int {
	return ElementIndex(cd.References, src)
}

// GetSignatureIndex returns the index of the signature with the given name
// If the index is not found -1 is returned.
func (cd *ComponentDescriptor) GetSignatureIndex(name string) int {
	return cd.Signatures.GetIndex(name)
}
