// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	. "github.com/open-component-model/ocm/v2/pkg/out"

	"github.com/mandelsoft/logging"
	"sigs.k8s.io/yaml"

	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/v2/pkg/errors"
)

type Object = interface{}

////////////////////////////////////////////////////////////////////////////////

// Output handles the output of elements.
// It consists of two phases:
// First, elements are added to the output using the Add method,
// This phase is finished calling the Close method. THis finalizes
// any ongoing input processing.
// Second, the final output is requested using the Out method.
type Output interface {
	Add(e interface{}) error
	Close() error
	Out() error
}

////////////////////////////////////////////////////////////////////////////////

type NopOutput struct{}

var _ Output = (*NopOutput)(nil)

func (NopOutput) Add(e interface{}) error {
	return nil
}

func (NopOutput) Close() error {
	return nil
}

func (n NopOutput) Out() error {
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type Manifest interface {
	AsManifest() interface{}
}

type ManifestOutput struct {
	data    []Object
	Context Context
}

func (this *ManifestOutput) Add(e interface{}) error {
	this.data = append(this.data, e)
	return nil
}

func (this *ManifestOutput) Close() error {
	return nil
}

type YAMLOutput struct {
	ManifestOutput
}

func (this *YAMLOutput) Out() error {
	for _, m := range this.data {
		Outf(this.Context, "---\n")
		d, err := yaml.Marshal(m.(Manifest).AsManifest())
		if err != nil {
			return err
		}
		this.Context.StdOut().Write(d)
	}
	return nil
}

type YAMLProcessingOutput struct {
	ElementOutput
}

var _ Output = &YAMLProcessingOutput{}

func NewProcessingYAMLOutput(log logging.Context, ctx Context, chain processing.ProcessChain) *YAMLProcessingOutput {
	return (&YAMLProcessingOutput{}).new(log, ctx, chain)
}

func (this *YAMLProcessingOutput) new(log logging.Context, ctx Context, chain processing.ProcessChain) *YAMLProcessingOutput {
	this.ElementOutput.new(log, ctx, chain)
	return this
}

func (this *YAMLProcessingOutput) Out() error {
	i := this.Elems.Iterator()
	for i.HasNext() {
		Outf(this.Context, "---\n")
		elem := i.Next()
		if m, ok := elem.(Manifest); ok {
			elem = m.AsManifest()
		}
		d, err := yaml.Marshal(elem)
		if err != nil {
			return err
		}
		this.Context.StdOut().Write(d)
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

func (this *JSONOutput) Out() error {
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

func NewProcessingJSONOutput(log logging.Context, ctx Context, chain processing.ProcessChain, pretty bool) *JSONProcessingOutput {
	return (&JSONProcessingOutput{}).new(log, ctx, chain, pretty)
}

func (this *JSONProcessingOutput) new(log logging.Context, ctx Context, chain processing.ProcessChain, pretty bool) *JSONProcessingOutput {
	this.ElementOutput.new(log, ctx, chain)
	this.pretty = pretty
	return this
}

func (this *JSONProcessingOutput) Out() error {
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
	this.Context.StdOut().Write(d)
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
		for k := range this {
			keys = append(keys, k)
		}
		k, _ := SelectBest(name, keys...)
		if k != "" {
			c = this[k]
		}
	}
	return c
}

func (this Outputs) Create(opts *Options) (Output, error) {
	f := opts.OutputMode
	c := this.Select(f)
	if c != nil {
		o := c(opts)
		if o != nil {
			return o, nil
		}
	}
	return nil, errors.Newf("invalid output format '%s'", f)
}

func (this Outputs) AddManifestOutputs() Outputs {
	this["yaml"] = func(opts *Options) Output {
		return &YAMLOutput{ManifestOutput{Context: opts.Context, data: []Object{}}}
	}
	this["json"] = func(opts *Options) Output {
		return &JSONOutput{ManifestOutput{Context: opts.Context, data: []Object{}}, true}
	}
	this["JSON"] = func(opts *Options) Output {
		return &JSONOutput{ManifestOutput{Context: opts.Context, data: []Object{}}, false}
	}
	return this
}

func (this Outputs) AddChainedManifestOutputs(chain ChainFunction) Outputs {
	this["yaml"] = func(opts *Options) Output {
		return NewProcessingYAMLOutput(opts.LogContext(), opts.Context, chain(opts))
	}
	this["json"] = func(opts *Options) Output {
		return NewProcessingJSONOutput(opts.LogContext(), opts.Context, chain(opts), true)
	}
	this["JSON"] = func(opts *Options) Output {
		return NewProcessingJSONOutput(opts.LogContext(), opts.Context, chain(opts), false)
	}
	return this
}

var log bool

func Print(list []Object, msg string, args ...interface{}) {
	if log {
		fmt.Printf(msg+":\n", args...)
		for i, e := range list {
			fmt.Printf("  %3d %s\n", i, e)
		}
	}
}
