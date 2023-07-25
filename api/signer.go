package api

import (
	"github.com/open-component-model/ocm/v2/pkg/common"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/v2/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/signing"
)

// Signer can sign objects using specific algorithm or sensible defaults.
type Signer interface {
	Sign(ctx ocm.Context, componentVersion ocm.ComponentVersionAccess, opts ...SignOptionFunc) (*metav1.DigestSpec, error)
}

// SignOptions defines high level options that can be configured for signing a component.
type SignOptions struct {
	state         signing.WalkingState
	printer       common.Printer
	signerOptions *signing.Options
}

type SignOptionFunc func(opts *SignOptions)

// WithPrinter allows for passing in a printer option.
func WithPrinter(printer common.Printer) SignOptionFunc {
	return func(opts *SignOptions) {
		opts.printer = printer
	}
}

// WithState allows fine-tuning the walking state. In reality, the default is sufficient in most cases.
func WithState(state signing.WalkingState) SignOptionFunc {
	return func(opts *SignOptions) {
		opts.state = state
	}
}

// WithSignerOptions allows for fine-tuning the signing options.
func WithSignerOptions(sopts *signing.Options) SignOptionFunc {
	return func(opts *SignOptions) {
		opts.signerOptions = sopts
	}
}

// ComponentSigner can sign a component version.
type ComponentSigner struct{}

// Sign takes a context and a component and sign it using some default values and various options passed in.
func (c *ComponentSigner) Sign(ctx ocm.Context, componentVersion ocm.ComponentVersionAccess, opts ...SignOptionFunc) (*metav1.DigestSpec, error) {
	defaults := &SignOptions{
		// these are more of less constant
		state:   signing.NewWalkingState(ctx.LoggingContext().WithContext(signing.REALM)),
		printer: common.NewPrinter(nil),
	}

	for _, o := range opts {
		o(defaults)
	}

	return signing.Apply(defaults.printer, &defaults.state, componentVersion, defaults.signerOptions, true)
}
