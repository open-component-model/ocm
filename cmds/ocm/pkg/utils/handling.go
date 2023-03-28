// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"os"
	"reflect"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/utils/panics"
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
	// Get returns the elements for a dedicated specification
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
	// This is acceptable as the calling chain will just take a nil slice.
	defer panics.HandlePanic()

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
			return fmt.Errorf("all mode not supported")
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
