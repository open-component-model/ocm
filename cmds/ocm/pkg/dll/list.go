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

package dll

type Appender interface {
	Append(interface{}) ListElement
}

type ListElement interface {
	Set(interface{})
	Get() interface{}
	Append(interface{}) ListElement
	Insert(interface{}) ListElement
}

type listElement struct {
	element interface{}
	dll     DLL
}

var _ ListElement = (*listElement)(nil)

func (l *listElement) Set(e interface{}) {
	l.element = e
}

func (l *listElement) Get() interface{} {
	return l.element
}

func (l *listElement) Append(e interface{}) ListElement {
	n := &listElement{element: e}
	n.dll.payload = n
	l.dll.Append(&n.dll)
	return n
}

func (l *listElement) Insert(e interface{}) ListElement {
	n := &listElement{element: e}
	n.dll.payload = n
	l.dll.Prev().Append(&n.dll)
	return n
}

type ElementIterator interface {
	HasNext() bool
	NextElement() ListElement
}

type mappedIterator struct {
	ElementIterator
}

func (m *mappedIterator) Next() interface{} {
	return m.ElementIterator.NextElement()
}

func ElementIteratorAsIterator(i ElementIterator) Iterator {
	return &mappedIterator{i}
}

type LinkedList struct {
	root DLLRoot
}

var _ Appender = (*LinkedList)(nil)

func NewLinkedList() *LinkedList {
	return (&LinkedList{}).New()
}

func (this *LinkedList) New() *LinkedList {
	if this == nil {
		this = &LinkedList{}
	}
	this.root.New(this)
	return this
}

func (this *LinkedList) Append(e interface{}) ListElement {
	n := &listElement{element: e}
	n.dll.payload = n
	this.root.Append(&n.dll)
	return n
}

func (this *LinkedList) Iterator() Iterator {
	return &listIterator{this.root.Iterator().(*dllIterator)}
}

func (this *LinkedList) ElementIterator() ElementIterator {
	return &listIterator{this.root.Iterator().(*dllIterator)}
}

type listIterator struct {
	it *dllIterator
}

var _ Iterator = (*listIterator)(nil)

func (l *listIterator) HasNext() bool {
	return l.it.HasNext()
}

func (l *listIterator) Next() interface{} {
	n := l.it.NextDLL()
	if n == nil {
		return nil
	}
	return n.payload.(*listElement).element
}

func (l *listIterator) NextElement() ListElement {
	n := l.it.NextDLL()
	if n == nil {
		return nil
	}
	return n.payload.(*listElement)
}
