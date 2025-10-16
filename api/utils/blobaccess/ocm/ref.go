package ocm

import (
	"github.com/mandelsoft/goutils/errors"

	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/resourcerefs"
	"ocm.software/ocm/api/ocm/selectors/rscsel"
	"ocm.software/ocm/api/utils"
)

// ResourceProvider selects a resource from a component version.
// It should not hold any separately closeabvle view on
// an object. The lifecycle of those objects should be left
// to the creator of the implementation of this interface.
type ResourceProvider interface {
	GetResource(cv cpi.ComponentVersionAccess) (cpi.ResourceAccess, cpi.ComponentVersionAccess, error)
}

////////////////////////////////////////////////////////////////////////////////

type byid struct {
	id metav1.Identity
}

var _ ResourceProvider = (*byid)(nil)

func ByResourceId(id metav1.Identity) ResourceProvider {
	return &byid{id}
}

func (r *byid) GetResource(cv cpi.ComponentVersionAccess) (cpi.ResourceAccess, cpi.ComponentVersionAccess, error) {
	cv, err := cv.Dup()
	if err != nil {
		return nil, nil, err
	}
	res, err := cv.GetResource(r.id)
	if err != nil {
		cv.Close()
		return nil, nil, err
	}
	return res, cv, nil
}

////////////////////////////////////////////////////////////////////////////////

type byref struct {
	resolver cpi.ComponentVersionResolver
	ref      metav1.ResourceReference
}

var _ ResourceProvider = (*byref)(nil)

func ByResourcePath(id metav1.Identity, path ...metav1.Identity) ResourceProvider {
	return &byref{nil, metav1.NewNestedResourceRef(id, path)}
}

func ByResourceRef(ref metav1.ResourceReference, res ...cpi.ComponentVersionResolver) ResourceProvider {
	return &byref{utils.Optional(res...), ref}
}

func (r *byref) GetResource(cv cpi.ComponentVersionAccess) (cpi.ResourceAccess, cpi.ComponentVersionAccess, error) {
	return resourcerefs.ResolveResourceReference(cv, r.ref, r.resolver)
}

////////////////////////////////////////////////////////////////////////////////

type bysel struct {
	sel []rscsel.Selector
}

var _ ResourceProvider = (*bysel)(nil)

func ByResourceSelector(sel ...rscsel.Selector) ResourceProvider {
	return &bysel{sel}
}

func (r *bysel) GetResource(cv cpi.ComponentVersionAccess) (cpi.ResourceAccess, cpi.ComponentVersionAccess, error) {
	res, err := cv.SelectResources(r.sel...)
	if err != nil {
		return nil, nil, err
	}
	if len(res) == 0 {
		return nil, nil, errors.ErrNotFound(cpi.KIND_RESOURCE)
	}

	cv, err = cv.Dup()
	if err != nil {
		return nil, nil, err
	}
	return res[0], cv, nil
}
