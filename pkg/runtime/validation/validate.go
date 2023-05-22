// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime/encoding"
	"github.com/open-component-model/ocm/pkg/utils"
)

// Validater is an object interface which may be implemented
// to validate its property settings.
type Validater interface {
	Validate() error
}

// UnmarshalWithValidation unmarshalls an object serialization by a given
// unmarshaler into an object according to the type parameter (it must not
// be a pointer). Optionally a set of validation rules may be given used
// to complain about additional (not used by the object) properties.
func UnmarshalWithValidation[T any](data []byte, unmarshaler encoding.Unmarshaler, additional ...AdditionalProperties) (*T, error) {
	var obj T
	err := UnmarshalProtoWithValidation(data, &obj, unmarshaler, additional...)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func UnmarshalProtoWithValidation(data []byte, proto interface{}, unmarshaler encoding.Unmarshaler, additional ...AdditionalProperties) error {
	err := unmarshaler.Unmarshal(data, proto)
	if err != nil {
		return err
	}

	var fields interface{}
	err = unmarshaler.Unmarshal(data, &fields)
	if err != nil {
		return errors.Wrapf(err, "cannot unmarshal data")
	}

	data, err = encoding.DefaultJSONEncoding.Marshal(proto)
	if err != nil {
		return errors.Wrapf(err, "cannot marshal object")
	}

	var found interface{}
	err = encoding.DefaultJSONEncoding.Unmarshal(data, &found)
	if err != nil {
		return errors.Wrapf(err, "cannot unmarshal accepted data")
	}
	err = checkFields(nil, fields, found, Next(true, Composition(additional))).ToAggregate()
	if err != nil {
		return err
	}
	if v, ok := proto.(Validater); ok {
		return v.Validate()
	}
	return nil
}

func checkFields(fldPath *field.Path, data, accepted interface{}, elem ElementInfo) field.ErrorList {
	if data == nil {
		return nil
	}

	allErrs := field.ErrorList{}

	switch a := accepted.(type) {
	case map[string]interface{}:
		if d, ok := data.(map[string]interface{}); ok {
			return checkMap(fldPath, d, a, elem.Next())
		}
		return append(allErrs, field.TypeInvalid(fldPath, field.OmitValueType{}, "field must be a map/struct"))
	case []interface{}:
		if d, ok := data.([]interface{}); ok {
			return checkList(fldPath, d, a, elem.Next())
		}
		return append(allErrs, field.TypeInvalid(fldPath, field.OmitValueType{}, "field must be a list"))
	default:
		if accepted == nil {
			if elem.IsEnabled() {
				return allErrs
			}
			return append(allErrs, field.Forbidden(fldPath, "unknown field"))
		}
		if !reflect.DeepEqual(accepted, data) {
			return append(allErrs, field.Invalid(fldPath, data, "value is invalid"))
		}
	}
	return nil
}

func checkMap(fldPath *field.Path, data, accepted map[string]interface{}, additional ...AdditionalProperties) field.ErrorList {
	allErrs := field.ErrorList{}
	for _, k := range utils.StringMapKeys(data) {
		f := fldPath.Child(k)
		allErrs = append(allErrs, checkFields(f, data[k], accepted[k], Composition(additional).ForMapField(k, f))...)
	}
	return allErrs
}

func checkList(fldPath *field.Path, data, accepted []interface{}, additional ...AdditionalProperties) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(data) != len(accepted) {
		return append(allErrs, field.InternalError(fldPath, fmt.Errorf("non matching list size")))
	}
	for i, v := range data {
		f := fldPath.Index(i)
		allErrs = append(allErrs, checkFields(f, v, accepted[i], Composition(additional).ForListEntry(i, v))...)
	}
	return allErrs
}

type ElementInfo interface {
	IsEnabled() bool
	Next() AdditionalProperties
}

// AdditionalProperties is the interface for rules
// deciding to accept unknown properties or not.
type AdditionalProperties interface {
	ForMapField(name string, elem interface{}) ElementInfo
	ForListEntry(index int, elem interface{}) ElementInfo
}

type elementInfo struct {
	flag bool
	next AdditionalProperties
}

func (e *elementInfo) IsEnabled() bool {
	return e.flag
}

func (e *elementInfo) Next() AdditionalProperties {
	return e.next
}

func Next(flag bool, next AdditionalProperties) ElementInfo {
	return &elementInfo{flag, next}
}

func NoAdditionalProperties() AdditionalProperties {
	a := &additionalProperties{}
	a.elem = Next(false, a)
	return a
}

func AdditionalRootProperties() AdditionalProperties {
	return &additionalProperties{Next(true, NoAdditionalProperties())}
}

type additionalProperties struct {
	elem ElementInfo
}

func (n *additionalProperties) ForMapField(name string, elem interface{}) ElementInfo {
	return n.elem
}

func (n *additionalProperties) ForListEntry(index int, elem interface{}) ElementInfo {
	return n.elem
}

type Composition []AdditionalProperties

func (c Composition) ForMapField(name string, elem interface{}) ElementInfo {
	for _, a := range c {
		e := a.ForMapField(name, elem)
		if e != nil {
			return e
		}
	}
	return Next(false, NoAdditionalProperties())
}

func (c Composition) ForListEntry(index int, elem interface{}) ElementInfo {
	for _, a := range c {
		e := a.ForListEntry(index, elem)
		if e != nil {
			return e
		}
	}
	return Next(false, NoAdditionalProperties())
}

type additionalMapField struct {
	name     string
	optional bool
	nested   AdditionalProperties
}

func (a *additionalMapField) ForMapField(name string, elem interface{}) ElementInfo {
	if name == a.name || a.name == "" {
		return Next(a.optional, a.nested)
	}
	return nil
}

func (a additionalMapField) ForListEntry(index int, elem interface{}) ElementInfo {
	return nil
}

func AdditionalMapField(name string, nested ...AdditionalProperties) AdditionalProperties {
	return &additionalMapField{name, true, Composition(nested)}
}

func MapField(name string, nested ...AdditionalProperties) AdditionalProperties {
	return &additionalMapField{name, false, Composition(nested)}
}

type additionalListField struct {
	index    int
	optional bool
	nested   AdditionalProperties
}

func (a *additionalListField) ForMapField(name string, elem interface{}) ElementInfo {
	return nil
}

func (a *additionalListField) ForListEntry(index int, elem interface{}) ElementInfo {
	if a.index == index || a.index < 0 {
		return Next(a.optional, a.nested)
	}
	return nil
}

func ListField(index int, nested ...AdditionalProperties) AdditionalProperties {
	return &additionalListField{index, false, Composition(nested)}
}
