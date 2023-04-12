// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"github.com/open-component-model/ocm/pkg/toi/install"
	"github.com/open-component-model/ocm/pkg/utils"
)

type Driver struct {
	handler func(*install.Operation) (*install.OperationResult, error)
}

var _ install.Driver = (*Driver)(nil)

func New(handler ...func(*install.Operation) (*install.OperationResult, error)) install.Driver {
	return &Driver{utils.Optional(handler...)}
}

func (d *Driver) SetConfig(props map[string]string) error {
	return nil
}

func (d *Driver) Exec(op *install.Operation) (*install.OperationResult, error) {
	if d.handler != nil {
		return d.handler(op)
	}
	return &install.OperationResult{}, nil
}
