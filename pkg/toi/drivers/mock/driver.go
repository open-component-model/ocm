// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"github.com/open-component-model/ocm/pkg/toi/install"
)

type Driver struct{}

var _ install.Driver = (*Driver)(nil)

func New() install.Driver {
	return &Driver{}
}

func (d Driver) SetConfig(props map[string]string) error {
	return nil
}

func (d Driver) Exec(op *install.Operation) (*install.OperationResult, error) {
	if handler != nil {
		return handler(op)
	}
	return &install.OperationResult{}, nil
}

var handler func(*install.Operation) (*install.OperationResult, error)

func SetHandler(h func(*install.Operation) (*install.OperationResult, error)) {
	handler = h
}
