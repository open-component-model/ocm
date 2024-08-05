package artifacthdlr

import (
	"strings"

	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/cmds/ocm/common/processing"
)

func Compare(a, b interface{}) int {
	aa := a.(*Object)
	ab := b.(*Object)

	c := strings.Compare(aa.Spec.UniformRepositorySpec.String(), ab.Spec.UniformRepositorySpec.String())
	if c == 0 {
		return common.CompareHistoryElement(aa, ab)
	}
	return c
}

// Sort is a processing chain sorting original objects provided by type handler.
var Sort = processing.Sort(Compare)
