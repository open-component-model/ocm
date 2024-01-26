package ociartifact

import (
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/transfer/filters"
	"github.com/open-component-model/ocm/pkg/optionutils"
)

type Option = optionutils.Option[*Options]

type Filter = filters.Filter

type Options struct {
	Context oci.Context
	Version string
	Filter  Filter
	Printer common.Printer
}

func (o *Options) OCIContext() oci.Context {
	if o.Context == nil {
		return oci.DefaultContext()
	}
	return o.Context
}

func (o *Options) GetPrinter() common.Printer {
	if o.Printer == nil {
		return common.NewPrinter(nil)
	}
	return o.Printer
}

func (o *Options) Printf(msg string, args ...interface{}) {
	if o.Printer != nil {
		o.Printer.Printf(msg, args...)
	}
}

func (o *Options) ApplyTo(opts *Options) {
	if opts == nil {
		return
	}
	if o.Context != nil {
		opts.Context = o.Context
	}
	if o.Version != "" {
		opts.Version = o.Version
	}
	if o.Printer != nil {
		opts.Printer = o.Printer
	}
	if o.Filter != nil {
		opts.Filter = o.Filter
	}
}

////////////////////////////////////////////////////////////////////////////////

type context struct {
	oci.Context
}

func (o context) ApplyTo(opts *Options) {
	opts.Context = o
}

func WithContext(ctx oci.ContextProvider) Option {
	return context{ctx.OCIContext()}
}

////////////////////////////////////////////////////////////////////////////////

type version string

func (o version) ApplyTo(opts *Options) {
	opts.Version = string(o)
}

func WithVersion(v string) Option {
	return version(v)
}

////////////////////////////////////////////////////////////////////////////////

type printer struct {
	common.Printer
}

func (o printer) ApplyTo(opts *Options) {
	opts.Printer = o
}

func WithPrinter(p common.Printer) Option {
	return printer{p}
}

////////////////////////////////////////////////////////////////////////////////

type _filter struct {
	filters.Filter
}

func (o _filter) ApplyTo(opts *Options) {
	opts.Filter = o.Filter
}

func WithFilter(f filters.Filter) Option {
	return _filter{f}
}
