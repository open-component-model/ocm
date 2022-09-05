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

package utils

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/errors"
)

func ResolveReferencePath(cv ocm.ComponentVersionAccess, path []metav1.Identity, resolver ocm.ComponentVersionResolver) (ocm.ComponentVersionAccess, error) {
	eff := cv
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
