// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"github.com/open-component-model/ocm/pkg/optionutils"
)

type Option = optionutils.Option[*Options]

type Options struct {
	APIHostName string
}

func (o *Options) Apply(opts *Options) {
	if o.APIHostName != "" {
		opts.APIHostName = o.APIHostName
	}
}

////////////////////////////////////////////////////////////////////////////////
// Local options

type apihostname string

func (h apihostname) ApplyTo(opts *Options) {
	opts.APIHostName = string(h)
}

func WithAPIHostName(h string) Option {
	return apihostname((h))
}
