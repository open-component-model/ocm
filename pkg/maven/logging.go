package maven

import "github.com/open-component-model/ocm/pkg/logging"

var REALM = logging.DefineSubRealm("Maven repository", "maven")

var Log = logging.DynamicLogger(REALM)
