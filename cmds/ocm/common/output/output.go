package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	. "ocm.software/ocm/api/utils/out"

	"github.com/mandelsoft/goutils/errors"
	"sigs.k8s.io/yaml"

	"ocm.software/ocm/cmds/ocm/common/processing"
)

type Object = interface{}

type Objects = []Object

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

// Destination is an optional interface for outputs to
// set the payload output stream to use.
type Destination interface {
	SetDestination(io.Writer)
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

type manifest struct {
	data interface{}
}

func (m *manifest) AsManifest() interface{} {
	return m.data
}

func AsManifest(i interface{}) Manifest {
	return &manifest{i}
}

type ManifestOutput struct {
	DestinationOutput
	opts   *Options
	data   []Object
	Status error
}

func NewManifestOutput(opts *Options) ManifestOutput {
	return ManifestOutput{
		DestinationOutput: DestinationOutput{
			Context: opts.Context,
		},
		opts: opts,
		data: []Object{},
	}
}

func (this *ManifestOutput) Add(e interface{}) error {
	this.data = append(this.data, e)
	if this.opts.StatusCheck != nil {
		this.Status = this.opts.StatusCheck(this.opts, e, this.Status)
	}
	return nil
}

func (this *ManifestOutput) Close() error {
	return nil
}

func (this *ManifestOutput) Out() error {
	return this.Status
}

type YAMLOutput struct {
	ManifestOutput
}

func (this *YAMLOutput) Out() error {
	for _, m := range this.data {
		this.Print("---\n")
		d, err := yaml.Marshal(m.(Manifest).AsManifest())
		if err != nil {
			return err
		}
		this.Write(d)
	}
	return this.ManifestOutput.Out()
}

type YAMLProcessingOutput struct {
	ElementOutput
}

var _ Output = &YAMLProcessingOutput{}

func NewProcessingYAMLOutput(opts *Options, chain processing.ProcessChain) *YAMLProcessingOutput {
	return (&YAMLProcessingOutput{}).new(opts, chain)
}

func (this *YAMLProcessingOutput) new(opts *Options, chain processing.ProcessChain) *YAMLProcessingOutput {
	this.ElementOutput.new(opts, chain)
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
	return this.ElementOutput.Out()
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
	this.Write(d)
	return this.ManifestOutput.Out()
}

type JSONProcessingOutput struct {
	ElementOutput
	pretty bool
}

var _ Output = &JSONProcessingOutput{}

func NewProcessingJSONOutput(opts *Options, chain processing.ProcessChain, pretty bool) *JSONProcessingOutput {
	return (&JSONProcessingOutput{}).new(opts, chain, pretty)
}

func (this *JSONProcessingOutput) new(opts *Options, chain processing.ProcessChain, pretty bool) *JSONProcessingOutput {
	this.ElementOutput.new(opts, chain)
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
	return this.ElementOutput.Out()
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
		return &YAMLOutput{NewManifestOutput(opts)}
	}
	this["json"] = func(opts *Options) Output {
		return &JSONOutput{NewManifestOutput(opts), true}
	}
	this["JSON"] = func(opts *Options) Output {
		return &JSONOutput{NewManifestOutput(opts), false}
	}
	return this
}

func (this Outputs) AddChainedManifestOutputs(chain ChainFunction) Outputs {
	this["yaml"] = func(opts *Options) Output {
		return NewProcessingYAMLOutput(opts, chain(opts))
	}
	this["json"] = func(opts *Options) Output {
		return NewProcessingJSONOutput(opts, chain(opts), true)
	}
	this["JSON"] = func(opts *Options) Output {
		return NewProcessingJSONOutput(opts, chain(opts), false)
	}
	return this
}

func DefaultYAMLOutput(opts *Options) Output {
	return &YAMLOutput{NewManifestOutput(opts)}
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
