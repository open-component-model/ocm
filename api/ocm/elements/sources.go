package elements

import (
	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/ocm/compdesc"
)

type SourceMetaOption interface {
	ApplyToSourceMeta(*compdesc.SourceMeta) error
}

func SourceMeta(name, typ string, opts ...SourceMetaOption) (*compdesc.SourceMeta, error) {
	m := compdesc.NewSourceMeta(name, typ)
	list := errors.ErrList()
	for _, o := range opts {
		if o != nil {
			list.Add(o.ApplyToSourceMeta(m))
		}
	}
	return m, list.Result()
}
