// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

// Package resource provided some support to implement
// closeable backing resources featuring multiple
// separately closeable references. The backing resource
// is finally closed, when the last reference is closed.
// hereby, the reference implements the intended resource
// interface including the reference related part, which
// includes a Dup method, which can be used to gain a
// new additional reference to the backing object.
//
// Those references are called View in the package.
// The backing object implements the pure resource
// object interface plus the final Close method.
//
// The final resource interface is described by a Go
// interface including the resource.ResourceView interface,
//
//	type MyResource interface {
//	   resource.ResourceView[MyResource]
//	   AdditionalMethods()...
//	}
//
// The resource.ResourceView interface offers the view-related
// methods.
//
// With NewResource a new reference management and a first
// view is created for this object. This method is typically
// wrapped by a dedicated resource creator function:
//
//	func New(args...) MyResource {
//	   i := MyResourceImpl{
//	          ...
//	        }
//	   _, r := resource.NewResource(i, myViewCreator)
//	   return r
//	}
//
// The management as well as the view can be used to create
// additional views.
//
// Therefore, the reference management uses a ResourceViewCreator
// function, which must be provided by the object implementation
// Its task is to create a new frontend view object implementing
// the desired pure backing object functionality plus the
// view-related interface.
//
// This is done by creating an object with two embedded fields:
//
//	type MyReference struct {
//	   resource.ReferenceView[MyInterface]
//	   MyImplementation
//	}
//
// the myViewCreator function creates a new resource reference using the
// resource.NewView function.
//
//	func myViewCreator(impl *ResourceImpl,
//	                   v resource.CloserView,
//	                   d resource.Dup[Resource]) MyResource {
//	  return &MyResource {
//	           resource.NewView(v, d),
//	           impl,
//	         }
//	}
package resource

import (
	"io"

	"github.com/open-component-model/ocm/pkg/common/accessio"
)

type CloserView = accessio.CloserView

var ErrClosed = accessio.ErrClosed

// resourceViewInterface is a helper type used to implement parameter type
// recursion for ResourceView[T ResourceView[T]], which is not allowed in Go.
type resourceViewInterface[T any] interface {
	io.Closer
	IsClosed() bool
	Dup[T]
}

// ResourceView is the view related part of a resource interface T.
// T must incorporate ResourceView[T], which cannot directly be expressed
// in go, but with the helper interface defining the API.
type ResourceView[T resourceViewInterface[T]] interface {
	resourceViewInterface[T]
}

type Dup[T any] interface {
	Dup() (T, error)
}

// ResourceViewCreator is a function which must be provided by the resource provider
// to map an implementation to the resource interface T.
// It must use NewView to create the view related part of a resource.
type ResourceViewCreator[T any, I io.Closer] func(I, CloserView, Dup[T]) T

////////////////////////////////////////////////////////////////////////////////

type resourceInt[T any, I io.Closer] struct {
	refs    accessio.ReferencableCloser
	creator ResourceViewCreator[T, I]
	impl    I
}

// NewResource creates a resource based on an implementation and a ResourceViewCreator.
// function.
func NewResource[T any, I io.Closer](impl I, c ResourceViewCreator[T, I], main ...bool) (Dup[T], T) {
	i := &resourceInt[T, I]{
		refs:    accessio.NewRefCloser(impl, true),
		creator: c,
		impl:    impl,
	}
	t, _ := i.View(main...)
	return i, t
}

func (i *resourceInt[T, I]) Dup() (T, error) {
	return i.View()
}

func (i *resourceInt[T, I]) View(main ...bool) (T, error) {
	var _nil T

	v, err := i.refs.View(main...)
	if err != nil {
		return _nil, err
	}
	return i.creator(i.impl, v, i), nil
}

////////////////////////////////////////////////////////////////////////////////

type resourceView[T any] struct {
	view CloserView
	res  Dup[T]
}

// NewView is to be called by a resource view creator to map
// the given resource implementation to complete resource interface.
// It should, create an object with two local embedded fields:
//   - the returned ResourceView and the
//   - given resource implementation.
func NewView[T resourceViewInterface[T]](v CloserView, d Dup[T]) ResourceView[T] {
	return &resourceView[T]{v, d}
}

func (v *resourceView[T]) IsClosed() bool {
	return v.view.IsClosed()
}

func (v *resourceView[T]) Close() error {
	return v.view.Close()
}

func (v *resourceView[T]) Dup() (T, error) {
	return v.res.Dup()
}
