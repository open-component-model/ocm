package genericblob

import (
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactblob/api"
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
