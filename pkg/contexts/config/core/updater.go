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

	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
)

// Updater implements the generation based update protocol
// to update data contexts based on the config requests
// made to a configuration context.
type Updater interface {
	// Update replays missing configuration requests
	// applicable for a dedicated type of data context
	// stored in a configuration context.
	// It should be called from with such a context with
	// the actual context as argument.
	Update(target datacontext.Context) error
	GetContext() Context

	Lock()
	Unlock()
	RLock()
	RUnlock()
}

type updater struct {
	sync.RWMutex
	ctx            Context
	lastGeneration int64
	inupdate       bool
}

func NewUpdater(ctx Context) Updater {
	return &updater{
		ctx: ctx,
	}
}

func (u *updater) GetContext() Context {
	return u.ctx
}

func (u *updater) Update(target datacontext.Context) error {
	u.Lock()
	if u.inupdate {
		u.Unlock()
		return nil
	}
	u.inupdate = true
	u.Unlock()

	gen, err := u.ctx.ApplyTo(u.lastGeneration, target)

	u.Lock()
	defer u.Unlock()
	u.inupdate = false
	u.lastGeneration = gen
	return err
}
