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

import "fmt"

type Appender interface {
	Append(interface{}) (ListElement, error)
}

type ListElement interface {
	Set(interface{})
	Get() interface{}
	Append(interface{}) (ListElement, error)
	Insert(interface{}) (ListElement, error)
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

func (l *listElement) Append(e interface{}) (ListElement, error) {
	n := &listElement{element: e}
	n.dll.payload = n
	if err := l.dll.Append(&n.dll); err != nil {
		return nil, fmt.Errorf("list element append failed: %s", err)
	}
	return n, nil
}

func (l *listElement) Insert(e interface{}) (ListElement, error) {
	n := &listElement{element: e}
	n.dll.payload = n
	if err := l.dll.Prev().Append(&n.dll); err != nil {
		return nil, fmt.Errorf("list element insert failed: %w", err)
	}
	return n, nil
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

func (this *LinkedList) Append(e interface{}) (ListElement, error) {
	n := &listElement{element: e}
	n.dll.payload = n
	if err := this.root.Append(&n.dll); err != nil {
		return nil, fmt.Errorf("linked list append failed: %w", err)
	}
	return n, nil
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
