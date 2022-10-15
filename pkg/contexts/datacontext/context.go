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
	"io"
	"reflect"
	"sync"

	"github.com/mandelsoft/logging"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/errors"
	ocmlog "github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const OCM_CONTEXT_SUFFIX = ".context" + common.OCM_TYPE_GROUP_SUFFIX

// BuilderMode controls the handling of unset information in the
// builder configuration when calling the New method.
type BuilderMode int

const (
	// MODE_SHARED uses the default contexts for unset nested context types.
	MODE_SHARED BuilderMode = iota
	// MODE_DEFAULTED uses dedicated context instances configured with the
	// context type specific default registrations.
	MODE_DEFAULTED
	// MODE_CONFIGURED uses dedicated context instances configured with the
	// context type registrations configured with the actual state of the
	// default registrations.
	MODE_CONFIGURED
	// MODE_INITIAL uses completely new contexts for unset nested context types
	// and initial registrations.
	MODE_INITIAL
)

func Mode(m ...BuilderMode) BuilderMode {
	mode := MODE_DEFAULTED
	if len(m) > 0 {
		mode = m[0]
	}
	return mode
}

// Context describes a common interface for a data context used for a dedicated
// purpose.
// Such has a type and always specific attribute store.
// Every Context can be bound to a context.Context.
type Context interface {
	// GetType returns the context type
	GetType() string

	// AttributesContext returns the shared attributes
	AttributesContext() AttributesContext
	// BindTo binds the context to a context.Context and makes it
	// retrievable by a ForContext method
	BindTo(ctx context.Context) context.Context
	GetAttributes() Attributes

	ocmlog.LogProvider
}

////////////////////////////////////////////////////////////////////////////////

// CONTEXT_TYPE is the global type for an attribute context.
const CONTEXT_TYPE = "attributes" + OCM_CONTEXT_SUFFIX

type AttributesContext interface {
	Context

	BindTo(ctx context.Context) context.Context
}

// AttributeFactory is used to atomicly create a new attribute for a context.
type AttributeFactory func(Context) interface{}

type Attributes interface {
	GetAttribute(name string, def ...interface{}) interface{}
	SetAttribute(name string, value interface{}) error
	SetEncodedAttribute(name string, data []byte, unmarshaller runtime.Unmarshaler) error
	GetOrCreateAttribute(name string, creator AttributeFactory) interface{}
}

// DefaultContext is the default context initialized by init functions.
var DefaultContext = New(nil)

// ForContext returns the Context to use for context.Context.
// This is either an explicit context or the default context.
func ForContext(ctx context.Context) AttributesContext {
	c, _ := ForContextByKey(ctx, key, DefaultContext)
	if c == nil {
		return nil
	}
	return c.(AttributesContext)
}

// WithContext create a new Context bound to a context.Context.
func WithContext(ctx context.Context, parentAttrs Attributes) (Context, context.Context) {
	c := New(parentAttrs)
	return c, c.BindTo(ctx)
}

////////////////////////////////////////////////////////////////////////////////

type Updater interface {
	Update() error
}

type UpdateFunc func() error

func (u UpdateFunc) Update() error {
	return u()
}

type contextBase struct {
	ctxtype    string
	key        interface{}
	effective  Context
	attributes Attributes
	logging    logging.Context
}

var _ Context = (*contextBase)(nil)

// NewContextBase creates a context base implementation supporting
// context attributes and the binding to a context.Context.
func NewContextBase(eff Context, typ string, key interface{}, parentAttrs Attributes, parentLogging logging.Context) Context {
	updater, _ := eff.(Updater)
	c := &contextBase{ctxtype: typ, key: key, effective: eff}
	c.attributes = newAttributes(eff, parentAttrs, &updater)
	c.logging = logging.NewWithBase(parentLogging)
	return c
}

func (c *contextBase) GetType() string {
	return c.ctxtype
}

// BindTo make the Context reachable via the resulting context.Context.
func (c *contextBase) BindTo(ctx context.Context) context.Context {
	return context.WithValue(ctx, c.key, c.effective)
}

