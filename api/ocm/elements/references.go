package elements

import (
	"github.com/mandelsoft/goutils/errors"

	"ocm.software/ocm/api/ocm/compdesc"
)

type ReferenceOption interface {
	ApplyToReference(reference *compdesc.Reference) error
}

func Reference(name, comp, vers string, opts ...ReferenceOption) (*compdesc.Reference, error) {
	m := compdesc.NewComponentReference(name, comp, vers, nil)
	list := errors.ErrList()
	for _, o := range opts {
		if o != nil {
			list.Add(o.ApplyToReference(m))
		}
	}
	return m, list.Result()
}
