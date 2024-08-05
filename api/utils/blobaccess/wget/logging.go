package wget

import (
	ocmlog "ocm.software/ocm/api/utils/logging"
)

var REALM = ocmlog.DefineSubRealm("blob access for wget", "blobaccess/wget")
