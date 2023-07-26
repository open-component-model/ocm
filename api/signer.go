package api

import (
	"fmt"

	"github.com/open-component-model/ocm/v2/pkg/common"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/attrs/signingattr"
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

type OptionFunc func(opts *Options) error

// WithPrinter allows for passing in a printer option.
func WithPrinter(printer common.Printer) OptionFunc {
	return func(opts *Options) error {
		opts.printer = printer

		return nil
	}
}

// WithState allows fine-tuning the walking state. In reality, the default is sufficient in most cases.
func WithState(state signing.WalkingState) OptionFunc {
	return func(opts *Options) error {
		opts.state = state

		return nil
	}
}

// WithSignerOptions allows for fine-tuning the signing options.
func WithSignerOptions(ctx ocm.Context, sopts *signing.Options) OptionFunc {
	return func(opts *Options) error {
		attr := signingattr.Get(ctx.OCMContext())
		if err := sopts.Complete(attr); err != nil {
			return fmt.Errorf("failed to complete signing options: %w", err)
		}

		opts.signingOptions = sopts

		return nil
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
		if err := o(defaults); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
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
		if err := o(defaults); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	return signing.Apply(defaults.printer, &defaults.state, componentVersion, defaults.signingOptions, true)
}
