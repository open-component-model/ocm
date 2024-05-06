package output

import (
	"github.com/mandelsoft/logging"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/data"
	. "github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	. "github.com/open-component-model/ocm/pkg/out"
)

type ElementOutput struct {
	source  ProcessingSource
	Elems   data.Iterable
	Context Context
	Status  error
}

var _ Output = (*ElementOutput)(nil)

func NewElementOutput(opts *Options, chain ProcessChain) *ElementOutput {
	return (&ElementOutput{}).new(opts, chain)
}

func (this *ElementOutput) new(opts *Options, chain ProcessChain) *ElementOutput {
	log := opts.LogContext()
	if log == nil {
		log = logging.DefaultContext()
	}
	this.source = NewIncrementalProcessingSource(log)
	this.Context = opts.Context
	chain = opts.AdaptChain(&this.Status, chain)
	if chain == nil {
		this.Elems = this.source
	} else {
		this.Elems = Process(log, this.source).Asynchronously().Apply(chain)
	}
	return this
}

func (this *ElementOutput) Add(e interface{}) error {
	this.source.Add(e)
	return nil
}

func (this *ElementOutput) Close() error {
	this.source.Close()
	return nil
}

func (this *ElementOutput) Out() error {
	return this.Status
}
