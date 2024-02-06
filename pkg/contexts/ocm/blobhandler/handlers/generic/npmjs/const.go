package npmjs

import "github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"

const (
	// CONSUMER_TYPE is the npmjs repository type.
	CONSUMER_TYPE     = "Registry.npmjs.com"
	BLOB_HANDLER_NAME = "ocm/npmPackage"

	// ATTR_USERNAME is the username attribute. Required for login at any npmjs registry.
	ATTR_USERNAME = cpi.ATTR_USERNAME
	// ATTR_PASSWORD is the password attribute. Required for login at any npmjs registry.
	ATTR_PASSWORD = cpi.ATTR_PASSWORD
	// ATTR_EMAIL is the email attribute. Required for login at any npmjs registry.
	ATTR_EMAIL = cpi.ATTR_EMAIL
)
