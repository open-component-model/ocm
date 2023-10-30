// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package vault

import (
	"golang.org/x/exp/slices"

	"github.com/open-component-model/ocm/pkg/utils"
)

type Option interface {
	ApplyTo(*Options)
}

type Options struct {
	Namespace                string   `json:"namespace"`
	SearchEngine             string   `json:"searchEngine,omitempty"`
	Path                     string   `json:"path,omitempty"`
	Secrets                  []string `json:"secrets"`
	PropgateConsumerIdentity bool     `json:"propagateConsumerIdentity,omitempty"`
}

type ns string

func (o ns) ApplyTo(opts *Options) {
	opts.Namespace = string(o)
}

func WithNamespace(s string) Option {
	return ns(s)
}

////////////////////////////////////////////////////////////////////////////////

type se string

func (o se) ApplyTo(opts *Options) {
	opts.SearchEngine = string(o)
}

func WithSearchEngine(s string) Option {
	return se(s)
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
