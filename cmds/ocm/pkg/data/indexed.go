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

type IndexedAccess interface {
	Len() int
	Get(int) interface{}
}

type IndexedIterator struct {
	access  IndexedAccess
	current int
}

var _ ResettableIterator = &IndexedIterator{}

func NewIndexedIterator(a IndexedAccess) *IndexedIterator {
	return (&IndexedIterator{}).new(a)
}

func (this *IndexedIterator) new(a IndexedAccess) *IndexedIterator {
	this.access = a
	this.current = -1
	return this
}

func (this *IndexedIterator) HasNext() bool {
	return this.access.Len() > this.current+1
}

func (this *IndexedIterator) Next() interface{} {
	if this.HasNext() {
		this.current++
		return this.access.Get(this.current)
	}
	return nil
}

func (this *IndexedIterator) Reset() {
	this.current = -1
}

////////////////////////////////////////////////////////////////////////////

type IndexedSliceAccess []interface{}

var _ IndexedAccess = IndexedSliceAccess{}
var _ Iterable = IndexedSliceAccess{}

func (this IndexedSliceAccess) Len() int {
	return len(this)
}

func (this IndexedSliceAccess) Get(i int) interface{} {
	return this[i]
}

func (this IndexedSliceAccess) Iterator() Iterator {
	return NewIndexedIterator(this)
}

func (this IndexedSliceAccess) Sort(cmp CompareFunction) IndexedSliceAccess {
	Sort(this, cmp)
	return this
}

/*
func (this IndexedSliceAccess) entry_iterator() entry_iterator {
	return (&_slice_entry_iterator{}).new(this)
}
*/

func NewSliceIterator(slice []interface{}) *IndexedIterator {
	return NewIndexedIterator(IndexedSliceAccess(slice))
}
