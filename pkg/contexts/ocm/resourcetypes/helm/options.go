// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common"
	base "github.com/open-component-model/ocm/pkg/common/accessio/blobaccess/helm"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes/rpi"
	"github.com/open-component-model/ocm/pkg/optionutils"
)

type Option = optionutils.Option[*Options]

type Options struct {
	rpi.Options
	Blob base.Options
}

var _ rpi.GeneralOptionsProvider = (*Options)(nil)

func (o *Options) Apply(opts *Options) {
	o.Options.ApplyTo(&opts.Options)
	o.Blob.ApplyTo(&opts.Blob)
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
// DirTree BlobAccess Options

func mapBaseOption(opts *Options) *base.Options {
	return &opts.Blob
}

func wrapBase(o base.Option) Option {
	return optionutils.OptionWrapperFunc[*base.Options, *Options](o, mapBaseOption)
}

func WithFileSystem(fs vfs.FileSystem) Option {
	return wrapBase(base.WithFileSystem(fs))
}

func WithContext(ctx oci.ContextProvider) Option {
	return wrapBase(base.WithContext(ctx))
}

func WithIVersion(v string) Option {
	return wrapBase(base.WithVersion(v))
}

func WithIVersionOverride(v string, flag ...bool) Option {
	return wrapBase(base.WithVersionOverride(v, flag...))
}

func WithCACert(v string) Option {
	return wrapBase(base.WithCACert(v))
}

func WithCACertFile(v string) Option {
	return wrapBase(base.WithCACertFile(v))
}

func WithHelmRepository(v string) Option {
	return wrapBase(base.WithHelmRepository(v))
}

func WithPrinter(v common.Printer) Option {
	return wrapBase(base.WithPrinter(v))
}
