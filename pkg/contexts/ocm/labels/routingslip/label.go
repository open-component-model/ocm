// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package routingslip

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/handlers/simplelistmerge"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/handlers/simplemapmerge"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/hpi"
	"github.com/open-component-model/ocm/pkg/utils"
)

const NAME = "routing-slips"

type Label map[string]RoutingSlip

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

func AddEntry(cv cpi.ComponentVersionAccess, name string, algo string, e Entry) error {
	var label Label
	_, err := cv.GetDescriptor().Labels.GetValue(NAME, &label)
	if err != nil {
		return err
	}
	if label == nil {
		label = Label{}
	}
	slip := label[name]
	err = slip.Add(cv.GetContext(), name, algo, e)
	if err != nil {
		return err
	}
	label[name] = slip

	return cv.GetDescriptor().Labels.SetValue(NAME, label)
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
