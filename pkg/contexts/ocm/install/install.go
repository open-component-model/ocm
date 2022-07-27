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

package install

import (
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const TypeOCMInstaller = "ocmInstaller"

func Install(d Driver, name string, rid metav1.Identity, params []byte, octx ocm.Context, cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver) (*OperationResult, error) {
	var ires ocm.ResourceAccess
	var err error

	if len(rid) == 0 {
		for i, r := range cv.GetDescriptor().Resources {
			if r.Type == TypeOCMInstaller {
				ires, err = cv.GetResourceByIndex(i)
				if err != nil {
					return nil, errors.Wrapf(err, "cannot access installer resource %d", i)
				}
				break
			}
		}
		if ires == nil {
			return nil, errors.ErrNotFound("installer resource", common.VersionedElementKey(cv).String())
		}
	} else {
		ires, err = cv.GetResource(rid)
		if err != nil {
			return nil, errors.Wrapf(err, "installer resource %s", rid)
		}
	}

	m, err := ires.AccessMethod()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to instantiate access")
	}
	data, err := m.Get()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get resource content")
	}
	var spec Specification
	err = runtime.DefaultYAMLEncoding.Unmarshal(data, &spec)
	if err != nil {
		return nil, errors.ErrInvalidWrap(err, "installer spec")
	}
	return ExecuteAction(d, name, &spec, params, octx, cv, resolver)
}
