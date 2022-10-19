// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package processing

import (
	"sync"
)

type ProcessorPool interface {
	Request()
	Release()
	Exec(func())
}

/////////////////////////////////////////////////////////////////////////////

type _UnlimitedPool struct{}

func NewUnlimitedProcessorPool() ProcessorPool {
	return &_UnlimitedPool{}
}

func (this *_UnlimitedPool) Request() {
}

func (this *_UnlimitedPool) Release() {
}

func (this *_UnlimitedPool) Exec(f func()) {
	go f()
}

/////////////////////////////////////////////////////////////////////////////

type _ProcessorPool struct {
	n    int
	uses int
	lock sync.Mutex
	set  *processorSet
}

func NewProcessorPool(n int) ProcessorPool {
	return &_ProcessorPool{n: n, uses: 0}
}

func (this *_ProcessorPool) Request() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.uses++
	if this.uses == 1 {
		this.set = (&processorSet{}).new(this.n)
	}
}

func (this *_ProcessorPool) Exec(f func()) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.uses == 0 {
		panic("unrequested processor pool used")
	}
	this.set.exec(f)
}

func (this *_ProcessorPool) Release() {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.uses > 0 {
		this.uses--
		if this.uses <= 0 && this.set != nil {
			this.set.stop()
			this.set = nil
		}
	}
}

/////////////////////////////////////////////////////////////////////////////

type processorSet struct {
	request chan func()
}

func (this *processorSet) new(n int) *processorSet {
	this.request = make(chan func(), n*2)
	for i := 0; i < n; i++ {
		go func() {
			for f := range this.request {
				f()
			}
		}()
	}
	return this
}

func (this *processorSet) exec(f func()) {
	this.request <- f
}

func (this *processorSet) stop() {
	close(this.request)
}
