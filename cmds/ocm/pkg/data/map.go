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

package data

type Map interface {
	Has(interface{}) bool
	Get(interface{}) interface{}
	Set(interface{}, interface{})
	Iterator() Iterator
	Keys() Iterator
	Values() Iterator
	Size() int
}

type MapEntry struct {
	Key   interface{}
	Value interface{}
}

type _Map map[interface{}]interface{}

func NewMap() Map {
	return _Map{}
}

func (this _Map) Has(key interface{}) bool {
	_, ok := this[key]
	return ok
}

func (this _Map) Get(key interface{}) interface{} {
	v, ok := this[key]
	if ok {
		return v
	}
	return nil
}

func (this _Map) Set(key interface{}, value interface{}) {
	this[key] = value
}

func (this _Map) Size() int {
	return len(this)
}

func (this _Map) Keys() Iterator {
	return &_MapIterator{this, newMapKeyIterator(this)}
}

func (this _Map) Iterator() Iterator {
	return &_EntryIterator{this, newMapKeyIterator(this)}
}

func (this _Map) Values() Iterator {
	return NewMappedIterator(this.Iterator(), func(e interface{}) interface{} {
		return e.(MapEntry).Value
	})
}

type _MapIterator struct {
	data _Map
	Iterator
}

func newMapKeyIterator(m _Map) ResettableIterator {
	keys := make([]interface{}, m.Size())
	i := 0
	for k, _ := range m {
		keys[i] = k
		i++
	}
	return NewSliceIterator(keys)
}

func (this *_MapIterator) Reset() {
	this.Iterator = newMapKeyIterator(this.data)
}

type _EntryIterator _MapIterator

func (this *_EntryIterator) Next() interface{} {
	if this.HasNext() {
		k := this.Iterator.Next()
		v, _ := this.data[k]
		return MapEntry{k, v}
	}
	return nil
}
