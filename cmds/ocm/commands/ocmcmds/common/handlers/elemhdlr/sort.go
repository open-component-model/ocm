package elemhdlr

import (
	"ocm.software/ocm/cmds/ocm/common/processing"
)

func Compare(a, b interface{}) int {
	aa := a.(*Object)
	ab := b.(*Object)

	return aa.Compare(ab)
}

// Sort is a processing chain sorting original objects provided by type handler.
var Sort = processing.Sort(Compare)
