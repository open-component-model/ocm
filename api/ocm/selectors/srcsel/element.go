package srcsel

import (
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/selectors"
	"ocm.software/ocm/api/ocm/selectors/labelsel"
)

// Identity selectors

func IdentityByKeyPairs(extras ...string) Selector {
	return selectors.IdentityByKeyPairs(extras...)
}

func Identity(id v1.Identity) Selector {
	return selectors.Identity(id)
}

func ExtraIdentity(id v1.Identity) Selector {
	return selectors.ExtraIdentity(id)
}

func ExtraIdentityByKeyPairs(extra ...string) Selector {
	return selectors.ExtraIdentityByKeyPairs(extra...)
}

func Partialdentity(id v1.Identity) Selector {
	return selectors.PartialIdentity(id)
}

func PartialIdentityByKeyPairs(extra ...string) Selector {
	return selectors.PartialIdentityByKeyPairs(extra...)
}

func Name(n string) Selector {
	return selectors.Name(n)
}

func Version(n string) Selector {
	return selectors.Version(n)
}

func VersionConstraint(expr string) Selector {
	return selectors.VersionConstraint(expr)
}

// Label selectors

func Label(sel ...selectors.LabelSelector) Selector {
	return selectors.Label(sel...)
}

func LabelName(n string) Selector {
	return labelsel.Name(n)
}

func LabelVersion(n string) Selector {
	return labelsel.Version(n)
}

func LabelValue(v interface{}) Selector {
	return labelsel.Value(v)
}
