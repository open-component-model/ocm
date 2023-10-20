// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"github.com/open-component-model/ocm/pkg/optionutils"
)

type Option = optionutils.Option[*Options]

type Options struct {
	Region    string
	Version   string
	MediaType string
}

func (o *Options) Apply(opts *Options) {
	if o.Region != "" {
		opts.Region = o.Region
	}
	if o.Version != "" {
		opts.Version = o.Version
	}
	if o.MediaType != "" {
		opts.MediaType = o.MediaType
	}
}

////////////////////////////////////////////////////////////////////////////////
// Local options

type region string

func (h region) ApplyTo(opts *Options) {
	opts.Region = string(h)
}

func WithRegion(h string) Option {
	return region((h))
}

////////////////////////////////////////////////////////////////////////////////

type version string

func (h version) ApplyTo(opts *Options) {
	opts.Version = string(h)
}

func WithVersion(h string) Option {
	return version((h))
}

////////////////////////////////////////////////////////////////////////////////

type mediatype string

func (h mediatype) ApplyTo(opts *Options) {
	opts.MediaType = string(h)
}

func WithMediaType(h string) Option {
	return mediatype((h))
}
