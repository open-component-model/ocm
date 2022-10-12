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

package core

import (
	"sync"
)

// Updater implements the generation based update protocol
// to update data contexts based on the config requests
// made to a configuration context.
type Updater interface {
	// Update replays missing configuration requests
	// applicable for a dedicated type of context or configuration target
	// stored in a configuration context.
	// It should be created for and called from within such a context
	Update() error
	GetContext() Context

	Lock()
	Unlock()
	RLock()
	RUnlock()
}

type updater struct {
	sync.RWMutex
	ctx            Context
	target         interface{}
	lastGeneration int64
	inupdate       bool
}

// NewUpdater create a configuration updater for a configuration target
// based on a dedicated configuration context.
func NewUpdater(ctx Context, target interface{}) Updater {
	return &updater{
		ctx:    ctx,
		target: target,
	}
}

func (u *updater) GetContext() Context {
	return u.ctx
}

func (u *updater) GetTarget() interface{} {
	return u.target
}

func (u *updater) Update() error {
	u.Lock()
	if u.inupdate {
		u.Unlock()
		return nil
	}
	u.inupdate = true
	u.Unlock()

	gen, err := u.ctx.ApplyTo(u.lastGeneration, u.target)

	u.Lock()
	defer u.Unlock()
	u.inupdate = false
	u.lastGeneration = gen
	return err
}
