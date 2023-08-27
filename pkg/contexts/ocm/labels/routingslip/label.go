// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package routingslip

import (
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/handlers/simplelistmerge"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/handlers/simplemapmerge"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/hpi"
	"github.com/open-component-model/ocm/pkg/utils"
)

const NAME = "routing-slips"

type Label map[string]HistoryEntries

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

func (l Label) Has(name string) bool {
	return l[name] != nil
}

func (l Label) Get(name string) *RoutingSlip {
	return &RoutingSlip{
		Name:    name,
		Entries: l[name],
	}
}

func (l Label) Query(name string) *RoutingSlip {
	a := l[name]
	if a == nil {
		return nil
	}
	return &RoutingSlip{
		Name:    name,
		Entries: a,
	}
}

func (l Label) Set(slip *RoutingSlip) {
	l[slip.Name] = slip.Entries
}

func AddEntry(cv cpi.ComponentVersionAccess, name string, algo string, e Entry, parent ...digest.Digest) (*HistoryEntry, error) {
	var label Label
	_, err := cv.GetDescriptor().Labels.GetValue(NAME, &label)
	if err != nil {
		return nil, err
	}
	if label == nil {
		label = Label{}
	}
	slip := label.Get(name)
	entry, err := slip.Add(cv.GetContext(), name, algo, e, parent...)
	if err != nil {
		return nil, err
	}
	label.Set(slip)

	err = cv.GetDescriptor().Labels.SetValue(NAME, label)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func Get(cv cpi.ComponentVersionAccess) (Label, error) {
	var label Label
	_, err := cv.GetDescriptor().Labels.GetValue(NAME, &label)
	if err != nil {
		return nil, err
	}
	return label, nil
}

func Set(cv cpi.ComponentVersionAccess, label Label) error {
	return cv.GetDescriptor().Labels.SetValue(NAME, label)
}
