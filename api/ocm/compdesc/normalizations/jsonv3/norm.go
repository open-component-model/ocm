// Package jsonv3 provides a normalization which is completely based on the
// abstract (internal) version of the component descriptor and is therefore
// agnostic of the final serialization format. Signatures using this algorithm
// can be transferred among different schema versions, as long as is able to
// handle the complete information using for the normalization.
// jsonv2 is the predecessor of this version but had internal defaulting logic
// that is no longer included as part of this normalization. Thus v3 should be preferred over v2.
// Note that between v2 and v3 differences can occur mainly if the "extra identity" field is not unique,
// in which case the v2 normalization opinionated on how to differentiate these items. This no longer
// happens in v3, meaning the component descriptor is normalized as is.
package jsonv3

import (
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/compdesc/normalizations/jsonv2"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/norm/jcs"
)

const Algorithm = compdesc.JsonNormalisationV3

func init() {
	compdesc.Normalizations.Register(Algorithm, normalization{})
}

type normalization struct{}

func (m normalization) Normalize(cd *compdesc.ComponentDescriptor) ([]byte, error) {
	data, err := signing.Normalize(jcs.Type, cd, jsonv2.CDExcludes)
	return data, err
}
