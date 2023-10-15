// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes/rpi"
)

type Option = rpi.ResourceOption[*Options]

type Options struct {
	rpi.Options
	MimeType string
}

var _ rpi.GeneralOptionsProvider = (*Options)(nil)

func (o *Options) Apply(opts *Options) {
	o.Options.ApplyTo(&opts.Options)
	if o.MimeType != "" {
		opts.MimeType = o.MimeType
	}
}

////////////////////////////////////////////////////////////////////////////////
// General Options

func WithHint(h string) Option {
	return rpi.WrapHint[Options](h)
}

func WithGlobalAccess(a cpi.AccessSpec) Option {
	return rpi.WrapGlobalAccess[Options](a)
}

////////////////////////////////////////////////////////////////////////////////
// Local Options

type mimetype struct {
	mime string
}

func (o mimetype) ApplyTo(opts *Options) {
	opts.MimeType = o.mime
}

func WithMimeType(mime string) Option {
	return mimetype{mime}
}
