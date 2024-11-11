package addhdlrs

import (
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/ocm"
)

type Options struct {
	Replace bool
}

var _ ocm.ModificationOption = (*Options)(nil)

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	f := fs.Lookup("replace")
	if f != nil {
		if f.Value.Type() == "bool" {
			return
		}
	}
	fs.BoolVarP(&o.Replace, "replace", "R", false, "replace existing elements")
}

func (o *Options) ApplyBlobModificationOption(opts *ocm.BlobModificationOptions) {
	o.ApplyTargetOption(&opts.TargetOptions)
}

func (o *Options) ApplyModificationOption(opts *ocm.ModificationOptions) {
	o.ApplyTargetOption(&opts.TargetOptions)
}

func (o *Options) ApplyTargetOption(opts *ocm.TargetOptions) {
	if !o.Replace {
		opts.TargetElement = ocm.AppendElement
	}
}

func (o *Options) Description() string {
	return `
The <code>--replace</code> option allows users to specify whether adding an
element with the same name and extra identity but different version as an 
existing element append (false) or replace (true) the existing element.
`
}
