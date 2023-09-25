// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package routingslip

import (
	"sort"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/handlers/simplelistmerge"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/handlers/simplemapmerge"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/hpi"
	"github.com/open-component-model/ocm/pkg/utils"
)

const NAME = "routing-slips"

type LabelValue map[string]HistoryEntries

var spec = utils.Must(hpi.NewSpecification(
	simplemapmerge.ALGORITHM,
	simplemapmerge.NewConfig(
		"",
		utils.Must(hpi.NewSpecification(
			simplelistmerge.ALGORITHM,
			simplelistmerge.NewConfig(),
		)),
	)),
)

func init() {
	hpi.Assign(hpi.LabelHint(NAME), spec)
}

func (l LabelValue) Has(name string) bool {
	return l[name] != nil
}

func (l LabelValue) Get(name string) *RoutingSlip {
	return NewRoutingSlip(name, l, l[name]...)
}

func (l LabelValue) Query(name string) *RoutingSlip {
	a := l[name]
	if a == nil {
		return nil
	}
	return l.Get(name)
}

func (l LabelValue) Leaves() []Link {
	var links []Link

	for k := range l {
		for _, d := range l.Get(k).Leaves() {
			links = append(links, Link{
				Name:   k,
				Digest: d,
			})
		}
	}
	sort.Slice(links, func(i, j int) bool { return links[i].Compare(links[j]) < 0 })
	return links
}

func (l LabelValue) Set(slip *RoutingSlip) {
	l[slip.name] = slip.entries
}

func AddEntry(cv cpi.ComponentVersionAccess, name string, algo string, e Entry, links []Link, parent ...digest.Digest) (*HistoryEntry, error) {
	var label LabelValue
	_, err := cv.GetDescriptor().Labels.GetValue(NAME, &label)
	if err != nil {
		return nil, err
	}
	if label == nil {
		label = LabelValue{}
	}
	slip := label.Get(name)
	entry, err := slip.Add(cv.GetContext(), name, algo, e, links, parent...)
	if err != nil {
		return nil, err
	}
	label.Set(slip)

	err = Set(cv, label)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func Get(cv cpi.ComponentVersionAccess) (LabelValue, error) {
	var label LabelValue
	_, err := cv.GetDescriptor().Labels.GetValue(NAME, &label)
	if err != nil {
		return nil, err
	}
	return label, nil
}

func Set(cv cpi.ComponentVersionAccess, label LabelValue) error {
	return cv.GetDescriptor().Labels.SetValue(NAME, label)
}
