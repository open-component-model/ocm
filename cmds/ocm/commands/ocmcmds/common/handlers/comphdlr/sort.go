package comphdlr

import (
	"strings"

	"ocm.software/ocm/api/utils/semverutils"
	"ocm.software/ocm/cmds/ocm/common/processing"
)

func Compare(a, b interface{}) int {
	aa := a.(*Object)
	ab := b.(*Object)

	c := strings.Compare(aa.ComponentVersion.GetName(), ab.ComponentVersion.GetName())
	if c != 0 {
		return c
	}
	return semverutils.Compare(aa.ComponentVersion.GetVersion(), ab.ComponentVersion.GetVersion())
}

// Sort is a processing chain sorting original objects provided by type handler.
var Sort = processing.Sort(Compare)
