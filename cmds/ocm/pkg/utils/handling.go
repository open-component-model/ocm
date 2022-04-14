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

	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/pkg/errors"
)

type ElemSpec interface {
	String() string
}

type StringSpec string

func (s StringSpec) String() string {
	return string(s)
}

// TypeHandler provides base input to an output processing chain
// using HandleArsg or HandleOutput(s).
// It provides the exploding of intials specifications
// to effective objects passed to the output processing chain.
type TypeHandler interface {
	// All returns all elements according to its context
	All() ([]output.Object, error)
	// Get returns the the elements for a dedicated specification
	// according to the handlers context.
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

func HandleArgs(opts *output.Options, handler TypeHandler, args ...string) error {
	return HandleOutputs(opts, handler, StringElemSpecs(args...)...)
}

func HandleOutputs(opts *output.Options, handler TypeHandler, args ...ElemSpec) error {
	return HandleOutput(opts.Output, handler, args...)
}

func HandleOutput(output output.Output, handler TypeHandler, specs ...ElemSpec) error {
	if len(specs) == 0 {
		result, err := handler.All()
		if err != nil {
			return err
		}
		if result == nil {
			return fmt.Errorf("all mode not support")
		}
		for _, r := range result {
			err := output.Add(r)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			}
		}
	}
	for _, s := range specs {
		result, err := handler.Get(s)
		if err != nil {
			return errors.Wrapf(err, "error processing %q", s.String())
		}
		for _, r := range result {
			err := output.Add(r)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			}
		}
	}
	err := output.Close()
	if err != nil {
		return err
	}
	return output.Out()
}
