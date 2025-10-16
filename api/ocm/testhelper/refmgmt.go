package testhelper

import (
	"github.com/mandelsoft/logging"
	ocmlog "ocm.software/ocm/api/utils/logging"
	"ocm.software/ocm/api/utils/refmgmt"
)

func EnableRefMgmtLog() {
	ocmlog.Context().AddRule(logging.NewConditionRule(logging.TraceLevel, refmgmt.ALLOC_REALM))
}
