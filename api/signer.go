package api

import (
	"github.com/open-component-model/ocm/v2/pkg/common"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/v2/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/signing"
)

// Signer can sign objects using specific algorithm or sensible defaults.
type Signer interface {
	Sign(componentVersion ocm.ComponentVersionAccess, opts ...OptionFunc) (*metav1.DigestSpec, error)
}

type Verifier interface {
	Verify(componentVersion ocm.ComponentVersionAccess, opts ...OptionFunc) (*metav1.DigestSpec, error)
}

type SigningVerifier interface {
	Signer
	Verifier
}

// Options defines high level options that can be configured for signing a component.
type Options struct {
	state          signing.WalkingState
	printer        common.Printer
	signingOptions *signing.Options
}

type OptionFunc func(opts *Options)

// WithPrinter allows for passing in a printer option.
func WithPrinter(printer common.Printer) OptionFunc {
	return func(opts *Options) {
		opts.printer = printer
	}
}

// WithState allows fine-tuning the walking state. In reality, the default is sufficient in most cases.
func WithState(state signing.WalkingState) OptionFunc {
	return func(opts *Options) {
		opts.state = state
	}
}

// WithSignerOptions allows for fine-tuning the signing options.
func WithSignerOptions(sopts *signing.Options) OptionFunc {
	return func(opts *Options) {
		opts.signingOptions = sopts
	}
}

// ComponentSigningVerifier can sign a component version.
type ComponentSigningVerifier struct{}

// Sign takes a context and a component and sign it using some default values and various options passed in.
func (c *ComponentSigningVerifier) Sign(componentVersion ocm.ComponentVersionAccess, opts ...OptionFunc) (*metav1.DigestSpec, error) {
	ctx := componentVersion.GetContext()
	defaults := &Options{
		// these are more of less constant
		state:   signing.NewWalkingState(ctx.LoggingContext().WithContext(signing.REALM)),
		printer: common.NewPrinter(nil),
	}

	for _, o := range opts {
		o(defaults)
	}

	return signing.Apply(defaults.printer, &defaults.state, componentVersion, defaults.signingOptions, true)
}

// Verify takes a context and a component and verifies its signature.
func (c *ComponentSigningVerifier) Verify(componentVersion ocm.ComponentVersionAccess, opts ...OptionFunc) (*metav1.DigestSpec, error) {
	ctx := componentVersion.GetContext()
	defaults := &Options{
		// these are more of less constant
		state:   signing.NewWalkingState(ctx.LoggingContext().WithContext(signing.REALM)),
		printer: common.NewPrinter(nil),
	}

	for _, o := range opts {
		o(defaults)
	}

	return signing.Apply(defaults.printer, &defaults.state, componentVersion, defaults.signingOptions, true)
}
