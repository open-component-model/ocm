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

package mock

import (
	"github.com/open-component-model/ocm/pkg/toi/install"
)

type Driver struct {
}

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
