package elemhdlr

import (
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
)

func Compare(a, b interface{}) int {
	aa := a.(*Object)
	ab := b.(*Object)

	return aa.Compare(ab)
}

// Sort is a processing chain sorting original objects provided by type handler.
var Sort = processing.Sort(Compare)
