// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package elements

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/errors"
)

type ReferenceOption interface {
	ApplyToReference(reference *compdesc.ComponentReference) error
}

func Reference(name, comp, vers string, opts ...ReferenceOption) (*compdesc.ComponentReference, error) {
	m := compdesc.NewComponentReference(name, comp, vers, nil)
	list := errors.ErrList()
	for _, o := range opts {
		if o != nil {
			list.Add(o.ApplyToReference(m))
		}
	}
	return m, list.Result()
}
