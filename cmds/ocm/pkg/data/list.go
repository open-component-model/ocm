// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package data

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

func (ll *LinkedList) New() *LinkedList {
	if ll == nil {
		ll = &LinkedList{}
	}
	ll.root.New(ll)
	return ll
}

func (ll *LinkedList) Append(e interface{}) ListElement {
	n := &listElement{element: e}
	n.dll.payload = n
	ll.root.Append(&n.dll)
	return n
}

func (ll *LinkedList) Iterator() Iterator {
	return &listIterator{ll.root.Iterator().(*dllIterator)}
}

func (ll *LinkedList) ElementIterator() ElementIterator {
	return &listIterator{ll.root.Iterator().(*dllIterator)}
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
