// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package install

import (
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio/blobaccess"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/toi"
)

func Execute(p common.Printer, d Driver, name string, rid metav1.Identity, credsrc blobaccess.DataSource, paramsrc blobaccess.DataSource, octx ocm.Context, cv ocm.ComponentVersionAccess, resolver ocm.ComponentVersionResolver) (*OperationResult, error) {
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

	ires, _, err := utils.MatchResourceReference(cv, toi.TypeTOIPackage, metav1.NewResourceRef(rid), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "package resource in %s", common.VersionedElementKey(cv).String())
	}

	var spec toi.PackageSpecification

	err = GetResource(ires, &spec)
	if err != nil {
		return nil, errors.ErrInvalidWrap(err, "package spec")
	}
	return ExecuteAction(p, d, name, &spec, creds, params, octx, cv, resolver)
}
