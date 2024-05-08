package compdesc

import (
	"github.com/open-component-model/ocm/pkg/logging"
)

var (
	REALM  = logging.DefineSubRealm("component descriptor handling", "compdesc")
	Logger = logging.DynamicLogger(REALM)
)
