package rscsel

import (
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/selectors"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/selectors/labelsel"
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
