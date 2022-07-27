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

func ResolveResourceReference(cv ocm.ComponentVersionAccess, ref metav1.ResourceReference, resolver ocm.ComponentVersionResolver) (ocm.ResourceAccess, ocm.ComponentVersionAccess, error) {
	eff := cv

	if len(ref.Resource) == 0 || len(ref.Resource["name"]) == 0 {
		return nil, nil, errors.Newf("at least resource name must be specified for resource reference")
	}

	for _, cr := range ref.ReferencePath {
		if eff != cv {
			defer eff.Close()
		}
		resolver := ocm.NewCompoundResolver(eff.Repository(), resolver)
		cref, err := cv.GetReference(cr)
		if err != nil {
			return nil, nil, err
		}

		eff, err = resolver.LookupComponentVersion(cref.GetComponentName(), cref.GetVersion())
		if err != nil {
			return nil, nil, errors.Wrapf(err, "cannot resolve resource reference")
		}
	}

	r, err := eff.GetResource(ref.Resource)
	if err != nil {
		if eff != cv {
			eff.Close()
		}
		return nil, nil, err
	}
	return r, eff, nil
}
