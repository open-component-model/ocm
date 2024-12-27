package builder

import (
	"ocm.software/ocm/api/ocm/refhints"
)

// AccessHint provides reference hints for
// an access (either an access method or blob data).
func (b *Builder) AccessHint(hint interface{}) {
	b.expect(b.blobhint, T_OCMACCESS)
	if b.ocm_acc != nil && *b.ocm_acc != nil {
		b.fail("access already set")
	}
	hints, err := refhints.HintsFor(hint, true)
	if err != nil {
		b.fail(err.Error())
	}
	*b.blobhint = hints
}

////////////////////////////////////////////////////////////////////////////////

const T_OCMARTIFACT = "ocm artifact"

// ArtifactHint provides reference hints stored
// togetjer with the metadata of an artifact
// (source or resource).
func (b *Builder) ArtifactHint(hint interface{}) {
	b.expect(b.ocm_metahints, T_OCMARTIFACT)
	hints, err := refhints.HintsFor(hint)
	if err != nil {
		b.fail(err.Error())
	}
	b.ocm_metahints.SetReferenceHints(hints)
}
