package genericblob

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/elements/artifactblob/api"
)

type (
	Options = api.Options
	Option  = api.Option
)

func WithHint(h string) Option {
	return api.WithHint(h)
}

func WithGlobalAccess(a cpi.AccessSpec) Option {
	return api.WithGlobalAccess(a)
}
