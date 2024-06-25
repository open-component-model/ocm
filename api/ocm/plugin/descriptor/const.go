package descriptor

import (
	"github.com/open-component-model/ocm/api/datacontext/action"
	"github.com/open-component-model/ocm/api/utils/errkind"
	ocmlog "github.com/open-component-model/ocm/api/utils/logging"
)

const (
	KIND_PLUGIN       = "plugin"
	KIND_DOWNLOADER   = "downloader"
	KIND_UPLOADER     = "uploader"
	KIND_ACCESSMETHOD = errkind.KIND_ACCESSMETHOD
	KIND_ACTION       = action.KIND_ACTION
	KIND_VALUESET     = "value set"
	KIND_PURPOSE      = "purposet"
)

var REALM = ocmlog.DefineSubRealm("OCM plugin handling", "plugins")
