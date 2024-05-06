package testhelper

import (
	"github.com/mandelsoft/logging"

	ocmlog "github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/refmgmt"
)

func EnableRefMgmtLog() {
	ocmlog.Context().AddRule(logging.NewConditionRule(logging.TraceLevel, refmgmt.ALLOC_REALM))
}
