// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ocireg

import (
	"path"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/oci/identity"
	"github.com/open-component-model/ocm/pkg/utils"
)

func GetCredentials(ctx credentials.ContextProvider, locator, repo string) (credentials.Credentials, error) {
	return credentials.CredentialsForConsumer(ctx.CredentialsContext(), GetConsumerId(locator, repo), identity.IdentityMatcher)
}

func GetConsumerId(locator, repo string) credentials.ConsumerIdentity {
	host, port, base := utils.SplitLocator(locator)
	id := credentials.ConsumerIdentity{
		identity.ID_TYPE:     identity.CONSUMER_TYPE,
		identity.ID_HOSTNAME: host,
	}
	if port != "" {
		id[identity.ID_PORT] = port
	}
	if repo == "" {
		id[identity.ID_PATHPREFIX] = base
	} else {
		id[identity.ID_PATHPREFIX] = path.Join(base, repo)
	}
	return id
}
