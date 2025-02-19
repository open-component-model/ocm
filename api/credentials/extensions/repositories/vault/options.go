package vault

import (
	"slices"

	"github.com/mandelsoft/goutils/optionutils"

	"ocm.software/ocm/api/utils"
)

type Option = optionutils.Option[*Options]

type Options struct {
	Namespace                string   `json:"namespace,omitempty"`
	MountPath                string   `json:"mountPath,omitempty"`
	Path                     string   `json:"path,omitempty"`
	Secrets                  []string `json:"secrets,omitempty"`
	PropgateConsumerIdentity bool     `json:"propagateConsumerIdentity,omitempty"`
}

var _ Option = (*Options)(nil)

func (o *Options) ApplyTo(opts *Options) {
	if o.Namespace != "" {
		opts.Namespace = o.Namespace
	}
	if o.MountPath != "" {
		opts.MountPath = o.MountPath
	}
	if o.Path != "" {
		opts.Path = o.Path
	}
	if o.Secrets != nil {
		opts.Secrets = slices.Clone(o.Secrets)
	}
	opts.PropgateConsumerIdentity = o.PropgateConsumerIdentity
}

////////////////////////////////////////////////////////////////////////////////

type ns string

func (o ns) ApplyTo(opts *Options) {
	opts.Namespace = string(o)
}

func WithNamespace(s string) Option {
	return ns(s)
}

////////////////////////////////////////////////////////////////////////////////

type m string

func (o m) ApplyTo(opts *Options) {
	opts.MountPath = string(o)
}

func WithMountPath(s string) Option {
	return m(s)
}

////////////////////////////////////////////////////////////////////////////////

type p string

func (o p) ApplyTo(opts *Options) {
	opts.Path = string(o)
}

func WithPath(s string) Option {
	return p(s)
}

////////////////////////////////////////////////////////////////////////////////

type sec []string

func (o sec) ApplyTo(opts *Options) {
	opts.Secrets = append(opts.Secrets, []string(o)...)
}

func WithSecrets(s ...string) Option {
	return sec(slices.Clone(s))
}

////////////////////////////////////////////////////////////////////////////////

type pr bool

func (o pr) ApplyTo(opts *Options) {
	opts.PropgateConsumerIdentity = bool(o)
}

func WithPropagation(b ...bool) Option {
	return pr(utils.OptionalDefaultedBool(true, b...))
}
