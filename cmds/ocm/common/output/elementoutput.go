package output

import (
	"fmt"
	"io"

	"github.com/mandelsoft/logging"
	. "ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/cmds/ocm/common/data"
	. "ocm.software/ocm/cmds/ocm/common/processing"
)

type DestinationOutput struct {
	Context Context
	out     io.Writer
}

var _ Destination = (*DestinationOutput)(nil)

func (this *DestinationOutput) SetDestination(d io.Writer) {
	this.out = d
}

func (this *DestinationOutput) Printf(msg string, args ...interface{}) {
	if this.out != nil {
		fmt.Fprintf(this.out, msg, args...)
	} else {
		fmt.Fprintf(this.Context.StdOut(), msg, args...)
	}
}

func (this *DestinationOutput) Print(args ...interface{}) {
	if this.out != nil {
		fmt.Fprint(this.out, args...)
	} else {
		fmt.Fprint(this.Context.StdOut(), args...)
	}
}

func (this *DestinationOutput) Write(data []byte) (int, error) {
	if this.out != nil {
		return this.out.Write(data)
	} else {
		return this.Context.StdOut().Write(data)
	}
}

type ElementOutput struct {
	DestinationOutput
	source ProcessingSource
	Elems  data.Iterable
	Status error
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
