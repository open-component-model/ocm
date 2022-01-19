// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package logcontext

import (
	"context"

	"github.com/go-logr/logr"
)

// logContextKey is the unique key for storing additional logging contexts
type logContextKey struct{}

// ContextValues describes the context values.
type ContextValues map[string]interface{}

func NewContext(parent context.Context) (context.Context, *ContextValues) {
	vals := &ContextValues{}
	return context.WithValue(parent, logContextKey{}, vals), vals
}

// FromContext returns the context values of a ctx.
// If nothing is defined nil is returned.
func FromContext(ctx context.Context) *ContextValues {
	c, ok := ctx.Value(logContextKey{}).(*ContextValues)
	if !ok {
		return nil
	}
	return c
}

// AddContextValue adds a key value pair to the logging context.
// If none is defined it will be added.
func AddContextValue(ctx context.Context, key string, value interface{}) context.Context {
	logCtx := FromContext(ctx)
	if logCtx == nil {
		ctx, logCtx = NewContext(ctx)
	}
	(*logCtx)[key] = value
	return ctx
}

// ctxSink defines a log sink that injects the provided context values
// and delegates the actual logging to a delegate.
type ctxSink struct {
	logr.LogSink
	ctx *ContextValues
}

// New creates a new context logger that delegates the actual requests to the delegate
// but injects the context log values.
func New(ctx context.Context, delegate logr.Logger) logr.Logger {
	val := FromContext(ctx)
	if val == nil {
		return delegate
	}
	return delegate.WithSink(newWithContextValues(val, delegate.GetSink()))
}

func newWithContextValues(ctx *ContextValues, del logr.LogSink) logr.LogSink {
	return &ctxSink{
		LogSink: del,
		ctx:     ctx,
	}
}

func (c ctxSink) Error(err error, msg string, keysAndValues ...interface{}) {
	// append log context values
	if c.ctx == nil {
		c.LogSink.Error(err, msg, keysAndValues...)
		return
	}
	for key, val := range *c.ctx {
		keysAndValues = append(keysAndValues, key, val)
	}
	c.LogSink.Error(err, msg, keysAndValues...)
}
