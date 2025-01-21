package legacy

import (
	"fmt"

	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/utils/logging"
)

var (
	REALM  = logging.DefineSubRealm("component descriptor legacy normalization defaulting", "compdesc", "normalizations", "legacy")
	Logger = logging.DynamicLogger(REALM)
)

// DefaultingOfVersionIntoExtraIdentityForDescriptor normalizes the extra identity of the resources.
// It sets the version of the resource, reference or source as extra identity field if the combination of name+extra identity
// is the same for multiple items. However, the last item in the list will not be updated as it is unique wihout this.
//
// TODO: To be removed once v1 + v2 are removed.
//
// Deprecated: This is a legacy normalization and should only be used as part of JsonNormalisationV1 and JsonNormalisationV2
// for backwards compatibility of normalization (for example used for signatures). It was needed because the original
// defaulting was made part of the normalization by accident and is now no longer included by default due to
// https://github.com/open-component-model/ocm/pull/1026
func DefaultingOfVersionIntoExtraIdentityForDescriptor(cd *compdesc.ComponentDescriptor) {
	resources := make([]IdentityDefaultable, len(cd.Resources))
	for i := range cd.Resources {
		resources[i] = &cd.Resources[i]
	}

	DefaultingOfVersionIntoExtraIdentity(resources)
}

type IdentityDefaultable interface {
	GetExtraIdentity() metav1.Identity
	SetExtraIdentity(metav1.Identity)
	GetName() string
	GetVersion() string
}

func DefaultingOfVersionIntoExtraIdentity(meta []IdentityDefaultable) {
	for i := range meta {
		for j := range meta {
			// don't match with itself and only match with the same name
			if meta[j].GetName() != meta[i].GetName() || i == j {
				continue
			}

			eid := meta[i].GetExtraIdentity()
			// if the extra identity is not the same, then there is not a clash
			if !meta[j].GetExtraIdentity().Equals(eid) {
				continue
			}

			eid.Set(compdesc.SystemIdentityVersion, meta[i].GetVersion())
			meta[i].SetExtraIdentity(eid)

			Logger.Warn(fmt.Sprintf("resource identity duplication was normalized for backwards compatibility, "+
				"to avoid this either specify a unique extra identity per item or switch to %s", compdesc.JsonNormalisationV3),
				"name", meta[i].GetName(), "index", i, "extra identity", meta[i].GetExtraIdentity())
			break
		}
	}
}
