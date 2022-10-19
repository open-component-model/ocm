// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/errors"
)

func ResolveReferencePath(cv ocm.ComponentVersionAccess, path []metav1.Identity, resolver ocm.ComponentVersionResolver) (ocm.ComponentVersionAccess, error) {
	eff := cv
	if cv == nil {
		return nil, fmt.Errorf("no component version specified")
	}
	for _, cr := range path {
		if eff != cv {
			defer eff.Close()
		}
		compundResolver := ocm.NewCompoundResolver(eff.Repository(), resolver)
		cref, err := cv.GetReference(cr)
		if err != nil {
			return nil, err
		}

		eff, err = compundResolver.LookupComponentVersion(cref.GetComponentName(), cref.GetVersion())
		if err != nil {
			return nil, errors.Wrapf(err, "cannot resolve component reference")
		}
		if eff == nil {
			return nil, errors.ErrNotFound(ocm.KIND_COMPONENTREFERENCE, cref.String())
		}
	}
	return eff, nil
}

func MatchResourceReference(cv ocm.ComponentVersionAccess, typ string, ref metav1.ResourceReference, resolver ocm.ComponentVersionResolver) (ocm.ResourceAccess, ocm.ComponentVersionAccess, error) {
	eff, err := ResolveReferencePath(cv, ref.ReferencePath, resolver)
	if err != nil {
		return nil, nil, err
	}

	if len(eff.GetDescriptor().Resources) == 0 && len(ref.Resource) == 0 {
		return nil, nil, errors.ErrNotFound(ocm.KIND_RESOURCE)
	}
outer:
	for i, r := range eff.GetDescriptor().Resources {
		if r.Type != typ && typ != "" {
			continue
		}
		for k, v := range ref.Resource {
			switch k {
			case metav1.SystemIdentityName:
				if v != r.Name {
					continue outer
				}
			case metav1.SystemIdentityVersion:
				if v != r.Version {
					continue outer
				}
			default:
				if r.ExtraIdentity == nil || r.ExtraIdentity[k] != v {
					continue outer
				}
			}
		}
		res, err := eff.GetResourceByIndex(i)
		if err != nil {
			if eff != cv {
				eff.Close()
			}
			return nil, nil, err
		}
		return res, eff, nil
	}
	if eff != cv {
		eff.Close()
	}
	return nil, nil, errors.ErrNotFound(ocm.KIND_RESOURCE, ref.Resource.String())
}

func ResolveResourceReference(cv ocm.ComponentVersionAccess, ref metav1.ResourceReference, resolver ocm.ComponentVersionResolver) (ocm.ResourceAccess, ocm.ComponentVersionAccess, error) {
	if len(ref.Resource) == 0 || len(ref.Resource["name"]) == 0 {
		return nil, nil, errors.Newf("at least resource name must be specified for resource reference")
	}

	eff, err := ResolveReferencePath(cv, ref.ReferencePath, resolver)
	if err != nil {
		return nil, nil, err
	}
	eff = Dup(cv, eff)
	r, err := eff.GetResource(ref.Resource)
	if err != nil {
		eff.Close()
		return nil, nil, err
	}
	return r, eff, nil
}

func Dup(orig ocm.ComponentVersionAccess, newcva ocm.ComponentVersionAccess) ocm.ComponentVersionAccess {
	if orig != newcva {
		return newcva
	}
	return &nopCloserAccess{orig}
}

// TODO: provide a Dup method on cv to get another separately closable view

type nopCloserAccess struct {
	ocm.ComponentVersionAccess
}

func (n *nopCloserAccess) Close() error {
	return nil
}
