package descriptor

import (
	"ocm.software/ocm/api/datacontext/action"
	"ocm.software/ocm/api/utils/errkind"
	ocmlog "ocm.software/ocm/api/utils/logging"
)

const (
	KIND_PLUGIN          = "plugin"
	KIND_DOWNLOADER      = "downloader"
	KIND_UPLOADER        = "uploader"
	KIND_ACCESSMETHOD    = errkind.KIND_ACCESSMETHOD
	KIND_ACTION          = action.KIND_ACTION
	KIND_TRANSFERHANDLER = "transferhandler"
	KIND_VALUESET        = "value set"
	KIND_PURPOSE         = "purposet"
)

var REALM = ocmlog.DefineSubRealm("OCM plugin handling", "plugins")
