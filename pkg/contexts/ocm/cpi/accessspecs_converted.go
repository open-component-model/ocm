// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/internal"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type AccessSpecConverter interface {
	ConvertFrom(object internal.AccessSpec) (runtime.TypedObject, error)
	ConvertTo(object interface{}) (internal.AccessSpec, error)
}

type AccessSpecVersion interface {
	AccessSpecConverter
	runtime.TypedObjectDecoder
	CreateData() interface{}
}

type accessSpecVersion struct {
	*runtime.ConvertingDecoder
	AccessSpecConverter
}

type typedObjectConverter struct {
	converter AccessSpecConverter
}

func (c *typedObjectConverter) ConvertTo(object interface{}) (runtime.TypedObject, error) {
	return c.converter.ConvertTo(object)
}

func NewAccessSpecVersion(proto runtime.TypedObject, converter AccessSpecConverter) AccessSpecVersion {
	return &accessSpecVersion{
		ConvertingDecoder:   runtime.MustNewConvertingDecoder(proto, &typedObjectConverter{converter}),
		AccessSpecConverter: converter,
	}
}

////////////////////////////////////////////////////////////////////////////////

type ConvertedAccessType struct {
	AccessSpecVersion
	accessType
}

var (
	_ AccessSpecVersion          = &ConvertedAccessType{}
	_ runtime.TypedObjectEncoder = &ConvertedAccessType{}
)

func NewConvertedAccessSpecType(name string, v AccessSpecVersion, opts ...AccessSpecTypeOption) *ConvertedAccessType {
	t := &ConvertedAccessType{
		accessType: accessType{
			ObjectVersionedType: runtime.NewVersionedObjectType(name),
			TypedObjectDecoder:  v,
		},
		AccessSpecVersion: v,
	}
	for _, o := range opts {
		o.ApplyToAccessSpecOptionTarget(accessTypeTarget{&t.accessType})
	}
	return t
}

func (t *ConvertedAccessType) Encode(obj runtime.TypedObject, m runtime.Marshaler) ([]byte, error) {
	c, err := t.ConvertFrom(obj.(AccessSpec))
	if err != nil {
		return nil, err
	}
	return m.Marshal(c)
}
