// Package jsonv1 provides a normalization which uses schema specific
// normalizations.
// It creates the requested schema for the component descriptor
// and just forwards the normalization to this version.
package jsonv1

import (
	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/compdesc/normalizations/legacy"
	"ocm.software/ocm/api/utils/errkind"
)

// Deprecated: use compdesc.JsonNormalisationV3 instead
const Algorithm = compdesc.JsonNormalisationV1

func init() {
	compdesc.Normalizations.Register(Algorithm, normalization{})
}

type normalization struct{}

func (m normalization) Normalize(cd *compdesc.ComponentDescriptor) ([]byte, error) {
	legacy.DefaultingOfVersionIntoExtraIdentityForDescriptor(cd)
	cv := compdesc.DefaultSchemes[cd.SchemaVersion()]
	if cv == nil {
		return nil, errors.ErrNotSupported(errkind.KIND_SCHEMAVERSION, cd.SchemaVersion())
	}
	v, err := cv.ConvertFrom(cd)
	if err != nil {
		return nil, err
	}
	return v.Normalize(Algorithm)
}
