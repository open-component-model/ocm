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

// RegisterHandler registers handler specific option set type
// for option based handler detection.
// Self-contained handlers should use priority 1000.
// Handlers incorporating extending option sets from other handlers
// should use priority 100.
// Handlers supporting the union of incompatible option sets
// should use lower priority according to the order they should be checked.
func RegisterHandler(prio int, creator TransferOptionsCreator) {
	optionsets.Register(prio, creator)
}

func OrderedTransferOptionCreators() []TransferOptionsCreator {
	return optionsets.OrderedTransferOptionCreators()
}
