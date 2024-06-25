package wget

import (
	ocmlog "github.com/open-component-model/ocm/api/utils/logging"
)

var REALM = ocmlog.DefineSubRealm("blob access for wget", "blobaccess/wget")
