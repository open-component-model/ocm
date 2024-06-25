package maven

import "github.com/open-component-model/ocm/api/utils/logging"

var REALM = logging.DefineSubRealm("Maven repository", "maven")

var Log = logging.DynamicLogger(REALM)
