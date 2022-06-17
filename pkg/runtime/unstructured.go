// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/modern-go/reflect2"
)

const ATTR_TYPE = "type"

// ATTENTION: UnstructuredTypedObject CANNOT be be used as anonymous
// field together with the default struct marshalling with the
// great json marshallers.
// Anonymous inline struct fields are always marshaled by the default struct
// marshales in a depth first manner without observing the Marshal interface!!!!
//
// Therefore all structs in this module deriving from UnstructuedTypedObject
// are explicitly implementing the marshal/unmarshal interface.
//
// Side Fact: Marshaling a map[interface{}] filled by unmarshaling a marshaled
// object with anonymous fields is not stable, because the inline fields
// are sorted depth firt for marshalling, while maps key are marshaled
// completely in order.
// Therefore we do not store the raw bytes but marshal them always from
// the UnstructuedMap.

// Unstructured is the interface to represent generic object data for
// types handled by schemes.
type Unstructured interface {
	TypeGetter
	GetRaw() ([]byte, error)
}

type Object interface{}

type JSONMarhaler interface {
	MarshalJSON() ([]byte, error)
}

// UnstructuredMap is a generic data map
type UnstructuredMap map[string]interface{}

// FlatMerge just joins the direct attribute set
func (m UnstructuredMap) FlatMerge(o UnstructuredMap) UnstructuredMap {
	for k, v := range o {
		m[k] = v
	}
	return m
}

// UnstructuredTypesEqual compares two unstructured object.
func UnstructuredTypesEqual(a, b *UnstructuredTypedObject) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.GetType() != b.GetType() {
		return false
	}
	rawA, err := a.GetRaw()
	if err != nil {
		return false
	}
	rawB, err := b.GetRaw()
	if err != nil {
		return false
	}
	return bytes.Equal(rawA, rawB)
}

// TypedObjectEqual compares two typed objects using the unstructured type.
func TypedObjectEqual(a, b TypedObject) bool {
	if a.GetType() != b.GetType() {
		return false
	}
	uA, err := ToUnstructuredTypedObject(a)
	if err != nil {
		return false
	}
	uB, err := ToUnstructuredTypedObject(b)
	if err != nil {
		return false
	}
	return UnstructuredTypesEqual(uA, uB)
}

// NewEmptyUnstructured creates a new typed object without additional data.
func NewEmptyUnstructured(ttype string) *UnstructuredTypedObject {
	return NewUnstructuredType(ttype, nil)
}

// NewEmptyUnstructuredVersioned creates a new typed object without additional data.
func NewEmptyUnstructuredVersioned(ttype string) *UnstructuredVersionedTypedObject {
	return &UnstructuredVersionedTypedObject{*NewUnstructuredType(ttype, nil)}
}

// NewUnstructuredType creates a new unstructured typed object.
func NewUnstructuredType(ttype string, data UnstructuredMap) *UnstructuredTypedObject {
	unstr := &UnstructuredTypedObject{}
	unstr.Object = data
	unstr.SetType(ttype)
	return unstr
}

// UnstructuredConverter converts the actual object to an UnstructuredTypedObject
type UnstructuredConverter interface {
	ToUnstructured() (*UnstructuredTypedObject, error)
}

// UnstructuredTypedObject describes a generic typed object.
type UnstructuredTypedObject struct {
	ObjectType `json:",inline"`
	Object     UnstructuredMap `json:"-"`
}

func (s *UnstructuredTypedObject) ToUnstructured() (*UnstructuredTypedObject, error) {
	return s, nil
}

func (u *UnstructuredTypedObject) SetType(ttype string) {
	u.ObjectType.SetType(ttype)
	if u.Object == nil {
		u.Object = UnstructuredMap{}
	}
	u.Object[ATTR_TYPE] = ttype
}

// DeepCopyInto is deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (u *UnstructuredTypedObject) DeepCopyInto(out *UnstructuredTypedObject) {
	*out = *u
	raw, _ := json.Marshal(u.Object)
	_ = out.setRaw(raw)
}

// DeepCopy is deepcopy function, copying the receiver, creating a new UnstructuredTypedObject.
func (u *UnstructuredTypedObject) DeepCopy() *UnstructuredTypedObject {
	if u == nil {
		return nil
	}
	out := new(UnstructuredTypedObject)
	u.DeepCopyInto(out)
	return out
}

func (u UnstructuredTypedObject) GetRaw() ([]byte, error) {
	return json.Marshal(u.Object)
}

func (u *UnstructuredTypedObject) setRaw(data []byte) error {
	obj := UnstructuredMap{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	u.Object = obj
	return nil
}

// Evaluate converts a unstructured object into a typed object.
func (u *UnstructuredTypedObject) Evaluate(types Scheme) (TypedObject, error) {
	data, err := u.GetRaw()
	if err != nil {
		return nil, fmt.Errorf("unable to get data from unstructured object: %w", err)
	}
	var decoder TypedObjectDecoder
	if types != nil {
		decoder = types.GetDecoder(u.GetType())
	}
	if decoder == nil {
		return u, nil
	}

	if obj, err := decoder.Decode(data, DefaultJSONEncoding); err != nil {
		return nil, fmt.Errorf("unable to decode object %q: %w", u.GetType(), err)
	} else {
		return obj, nil
	}
}

// UnmarshalJSON implements a custom json unmarshal method for a unstructured typed object.
func (u *UnstructuredTypedObject) UnmarshalJSON(data []byte) error {
	//fmt.Printf("unmarshal raw: %s\n", string(data))
	typedObj := ObjectType{}
	if err := json.Unmarshal(data, &typedObj); err != nil {
		return err
	}

	obj := UnstructuredTypedObject{
		ObjectType: typedObj,
	}
	if err := obj.setRaw(data); err != nil {
		return err
	}
	*u = obj
	return nil
}

// MarshalJSON implements a custom json unmarshal method for a unstructured type.
func (u UnstructuredTypedObject) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(u.Object)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (_ UnstructuredTypedObject) OpenAPISchemaType() []string { return []string{"object"} }
func (_ UnstructuredTypedObject) OpenAPISchemaFormat() string { return "" }

////////////////////////////////////////////////////////////////////////////////
// Utils
////////////////////////////////////////////////////////////////////////////////

// ToUnstructuredTypedObject converts a typed object to a unstructured object.
func ToUnstructuredTypedObject(obj TypedObject) (*UnstructuredTypedObject, error) {
	if reflect2.IsNil(obj) {
		return nil, nil
	}
	if un, ok := obj.(UnstructuredConverter); ok {
		return un.ToUnstructured()
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	uObj := &UnstructuredTypedObject{}
	if err := json.Unmarshal(data, uObj); err != nil {
		return nil, err
	}
	return uObj, nil
}

// ToUnstructuredObject converts any object into a structure map.
func ToUnstructuredObject(obj interface{}) (UnstructuredMap, error) {
	if reflect2.IsNil(obj) {
		return nil, nil
	}
	if un, ok := obj.(map[string]interface{}); ok {
		return UnstructuredMap(un), nil
	}

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	uObj := UnstructuredMap{}
	if err := json.Unmarshal(data, &uObj); err != nil {
		return nil, err
	}
	return uObj, nil
}

type UnstructuredTypedObjectList []*UnstructuredTypedObject

func (l UnstructuredTypedObjectList) Copy() UnstructuredTypedObjectList {
	n := make(UnstructuredTypedObjectList, len(l), len(l))
	for i, u := range l {
		copy := *u
		n[i] = &copy
	}
	return n
}
