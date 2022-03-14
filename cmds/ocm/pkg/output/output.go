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
	"context"
	"encoding/json"
	"fmt"
	"os"

	"sigs.k8s.io/yaml"
)

type Object interface{}

type Manifest interface {
	AsManifest() interface{}
}

type Output interface {
	Add(ctx *context.Context, e interface{}) error
	Close(ctx *context.Context) error
	Out(*context.Context) error
}

type ManifestOutput struct {
	data []Object
}

type YAMLOutput struct {
	ManifestOutput
}

type JSONOutput struct {
	ManifestOutput
	pretty bool
}

func (this *ManifestOutput) Add(ctx *context.Context, e interface{}) error {
	this.data = append(this.data, e)
	return nil
}

func (this *ManifestOutput) Close(ctx *context.Context) error {
	return nil
}

func (this *YAMLOutput) Out(ctx *context.Context) error {
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

type ItemList struct {
	Items []interface{} `json:"items"`
}

func (this *JSONOutput) Out(*context.Context) error {
	items := &ItemList{}
	for _, m := range this.data {
		items.Items = append(items.Items, m.(Manifest).AsManifest())
	}
	d, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	os.Stdout.Write(d)
	if this.pretty {
		fmt.Println()
	}
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