func (c *contextBase) AttributesContext() AttributesContext {
	return c
}

func (c *contextBase) GetAttributes() Attributes {
	return c.attributes
}

func (c *contextBase) LoggingContext() logging.Context {
	return c.logging
}

func (c *contextBase) Logger(messageContext ...logging.MessageContext) logging.Logger {
	return c.logging.Logger(messageContext...)
}

////////////////////////////////////////////////////////////////////////////////

type _context struct {
	Context
	updater Updater
}

var key = reflect.TypeOf(contextBase{})

// New provides a root attribute context.
func New(parentAttrs Attributes) AttributesContext {
	c := &_context{}

	c.Context = &contextBase{
		ctxtype:    CONTEXT_TYPE,
		key:        key,
		effective:  c,
		attributes: newAttributes(c, parentAttrs, &c.updater),
		logging:    logging.NewWithBase(ocmlog.Context()),
	}
	return c
}

// AssureUpdater is used to assure the existence of an updater in
// a root context if a config context is down the context hierarchy.
// This method SHOULD only be called by a config context.
func AssureUpdater(attrs AttributesContext, u Updater) {
	c, ok := attrs.(*_context)
	if !ok {
		return
	}
	if c.updater == nil {
		c.updater = u
	}
}

func (c *_context) LoggingContext() logging.Context {
	if c.updater != nil {
		c.updater.Update()
	}
	return c.Context.LoggingContext()
}

func (c *_context) Logger(messageContext ...logging.MessageContext) logging.Logger {
	if c.updater != nil {
		c.updater.Update()
	}
	return c.Context.Logger(messageContext...)
}

////////////////////////////////////////////////////////////////////////////////

type _attributes struct {
	sync.RWMutex
	ctx        Context
	parent     Attributes
	updater    *Updater
	attributes map[string]interface{}
}

var _ Attributes = &_attributes{}

func NewAttributes(ctx Context, parent Attributes, updater *Updater) Attributes {
	return newAttributes(ctx, parent, updater)
}

func newAttributes(ctx Context, parent Attributes, updater *Updater) *_attributes {
	return &_attributes{
		ctx:        ctx,
		parent:     parent,
		updater:    updater,
		attributes: map[string]interface{}{},
	}
}

func (c *_attributes) GetAttribute(name string, def ...interface{}) interface{} {
	if *c.updater != nil {
		(*c.updater).Update()
	}
	c.RLock()
	defer c.RUnlock()
	if a := c.attributes[name]; a != nil {
		return a
	}
	if c.parent != nil {
		if a := c.parent.GetAttribute(name); a != nil {
			return a
		}
	}
	for _, d := range def {
		if d != nil {
			return d
		}
	}
	return nil
}

func (c *_attributes) SetEncodedAttribute(name string, data []byte, unmarshaller runtime.Unmarshaler) error {
	s := DefaultAttributeScheme.Shortcuts()[name]
	if s != "" {
		name = s
	}
	v, err := DefaultAttributeScheme.Decode(name, data, unmarshaller)
	if err != nil {
		return err
	}
	c.SetAttribute(name, v)
	return nil
}

func (c *_attributes) SetAttribute(name string, value interface{}) error {
	c.Lock()
	defer c.Unlock()

	_, err := DefaultAttributeScheme.Encode(name, value, nil)
	if err != nil && !errors.IsErrUnknownKind(err, "attribute") {
		return err
	}
	old := c.attributes[name]
	if old != nil && old != value {
		if c, ok := old.(io.Closer); ok {
			c.Close()
		}
	}
	c.attributes[name] = value
	return nil
}

func (c *_attributes) GetOrCreateAttribute(name string, creator AttributeFactory) interface{} {
	c.Lock()
	defer c.Unlock()
	if v := c.attributes[name]; v != nil {
		return v
	}
	if c.parent != nil {
		if v := c.parent.GetAttribute(name); v != nil {
			return v
		}
	}
	v := creator(c.ctx)
	c.attributes[name] = v
	return v
}
