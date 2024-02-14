// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package npm

import (
	"github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"
	"github.com/open-component-model/ocm/pkg/logging"
)

const (
	// CONSUMER_TYPE is the npm repository type.
	CONSUMER_TYPE     = "Registry.npmjs.com"
	BLOB_HANDLER_NAME = "ocm/npmPackage"

	// ATTR_USERNAME is the username attribute. Required for login at any npm registry.
	ATTR_USERNAME = cpi.ATTR_USERNAME
	// ATTR_PASSWORD is the password attribute. Required for login at any npm registry.
	ATTR_PASSWORD = cpi.ATTR_PASSWORD
	// ATTR_EMAIL is the email attribute. Required for login at any npm registry.
	ATTR_EMAIL = cpi.ATTR_EMAIL
)

// Logging Realm.
var REALM = logging.DefineSubRealm("NPM registry", "NPM")
