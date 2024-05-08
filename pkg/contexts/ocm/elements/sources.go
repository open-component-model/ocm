package elements

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/errors"
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
