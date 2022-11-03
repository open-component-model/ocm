// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type accessType struct {
	cpi.AccessType
	plug plugin.Plugin
}

var _ cpi.AccessType = (*accessType)(nil)

func NewType(name string, p plugin.Plugin, desc, format string) cpi.AccessType {
	if format != "" {
		format = "\n" + format
	}
	t := &accessType{
		AccessType: cpi.NewAccessSpecType(name, &AccessSpec{}, cpi.WithDescription(desc), cpi.WithFormatSpec(format)),
		plug:       p,
	}
	return t
}

func (t *accessType) Decode(data []byte, unmarshaler runtime.Unmarshaler) (runtime.TypedObject, error) {
	spec, err := t.AccessType.Decode(data, unmarshaler)
	if err != nil {
		return nil, err
	}
	spec.(*AccessSpec).handler = NewPluginHandler(t.plug)
	return spec, nil
}
