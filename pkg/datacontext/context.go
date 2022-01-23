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

package datacontext

import (
	"context"
	"sync"
)

// Context describes a common interface for a data context used for a dedicated
// purpose.
// Such a context incorporates a context.Context and some context
// specific attribute store
type Context interface {
	context.Context
	GetAttributes() Attributes
}

type Attributes interface {
	GetAttribute(name string) interface{}
	SetAttribute(name string, value interface{})
	GetOrCreateAttribute(name string, creator func(context.Context) interface{}) interface{}
}

////////////////////////////////////////////////////////////////////////////////
// Here we provide a default implementation for the data context mechanics.
//
// The context.Context always provides access to a dedicated instance of
// a data context. This is done by a context values and a type specif key.
//
// Such a context always incorporates a regular context.Context and can
// therefore be used and passed as regular context.Context
//
// To support this the effective data can be rebased to any leaf context it is
// taken from. After rebasing the context object represents the same data context
// but with the effective latest context.Context
//
// The following default implementation implements this behaviour in a generic way.
// Therefore is provides a base context implementation that should be used
// as only anonymous field in the effective context type. This one implements
// the dedicated context api based on a context data object.
// The context data is separated into a dedicated type which must fulfill the
// DataContext interface. It must provide a Wrap function to Wrap
// an new instance of the DefaultContext into an appropriate context type.
//
// The default implementation provides a DefaultContext, which acts as link
// between the context.Context and the context data. It will then handle
// the rebase mechanics for the context data.

// DataContext is the interface for the context data described by a data context:
// It is used by the default context base implementation
// for providing access to the context data up the context hierarchy for context.Context
// objects down the hierarchy.
type DataContext interface {
	AttributesContext

	// Wrap returns a new type specific implementation for the data context and the value key
	// to make it accessible by the context.Context
	Wrap(DefaultContext) (DefaultContext, interface{})
}

// AttributesContext is the attribute access interface of the DataContext contract.
// The default implementation assures thet the create functionality is always
// used with the latest rebased/known context.Context
type AttributesContext interface {
	GetAttribute(name string) interface{}
	SetAttribute(name string, value interface{})
	GetOrCreateAttribute(ctx context.Context, name string, creator func(context.Context) interface{}) interface{}
}

////////////////////////////////////////////////////////////////////////////////

// DefaultContext is the interface provided the default rebase implementation
// It supports the rebasing and the attribute access of the context based on the
// DataContext interface provided by the context provider.
type DefaultContext interface {
	Context

	// DefaultAccess provides access to the functionality of the default implementation
	DefaultAccess() DefaultAccess
}

// DefaultAccess provides access to the default implementation.
// It cannot be  part of the context interface, because the signature
// of the with method would violate the signature of the effective type.
// Basically this separated the method namespace of the default implementation
// from the one of the effective context type.
type DefaultAccess interface {
	// With implements a standard behaviour for the With interface of a dedicated data context
	// using this standard implementation. It must be forwarded to the With method
	// of the actual context type with an appropriate type cast.
	// It works together with the Wrap function of the DataContext interface.
	// The dedicated context type should just use a single field with this type
	// and provide With method with an appropriately typed one.
	With(ctx context.Context) Context

	// DataContext returns the actual context data
	DataContext() DataContext
}

type _context struct {
	context.Context
	data DataContext
}

// NewContext provides a default base implementation for a data context.
// featuring context and attribute access.
func NewContext(ctx context.Context, data DataContext) DefaultContext {
	return forContext(ctx, data)
}

// ForContext is a rebase utility for data context.
// It can be used by data context implementations using this default implementation
// to implement the data context access based on context.Context.
//
// Every such data context type should support such a ForContext operation for
// its consumers yielding the appropriate type.
//
// This default support implementation returns never nil, if the default is
// set. Otherwise, nit is returned, if there is no data context found for the
// actual context.
//
// The returned value is a data context rebased to the actual context.
func ForContext(ctx context.Context, key interface{}, def Context) DefaultContext {
	data := ctx.Value(key)
	if data == nil {
		data = def
	}
	if data == nil {
		return nil
	}
	return forContext(ctx, data.(DefaultContext).DefaultAccess().DataContext())
}

func forContext(ctx context.Context, data DataContext) DefaultContext {
	c := &_context{
		data: data,
	}
	eff, key := c.data.Wrap(c)
	c.Context = context.WithValue(ctx, key, eff)
	return eff
}

func (c *_context) GetAttributes() Attributes {
	return &_attributesContext{c}
}

func (c *_context) DefaultAccess() DefaultAccess {
	return c
}

func (c *_context) With(ctx context.Context) Context {
	if ctx == c.Context {
		return c
	}
	return forContext(ctx, c.data)
}

func (c *_context) DataContext() DataContext {
	return c.data
}

////////////////////////////////////////////////////////////////////////////////

type _attributesContext struct {
	ctx *_context
}

var _ Attributes = &_attributesContext{}

func (a *_attributesContext) GetAttribute(name string) interface{} {
	return a.ctx.data.GetAttribute(name)
}

func (a *_attributesContext) SetAttribute(name string, value interface{}) {
	a.ctx.data.SetAttribute(name, value)
}

func (a *_attributesContext) GetOrCreateAttribute(name string, creator func(context.Context) interface{}) interface{} {
	return a.ctx.data.GetOrCreateAttribute(a.ctx.Context, name, creator)
}

////////////////////////////////////////////////////////////////////////////////

type _attributes struct {
	sync.RWMutex
	parent     Attributes
	attributes map[string]interface{}
}

var _ AttributesContext = &_attributes{}

func NewAttributes(parent Attributes) AttributesContext {
	return &_attributes{
		parent:     parent,
		attributes: map[string]interface{}{},
	}
}

func (c *_attributes) GetAttribute(name string) interface{} {
	c.RLock()
	defer c.RUnlock()
	a := c.attributes[name]
	if a != nil {
		return a
	}
	if c.parent != nil {
		return c.parent.GetAttribute(name)
	}
	return nil
}

func (c *_attributes) SetAttribute(name string, value interface{}) {
	c.Lock()
	defer c.Unlock()
	c.attributes[name] = value
}

func (c *_attributes) GetOrCreateAttribute(ctx context.Context, name string, creator func(context.Context) interface{}) interface{} {
	c.Lock()
	defer c.Unlock()
	if v, ok := c.attributes[name]; ok {
		return v
	}
	v := creator(ctx)
	c.attributes[name] = v
	return v
}
