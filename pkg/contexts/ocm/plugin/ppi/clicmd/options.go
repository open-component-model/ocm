package clicmd

import (
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/optionutils"
)

type Options struct {
	RequireCLIConfig *bool
	Verb             string
	Realm            string
}

type Option = optionutils.Option[*Options]

////////////////////////////////////////////////////////////////////////////////

func (o *Options) ApplyTo(opts *Options) {
	if opts == nil {
		return
	}
	if o.Verb != "" {
		opts.Verb = o.Verb
	}
	if o.Realm != "" {
		opts.Realm = o.Realm
	}
	if o.RequireCLIConfig != nil {
		opts.RequireCLIConfig = o.RequireCLIConfig
	}
}

////////////////////////////////////////////////////////////////////////////////

type verb string

func (o verb) ApplyTo(opts *Options) {
	opts.Verb = string(o)
}

func WithVerb(v string) Option {
	return verb(v)
}

////////////////////////////////////////////////////////////////////////////////

type realm string

func (o realm) ApplyTo(opts *Options) {
	opts.Realm = string(o)
}

func WithRealm(r string) Option {
	return realm(r)
}

////////////////////////////////////////////////////////////////////////////////

type cliconfig bool

func (o cliconfig) ApplyTo(opts *Options) {
	opts.RequireCLIConfig = optionutils.BoolP(o)
}

func WithCLIConfig(r ...bool) Option {
	return cliconfig(general.OptionalDefaultedBool(true, r...))
}
