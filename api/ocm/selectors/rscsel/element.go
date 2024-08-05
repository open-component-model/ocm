package rscsel

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
