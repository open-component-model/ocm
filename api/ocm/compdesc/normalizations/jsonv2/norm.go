// Package jsonv2 provides a normalization which is completely based on the
// abstract (internal) version of the component descriptor and is therefore
// agnostic of the final serialization format. Signatures using this algorithm
// can be transferred among different schema versions, as long as is able to
// handle the complete information using for the normalization.
// Older format might omit some info, therefore the signatures cannot be
// validated for such representations, if the original component descriptor
// has used such parts.
package jsonv2

import (
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/compdesc/normalizations/legacy"
	"ocm.software/ocm/api/ocm/compdesc/normalizations/rules"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/norm/jcs"
)

// Deprecated: use compdesc.JsonNormalisationV3 instead
const Algorithm = compdesc.JsonNormalisationV2

func init() {
	compdesc.Normalizations.Register(Algorithm, normalization{})
}

type normalization struct{}

func (m normalization) Normalize(cd *compdesc.ComponentDescriptor) ([]byte, error) {
	legacy.DefaultingOfVersionIntoExtraIdentity(cd)
	data, err := signing.Normalize(jcs.Type, cd, CDExcludes)
	return data, err
}

// CDExcludes describes the fields relevant for Signing
// ATTENTION: if changed, please adapt the Equivalent Functions
// in the generic part, accordingly.
var CDExcludes = signing.MapExcludes{
	"meta": nil,
	"component": signing.MapExcludes{
		"repositoryContexts": nil,
		"provider": signing.MapExcludes{
			"labels": rules.LabelExcludes,
		},
		"labels": rules.LabelExcludes,
		"resources": signing.DynamicArrayExcludes{
			ValueMapper: rules.MapResourcesWithNoneAccess,
			Continue: signing.MapExcludes{
				"access":  nil,
				"srcRefs": nil,
				"labels":  rules.LabelExcludes,
			},
		},
		"sources": signing.ArrayExcludes{
			Continue: signing.MapExcludes{
				"access": nil,
				"labels": rules.LabelExcludes,
			},
		},
		"references": signing.ArrayExcludes{
			signing.MapExcludes{
				"labels": rules.LabelExcludes,
			},
		},
	},
	"signatures":    nil,
	"nestedDigests": nil,
}
