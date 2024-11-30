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
	"sync"

	"github.com/go-logr/logr"
)

type dynamicLogger struct {
	attribution AttributionContext
	lock        sync.Mutex
	watermark   int64
	logger      Logger
}

var _ Logger = (*dynamicLogger)(nil)
var _ ContextProvider = (*dynamicLogger)(nil)
var _ AttributionContextProvider = (*dynamicLogger)(nil)

// DynamicLogger returns an unbound logger, which automatically adapts to rule
// configuration changes applied to its logging context.
//
// Such a logger can be reused for multiple independent call trees
// without losing track to the config.
// Regular loggers provided by a context keep their setting from the
// matching rule valid during its creation.
func DynamicLogger(ctxp AttributionContextProvider, messageContext ...MessageContext) UnboundLogger {
	l := &dynamicLogger{
		attribution: ctxp.AttributionContext().WithContext(messageContext...),
	}
	return l
}

func (d *dynamicLogger) update() Logger {
	d.lock.Lock()
	defer d.lock.Unlock()
	// get watermark first to assure logger for at least the actual watermark.
	// this is not accurate in the sense of not necessarily being uptodate
	// with intermediate config requests, but this glitch does not hamper,
	// because the watermark assures update with the next call,
	// so no configs are finally lost.
	watermark := d.LoggingContext().Tree().Updater().Watermark()
	if d.logger == nil || watermark > d.watermark {
		// update logger and incorporate local modifications
		d.logger = d.attribution.Logger()
		d.watermark = watermark
	}
	return d.logger
}

func (d *dynamicLogger) LoggingContext() Context {
	return d.attribution.LoggingContext()
}

func (d *dynamicLogger) AttributionContext() AttributionContext {
	return d.attribution
}

func (d *dynamicLogger) LogError(err error, msg string, keypairs ...interface{}) {
	d.update().LogError(err, msg, keypairs...)
}

func (d *dynamicLogger) Error(msg string, keypairs ...interface{}) {
	d.update().Error(msg, keypairs...)
}

func (d *dynamicLogger) Warn(msg string, keypairs ...interface{}) {
	d.update().Warn(msg, keypairs...)
}

func (d *dynamicLogger) Info(msg string, keypairs ...interface{}) {
	d.update().Info(msg, keypairs...)
}

func (d *dynamicLogger) Debug(msg string, keypairs ...interface{}) {
	d.update().Debug(msg, keypairs...)
}

func (d *dynamicLogger) Trace(msg string, keypairs ...interface{}) {
	d.update().Trace(msg, keypairs...)
}

func (d *dynamicLogger) GetMessageContext() []MessageContext {
	return d.attribution.GetMessageContext()
}

func (d *dynamicLogger) WithName(name string) Logger {
	l := *d
	l.attribution = l.attribution.WithName(name)
	l.logger = nil
	return &l
}

func (d *dynamicLogger) WithValues(keypairs ...interface{}) Logger {
	if len(keypairs) == 0 {
		return d
	}
	l := *d
	l.attribution = l.attribution.WithValues(keypairs...)
	l.logger = nil
	return &l
}

func (d *dynamicLogger) WithContext(messageContext ...MessageContext) UnboundLogger {
	if len(messageContext) == 0 {
		return d
	}
	l := *d
	l.attribution = l.attribution.WithContext(messageContext...)
	l.logger = nil
	return &l
}

func (d *dynamicLogger) Enabled(level int) bool {
	return d.update().Enabled(level)
}

func (d *dynamicLogger) V(delta int) logr.Logger {
	return d.update().V(delta)
}

func (d *dynamicLogger) BoundLogger() Logger {
	return d.update()
}
