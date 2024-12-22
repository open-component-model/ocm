package builder

import (
	"ocm.software/ocm/api/ocm/refhints"
)

func (b *Builder) AccessHint(hint string) {
	b.expect(b.blobhint, T_OCMACCESS)
	if b.ocm_acc != nil && *b.ocm_acc != nil {
		b.fail("access already set")
	}
	hints := refhints.ParseHints(hint, true)
	*b.blobhint = hints
}

////////////////////////////////////////////////////////////////////////////////

const T_OCMARTIFACT = "ocm artifact"

func (b *Builder) ElementHint(hint string) {
	b.expect(b.ocm_metahints, T_OCMARTIFACT)
	hints := refhints.ParseHints(hint)
	b.ocm_metahints.SetReferenceHints(hints)
}
