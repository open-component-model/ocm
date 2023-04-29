// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/identity/hostpath"
)

// CONSUMER_TYPE is the Helm chart repository type.
const CONSUMER_TYPE = "HelmChartRepository"

// ID_TYPE is the type field of a consumer identity.
const ID_TYPE = cpi.ID_TYPE

// ID_SCHEME is the scheme of the repository.
const ID_SCHEME = hostpath.ID_SCHEME

// ID_HOSTNAME is the hostname of a repository.
const ID_HOSTNAME = hostpath.ID_HOSTNAME

// ID_PORT is the port number of a repository.
const ID_PORT = hostpath.ID_PORT

// ID_PATHPREFIX is the path of a repository.
const ID_PATHPREFIX = hostpath.ID_PATHPREFIX

func init() {
	cpi.RegisterStandardIdentityMatcher(CONSUMER_TYPE, IdentityMatcher, `Helm chart repository

It matches the <code>`+CONSUMER_TYPE+`</code> consumer type and additionally acts like 
the <code>`+hostpath.IDENTITY_TYPE+`</code> type.`)
}

var identityMatcher = hostpath.IdentityMatcher(CONSUMER_TYPE)

func IdentityMatcher(pattern, cur, id cpi.ConsumerIdentity) bool {
	return identityMatcher(pattern, cur, id)
}

// used crednetial attributes

const ATTR_USERNAME = credentials.ATTR_USERNAME

const ATTR_PASSWORD = credentials.ATTR_PASSWORD

const ATTR_CERTIFICATE = credentials.ATTR_CERTIFICATE
