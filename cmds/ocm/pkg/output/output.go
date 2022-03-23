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

package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gardener/ocm/cmds/ocm/pkg/data"
	"sigs.k8s.io/yaml"
)

type Object interface{}

////////////////////////////////////////////////////////////////////////////////

// Output handles the output of elements.
// It consists of two phases:
// First, elements are added to the output using the Add method,
// This phase is finished calling the Close method. THis finalizes
// any ongoing input processing.
// Second, the final output is requested using the Out method.
type Output interface {
	Add(processingContext interface{}, e interface{}) error
	Close(processingContext interface{}) error
	Out(processingContext interface{}) error
}

////////////////////////////////////////////////////////////////////////////////

type NopOutput struct{}

var _ Output = (*NopOutput)(nil)

func (NopOutput) Add(processingContext interface{}, e interface{}) error {
	return nil
}

func (NopOutput) Close(processingContext interface{}) error {
	return nil
}

func (n NopOutput) Out(processingContext interface{}) error {
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type Manifest interface {
	AsManifest() interface{}
}

type ManifestOutput struct {
	data []Object
}

func (this *ManifestOutput) Add(processingContext interface{}, e interface{}) error {
	this.data = append(this.data, e)
	return nil
}

func (this *ManifestOutput) Close(processingContext interface{}) error {
	return nil
}

type YAMLOutput struct {
	ManifestOutput
}

func (this *YAMLOutput) Out(processingContext interface{}) error {
	for _, m := range this.data {
		fmt.Println("---")
		d, err := yaml.Marshal(m.(Manifest).AsManifest())
		if err != nil {
			return err
		}
		os.Stdout.Write(d)
	}
	return nil
}

type YAMLProcessingOutput struct {
	ElementOutput
}

var _ Output = &YAMLProcessingOutput{}

func NewProcessingYAMLOutput(chain data.ProcessChain) *YAMLProcessingOutput {
	return (&YAMLProcessingOutput{}).new(chain)
}

func (this *YAMLProcessingOutput) new(chain data.ProcessChain) *YAMLProcessingOutput {
	this.ElementOutput.new(chain)
	return this
}

func (this *YAMLProcessingOutput) Out(interface{}) error {
	i := this.Elems.Iterator()
	for i.HasNext() {
		fmt.Printf("---\n")
		elem := i.Next()
		if m, ok := elem.(Manifest); ok {
			elem = m.AsManifest()
		}
		d, err := yaml.Marshal(elem)
		if err != nil {
			return err
		}
		os.Stdout.Write(d)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type JSONOutput struct {
	ManifestOutput
	pretty bool
}

type ItemList struct {
	Items []interface{} `json:"items"`
}

func (this *JSONOutput) Out(interface{}) error {
	items := &ItemList{}
	for _, m := range this.data {
		items.Items = append(items.Items, m.(Manifest).AsManifest())
	}
	d, err := json.Marshal(items)
	if err != nil {
		return err
	}
	if this.pretty {
		var buf bytes.Buffer
		err = json.Indent(&buf, d, "", "  ")
		if err != nil {
			return err
		}
		buf.WriteByte('\n')
		d = buf.Bytes()
	}
	os.Stdout.Write(d)
	return nil
}

type JSONProcessingOutput struct {
	ElementOutput
	pretty bool
}

var _ Output = &JSONProcessingOutput{}

func NewProcessingJSONOutput(chain data.ProcessChain, pretty bool) *JSONProcessingOutput {
	return (&JSONProcessingOutput{}).new(chain, pretty)
}

func (this *JSONProcessingOutput) new(chain data.ProcessChain, pretty bool) *JSONProcessingOutput {
	this.ElementOutput.new(chain)
	this.pretty = pretty
	return this
}

func (this *JSONProcessingOutput) Out(interface{}) error {
	items := &ItemList{}
	i := this.Elems.Iterator()
	for i.HasNext() {
		elem := i.Next()
		if m, ok := elem.(Manifest); ok {
			elem = m.AsManifest()
		}
		items.Items = append(items.Items, elem)
	}
	d, err := json.Marshal(items)
	if err != nil {
		return err
	}
	if this.pretty {
		var buf bytes.Buffer
		err = json.Indent(&buf, d, "", "  ")
		if err != nil {
			return err
		}
		buf.WriteByte('\n')
		d = buf.Bytes()
	}
	os.Stdout.Write(d)
	return nil
}

////////////////////////////////////////////////////////////////////////////

type OutputFactory func(*Options) Output

type Outputs map[string]OutputFactory

func NewOutputs(def OutputFactory, others ...Outputs) Outputs {
	o := Outputs{"": def}
	for _, other := range others {
		for k, v := range other {
			o[k] = v
		}
	}
	return o
}

func (this Outputs) Select(name string) OutputFactory {
	c, ok := this[name]
	if !ok {
		keys := []string{}
		for k, _ := range this {
			keys = append(keys, k)
		}
		k := SelectBest(name, keys...)
		if k != "" {
			c = this[k]
		}
	}
	return c
}

func (this Outputs) Create(opts *Options) (Output, error) {
	f := opts.Output
	if f == nil {
		return this[""](opts), nil
	}
	c := this.Select(*f)
	if c != nil {
		o := c(opts)
		if o != nil {
			return o, nil
		}
	}
	return nil, fmt.Errorf("invalid output format '%s'", *f)
}

func (this Outputs) AddManifestOutputs() Outputs {
	this["yaml"] = func(opts *Options) Output {
		return &YAMLOutput{ManifestOutput{data: []Object{}}}
	}
	this["json"] = func(opts *Options) Output {
		return &JSONOutput{ManifestOutput{data: []Object{}}, true}
	}
	this["JSON"] = func(opts *Options) Output {
		return &JSONOutput{ManifestOutput{data: []Object{}}, false}
	}
	return this
}

func (this Outputs) AddChainedManifestOutputs(chain func(opts *Options) data.ProcessChain) Outputs {
	this["yaml"] = func(opts *Options) Output {
		return NewProcessingYAMLOutput(chain(opts))
	}
	this["json"] = func(opts *Options) Output {
		return NewProcessingJSONOutput(chain(opts), true)
	}
	this["JSON"] = func(opts *Options) Output {
		return NewProcessingJSONOutput(chain(opts), false)
	}
	return this
}

func GetOutput(opts *Options, def Output) (Output, error) {
	o := def
	f := opts.Output
	if f != nil {
		switch *f {
		case "yaml":
			o = &YAMLOutput{ManifestOutput{data: []Object{}}}
		case "json":
			o = &JSONOutput{ManifestOutput{data: []Object{}}, true}
		case "JSON":
			o = &JSONOutput{ManifestOutput{data: []Object{}}, false}
		default:
			return nil, fmt.Errorf("invalid output format '%s'", *f)
		}
	}
	return o, nil
}
