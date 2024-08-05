package v3alpha1

import (
	"fmt"

	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/compdesc/normalizations/rules"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/norm/entry"
)

// CDExcludes describes the fields relevant for Signing
// ATTENTION: if changed, please adapt the HashEqual Functions
// in the generic part, accordingly.
var CDExcludes = signing.MapExcludes{
	"repositoryContexts": nil,
	"metadata": signing.MapExcludes{
		"labels": rules.LabelExcludes,
	},
	"spec": signing.MapExcludes{
		"provider": signing.MapExcludes{
			"labels": rules.LabelExcludes,
		},
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

func (cd *ComponentDescriptor) Normalize(normAlgo string) ([]byte, error) {
	if normAlgo != compdesc.JsonNormalisationV1 {
		return nil, fmt.Errorf("unsupported cd normalization %q", normAlgo)
	}
	data, err := signing.Normalize(entry.Type, cd, CDExcludes)
	return data, err
}
