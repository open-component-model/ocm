package builder

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
)

const T_OCMACCESS = "access"

////////////////////////////////////////////////////////////////////////////////

func (b *Builder) Access(acc compdesc.AccessSpec) {
	b.expect(b.ocm_acc, T_OCMACCESS)
	if b.blob != nil && *b.blob != nil {
		b.fail("access already set")
	}
	if b.hint != nil && *b.hint != "" {
		b.fail("hint requires blob", 1)
	}

	*(b.ocm_acc) = acc
}
