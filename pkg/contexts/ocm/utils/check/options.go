// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package check

import (
	"github.com/open-component-model/ocm/pkg/optionutils"
	"github.com/open-component-model/ocm/pkg/utils"
)

type Option = optionutils.Option[*Options]

type Options struct {
	CheckLocalResources *bool
	CheckLocalSources   *bool
}

var _ Option = (*Options)(nil)

func (o *Options) ApplyTo(opts *Options) {
	optionutils.ApplyOption(o.CheckLocalResources, &opts.CheckLocalResources)
	optionutils.ApplyOption(o.CheckLocalSources, &opts.CheckLocalSources)
}

////////////////////////////////////////////////////////////////////////////////

type localSources bool

func LocalSourcesOnly(b ...bool) Option {
	return localSources(utils.OptionalDefaultedBool(true, b...))
}

func (l localSources) ApplyTo(t *Options) {
	t.CheckLocalSources = optionutils.PointerTo(bool(l))
}

////////////////////////////////////////////////////////////////////////////////

type localResources bool

func LocalResourcesOnly(b ...bool) Option {
	return localResources(utils.OptionalDefaultedBool(true, b...))
}

func (l localResources) ApplyTo(t *Options) {
	t.CheckLocalResources = optionutils.PointerTo(bool(l))
}
