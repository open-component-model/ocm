/*
 * Copyright 2023 Mandelsoft. All rights reserved.
 *  This file is licensed under the Apache Software License, v. 2 except as noted
 *  otherwise in the LICENSE file
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package logging

import (
	"github.com/go-logr/logr"
)

type attributionContext struct {
	ctx            Context
	messageContext []MessageContext

	values []interface{}
}

var _ AttributionContext = (*attributionContext)(nil)

// NewAttributionContext returns a new [AttributionContext]
// for a [Context] with an optional additional [MessageContext].
func NewAttributionContext(ctxp ContextProvider, messageContext ...MessageContext) AttributionContext {
	l := &attributionContext{
		ctx:            ctxp.LoggingContext(),
		messageContext: explode(messageContext),
	}
	return l
}

func (d *attributionContext) AttributionContext() AttributionContext {
	return d
}

func (d *attributionContext) LoggingContext() Context {
	return d.ctx
}

func (d *attributionContext) GetMessageContext() []MessageContext {
	return append(d.ctx.Tree().GetMessageContext(), d.messageContext...)
}

func (d *attributionContext) WithName(name string) AttributionContext {
	return d.WithContext(NewName(name))
}

func (d *attributionContext) WithValues(keypairs ...interface{}) AttributionContext {
	if len(keypairs) == 0 {
		return d
	}
	l := *d
	l.values = sliceAppend(l.values, keypairs[:2*(len(keypairs)/2)]...)
	return &l
}

func (d *attributionContext) WithContext(messageContext ...MessageContext) AttributionContext {
	if len(messageContext) == 0 {
		return d
	}
	l := *d
	l.messageContext = sliceAppend(l.messageContext, explode(messageContext)...)
	return &l
}

func (d *attributionContext) Match(cond Condition) bool {
	return cond.Match(append(d.ctx.GetMessageContext(), d.messageContext...))
}

func (d *attributionContext) Logger(messageContext ...MessageContext) Logger {
	l := d.ctx.Logger(sliceAppend(d.messageContext, messageContext...))

	if len(d.values) > 0 {
		l = l.WithValues(d.values...)
	}
	return l
}

func (d *attributionContext) LoggerFor(messageContext ...MessageContext) Logger {
	l := d.ctx.LoggerFor(messageContext...)
	if len(d.values) > 0 {
		l = l.WithValues(d.values...)
	}
	return l
}

func (d *attributionContext) V(level int, mctx ...MessageContext) logr.Logger {
	return d.Logger(mctx...).V(level)
}
