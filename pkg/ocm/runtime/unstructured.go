// Copyright 2022 Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/modern-go/reflect2"
)

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

// NewUnstructuredType creates a new unstructured typed object.
func NewUnstructuredType(ttype string, data map[string]interface{}) *UnstructuredTypedObject {
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
// +k8s:openapi-gen=true
type UnstructuredTypedObject struct {
	ObjectType `json:",inline"`
	Raw        []byte                 `json:"-"`
	Object     map[string]interface{} `json:"-"`
}

func (s *UnstructuredTypedObject) ToUnstructured() (*UnstructuredTypedObject, error) {
	return s, nil
}

func (u *UnstructuredTypedObject) SetType(ttype string) {
	u.ObjectType.SetType(ttype)
	if u.Object == nil {
		u.Object = make(map[string]interface{})
	}
	u.Object["type"] = ttype
}

// DeepCopyInto is deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (u *UnstructuredTypedObject) DeepCopyInto(out *UnstructuredTypedObject) {
	*out = *u
	raw := make([]byte, len(u.Raw))
	copy(raw, u.Raw)
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
	data, err := json.Marshal(u.Object)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(data, u.Raw) {
		u.Raw = data
	}
	return u.Raw, nil
}

func (u *UnstructuredTypedObject) setRaw(data []byte) error {
	obj := map[string]interface{}{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	u.Raw = data
	u.Object = obj
	return nil
}

// Evaluate converts a unstructured object into a typed object.
func (u *UnstructuredTypedObject) Evaluate(types KnownTypes) (TypedObject, error) {
	data, err := u.GetRaw()
	if err != nil {
		return nil, fmt.Errorf("unable to get data from unstructured object: %w", err)
	}
	var codec TypedObjectCodec
	if types != nil {
		codec = types.GetCodec(u.GetType())
	}
	if codec == nil {
		return u, nil
	}

	if obj, err := codec.Decode(data); err != nil {
		return nil, fmt.Errorf("unable to decode object %q: %w", u.GetType(), err)
	} else {
		return obj, nil
	}
}

// UnmarshalJSON implements a custom json unmarshal method for a unstructured typed object.
func (u *UnstructuredTypedObject) UnmarshalJSON(data []byte) error {
	fmt.Printf("unmarshal raw: %s\n", string(data))
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
func (u *UnstructuredTypedObject) MarshalJSON() ([]byte, error) {
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

type UnstructuredTypedObjectList []*UnstructuredTypedObject

func (l UnstructuredTypedObjectList) Copy() UnstructuredTypedObjectList {
	n := make(UnstructuredTypedObjectList, len(l), len(l))
	for i, u := range l {
		copy := *u
		n[i] = &copy
	}
	return n
}
