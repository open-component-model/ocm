// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transferhandler

import (
	"sync"
)

type entry struct {
	prio    int
	creator TransferOptionsCreator
}

type registry struct {
	lock sync.Mutex

	list []*entry
}

func (r *registry) Register(prio int, c TransferOptionsCreator) {
	r.lock.Lock()
	defer r.lock.Unlock()

	n := &entry{prio, c}
	for i, e := range r.list {
		if e.prio < n.prio {
			r.list = append(r.list[:i], append([]*entry{n}, r.list[i:]...)...)
			return
		}
	}
	r.list = append(r.list, n)
}

func (r *registry) OrderedTransferOptionCreators() []TransferOptionsCreator {
	var list []TransferOptionsCreator
	r.lock.Lock()
	defer r.lock.Unlock()

	for _, e := range r.list {
		list = append(list, e.creator)
	}
	return list
}

var optionsets = &registry{}

// RegisterHandler registers handler specific option set types
// for option based handler detection.
// Every transfer options provides a method used to create
// an initial handler option set it is originally defined for.
// This way, in a first step, by evaluating all given options
// the detection process tries to find an option set applicable
// to consume all given options. The set then is responsible to
// provide an appropriately configured handler.
//
// Option set handlers are used if none of the given options
// provides an option set applicable to accept all given
// handler related options. In a second step all such handlers
// ordered by their priority are used to create an option set.
// If it matches all given options, it is used to finally create
// the transfer handler.
//
// Self-contained handlers should use priority 1000.
// Handlers extending option sets from other handlers
// should use priority 100.
// Handlers supporting the union of incompatible option sets
// should use lower priority according to the order they should be checked.
func RegisterHandler(prio int, creator TransferOptionsCreator) {
	optionsets.Register(prio, creator)
}

func OrderedTransferOptionCreators() []TransferOptionsCreator {
	return optionsets.OrderedTransferOptionCreators()
}
