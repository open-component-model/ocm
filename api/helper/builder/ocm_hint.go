package builder

import (
	"fmt"

	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/modern-go/reflect2"

	"ocm.software/ocm/api/ocm/refhints"
)

// AccessHint provides reference hints for
// an access (either an access method or blob data).
func (b *Builder) AccessHint(hint interface{}) {
	b.expect(b.blobhint, T_OCMACCESS)
	if b.ocm_acc != nil && *b.ocm_acc != nil {
		b.fail("access already set")
	}
	hints := b.hints(hint, true)
	*b.blobhint = hints
}

////////////////////////////////////////////////////////////////////////////////

const T_OCMARTIFACT = "ocm artifact"

// ArtifactHint provides reference hints stored
// togetjer with the metadata of an artifact
// (source or resource).
func (b *Builder) ArtifactHint(hint interface{}) {
	b.expect(b.ocm_metahints, T_OCMARTIFACT)
	hints := b.hints(hint)
	b.ocm_metahints.SetReferenceHints(hints)
}

func (b *Builder) hints(spec interface{}, implicit ...bool) refhints.ReferenceHints {
	if reflect2.IsNil(spec) {
		return nil
	}
	switch t := spec.(type) {
	case string:
		return refhints.ParseHints(t, implicit...)
	case refhints.ReferenceHints:
		return t
	case refhints.DefaultReferenceHints:
		return sliceutils.Convert[refhints.ReferenceHint](t)
	case refhints.ReferenceHint:
		return refhints.ReferenceHints{t}
	default:
		b.fail(fmt.Sprintf("unknown hint specification type (%T)", spec), 1)
	}
	return nil
}
