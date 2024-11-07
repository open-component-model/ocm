package addhdlrs

import (
	"github.com/mandelsoft/goutils/generics"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/ocm"
)

type Options struct {
	// Replace enables to replace existing elements (same raw identity) with a different version instead
	// of appending a new element.
	Replace bool
	// PreserveSignature disables the modification of signature relevant information.
	PreserveSignature bool
}

var (
	_ ocm.ModificationOption        = (*Options)(nil)
	_ ocm.ElementModificationOption = (*Options)(nil)
	_ ocm.BlobModificationOption    = (*Options)(nil)
	_ ocm.TargetElementOption       = (*Options)(nil)
)

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	f := fs.Lookup("replace")
	if f != nil {
		if f.Value.Type() == "bool" {
			return
		}
		if f == nil {
			fs.BoolVarP(&o.Replace, "replace", "R", false, "replace existing elements")
		}
	}

	f = fs.Lookup("preserve-signature")
	if f == nil {
		fs.BoolVarP(&o.PreserveSignature, "preserve-signature", "P", false, "preserve existing signatures")
	}
}

func (o *Options) applyPreserve(opts *ocm.ElementModificationOptions) {
	if !o.PreserveSignature {
		opts.ModifyElement = generics.Pointer(true)
	}
}

func (o *Options) ApplyBlobModificationOption(opts *ocm.BlobModificationOptions) {
	o.applyPreserve(&opts.ElementModificationOptions)
	o.ApplyTargetOption(&opts.TargetElementOptions)
}

func (o *Options) ApplyModificationOption(opts *ocm.ModificationOptions) {
	o.applyPreserve(&opts.ElementModificationOptions)
	o.ApplyTargetOption(&opts.TargetElementOptions)
}

func (o *Options) ApplyElementModificationOption(opts *ocm.ElementModificationOptions) {
	o.applyPreserve(opts)
	o.ApplyTargetOption(&opts.TargetElementOptions)
}

func (o *Options) ApplyTargetOption(opts *ocm.TargetElementOptions) {
	if !o.Replace {
		opts.TargetElement = ocm.AppendElement
	}
}

func (o *Options) Description() string {
	return `
The <code>--replace</code> option allows users to specify whether adding an
element with the same name and extra identity but different version as an 
existing element, append (false) or replace (true) the existing element.

The <code>--preserve-signature</code> option prohibits changes of signature 
relevant elements.
`
}
