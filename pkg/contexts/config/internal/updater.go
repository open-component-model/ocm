// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal

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
	State() (int64, bool)
	GetContext() Context

	Lock()
	Unlock()
	RLock()
	RUnlock()
}

// updater implements the Updater interface.
// It must be prepared to work with an internal config context
// representation. Therefore, it must pass a self reference
// of the context to outbound calls.
type updater struct {
	sync.RWMutex
	ctx            Context
	targetFunc     func() interface{}
	lastGeneration int64
	inupdate       bool
}

// TargetFunction can be used to map any type specific factory function
// to a target function returning a formal interface{} type.
func TargetFunction[T any](f func() T) func() interface{} {
	return func() interface{} { return f() }
}

// NewUpdater create a configuration updater for a configuration target
// based on a dedicated configuration context.
func NewUpdater(ctx Context, target interface{}) Updater {
	var targetFunc func() interface{}
	if f, ok := target.(func() interface{}); ok {
		targetFunc = f
	} else {
		targetFunc = func() interface{} { return target }
	}
	return &updater{
		ctx:        ctx,
		targetFunc: targetFunc,
	}
}

func NewUpdaterForFactory[T any](ctx Context, t func() T) Updater {
	return &updater{
		ctx:        ctx,
		targetFunc: TargetFunction(t),
	}
}

func (u *updater) GetContext() Context {
	return u.ctx.ConfigContext()
}

func (u *updater) GetTarget() interface{} {
	return u.targetFunc()
}

func (u *updater) State() (int64, bool) {
	u.RLock()
	defer u.RUnlock()
	return u.lastGeneration, u.inupdate
}

func (u *updater) Update() error {
	u.Lock()
	if u.inupdate {
		u.Unlock()
		return nil
	}
	u.inupdate = true
	u.Unlock()

	gen, err := u.ctx.ApplyTo(u.lastGeneration, u.GetTarget())

	u.Lock()
	defer u.Unlock()
	u.inupdate = false
	u.lastGeneration = gen

	if err != nil {
		Logger.LogError(err, "config update failed", "id", u.ctx.GetId())
	}
	return err
}
