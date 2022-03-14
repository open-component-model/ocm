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

package utils

import (
	"github.com/gardener/ocm/cmds/ocm/pkg/output"
	"github.com/gardener/ocm/pkg/errors"
)

type TypeHandler interface {
	Get(name string) ([]output.Object, error)
	Close() error
}

func HandleArgs(outputs output.Outputs, opts *output.Options, handler TypeHandler, args ...string) error {
	defer handler.Close()
	if err := opts.Complete(); err != nil {
		return err
	}
	output, err := outputs.Create(opts)
	if err != nil {
		return err
	}
	for _, a := range args {
		result, err := handler.Get(a)
		if err != nil {
			return errors.Wrapf(err, "error processing %q", a)
		}
		for _, r := range result {
			output.Add(nil, r)
		}
	}
	err = output.Close(nil)
	if err != nil {
		return err
	}
	return output.Out(nil)
}
