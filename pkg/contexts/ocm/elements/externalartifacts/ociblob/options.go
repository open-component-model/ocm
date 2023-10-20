// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"github.com/open-component-model/ocm/pkg/optionutils"
)

type Option = optionutils.Option[*Options]

type Options struct {
	MediaType string
}

func (o *Options) Apply(opts *Options) {
	if o.MediaType != "" {
		opts.MediaType = o.MediaType
	}
}

////////////////////////////////////////////////////////////////////////////////
// Local options

type mediatype string

func (h mediatype) ApplyTo(opts *Options) {
	opts.MediaType = string(h)
}

func WithMediaType(h string) Option {
	return mediatype((h))
}
