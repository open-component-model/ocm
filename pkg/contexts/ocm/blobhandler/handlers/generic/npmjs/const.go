package npmjs

import "github.com/open-component-model/ocm/pkg/contexts/credentials/cpi"

// CONSUMER_TYPE is the npmjs repository type.
const CONSUMER_TYPE = "Registry.npmjs.com"
const BLOB_HANDLER_NAME = "ocm/npmPackage"

const (
	ATTR_USERNAME = cpi.ATTR_USERNAME
	ATTR_PASSWORD = cpi.ATTR_PASSWORD
	ATTR_EMAIL    = cpi.ATTR_EMAIL
)
