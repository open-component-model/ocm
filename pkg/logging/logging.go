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

package logging

import (
	"encoding/json"
	"sync"

	"github.com/mandelsoft/logging"
	logcfg "github.com/mandelsoft/logging/config"
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/errors"
)

type StaticContext struct {
	logging.Context
	applied map[string]struct{}
	lock    sync.Mutex
}

func NewContext(ctx logging.Context) *StaticContext {
	if ctx == nil {
		ctx = logging.DefaultContext()
	}
	return &StaticContext{
		Context: ctx,
		applied: map[string]struct{}{},
	}
}

// Configure applies a configuration once.
// Every config identified by its hash is applied
// only once.
func (s *StaticContext) Configure(config *logcfg.Config, extra ...string) error {
	add := ""
	for _, e := range extra {
		if e != "" {
			add += "/" + e
		}
	}
	data, err := json.Marshal(config)
	if err != nil {
		return errors.Wrapf(err, "cannot marshal log config")
	}
	d := digest.FromBytes(data).String() + add

	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.applied[d]; ok {
		return nil
	}
	s.applied[d] = struct{}{}
	return logcfg.Configure(logContext, config)
}

var logContext = NewContext(nil)

// SetContext sets a new precondigure context.
// This function should be called prior to any configuration
// to avoid loosing them.
func SetContext(ctx logging.Context) {
	logContext = NewContext(ctx)
}

// Context returns the default logging configuration used for this library.
func Context() *StaticContext {
	return logContext
}

// Logger determines a default logger for this given message context
// based on the rule settings for this library.
func Logger(messageContext ...logging.MessageContext) logging.Logger {
	return logContext.Logger(messageContext...)
}

// Configure applies configuration for the default log context
// provided by this package.
func Configure(config *logcfg.Config, extra ...string) error {
	return logContext.Configure(config, extra...)
}
