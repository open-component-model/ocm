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
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/errors"
)

func Execute(d Driver, name string, rid metav1.Identity, credsrc accessio.DataSource, paramsrc accessio.DataSource, octx ocm.Context, cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver) (*OperationResult, error) {
	var creds *Credentials
	var params []byte
	var err error

	if paramsrc != nil {
		params, err = paramsrc.Get()
		if err != nil {
			return nil, errors.Wrapf(err, "parameters")
		}
	}

	if credsrc != nil {
		data, err := credsrc.Get()
		if err == nil {
			creds, err = ParseCredentialSpecification(data, credsrc.Origin())
		}
		if err != nil {
			return nil, errors.Wrapf(err, "credentials")
		}
	}

	ires, _, err := utils.MatchResourceReference(cv, TypeTOIPackage, metav1.NewResourceRef(rid), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "installer resource in %s", common.VersionedElementKey(cv).String())
	}

	var spec PackageSpecification

	err = GetResource(ires, &spec)
	if err != nil {
		return nil, errors.ErrInvalidWrap(err, "installer spec")
	}
	return ExecuteAction(d, name, &spec, creds, params, octx, cv, resolver)
}
