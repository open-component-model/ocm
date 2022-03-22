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
	"fmt"
	"os"
	"reflect"

	"github.com/gardener/ocm/cmds/ocm/pkg/output"
	"github.com/gardener/ocm/pkg/errors"
)

type ElemSpec interface {
	String() string
}

type StringSpec string

func (s StringSpec) String() string {
	return string(s)
}

type TypeHandler interface {
	All() ([]output.Object, error)
	Get(name ElemSpec) ([]output.Object, error)
	Close() error
}

func StringElemSpecs(args ...string) []ElemSpec {
	r := make([]ElemSpec, len(args))
	for i, v := range args {
		r[i] = StringSpec(v)
	}
	return r
}

func ElemSpecs(list interface{}) []ElemSpec {
	if list == nil {
		return nil
	}
	v := reflect.ValueOf(list)
	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		panic("no array")
	}
	r := make([]ElemSpec, v.Len())
	for i := 0; i < v.Len(); i++ {
		r[i] = v.Index(i).Interface().(ElemSpec)
	}
	return r
}

func HandleArgs(outputs output.Outputs, opts *output.Options, handler TypeHandler, args ...string) error {
	return HandleOutputs(outputs, opts, handler, StringElemSpecs(args...)...)
}

func HandleOutputs(outputs output.Outputs, opts *output.Options, handler TypeHandler, args ...ElemSpec) error {
	if err := opts.Complete(); err != nil {
		return err
	}
	output, err := outputs.Create(opts)
	if err != nil {
		return err
	}
	return HandleOutput(output, handler, args...)
}

func HandleOutput(output output.Output, handler TypeHandler, specs ...ElemSpec) error {
	if len(specs) == 0 {
		result, err := handler.All()
		if err != nil {
			return err
		}
		if result == nil {
			fmt.Fprintf(os.Stderr, "not supported by source")
			return nil
		}
		for _, r := range result {
			output.Add(nil, r)
		}
	}
	for _, s := range specs {
		result, err := handler.Get(s)
		if err != nil {
			return errors.Wrapf(err, "error processing %q", s.String())
		}
		for _, r := range result {
			output.Add(nil, r)
		}
	}
	err := output.Close(nil)
	if err != nil {
		return err
	}
	return output.Out(nil)
}
