// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package gardenerconfig

import (
	"github.com/open-component-model/ocm/v2/pkg/contexts/credentials/cpi"
)

type credentialGetter struct {
	getCredentials func() (cpi.Credentials, error)
}

var _ cpi.CredentialsSource = credentialGetter{}

func (c credentialGetter) Credentials(ctx cpi.Context, cs ...cpi.CredentialsSource) (cpi.Credentials, error) {
	return c.getCredentials()
}
