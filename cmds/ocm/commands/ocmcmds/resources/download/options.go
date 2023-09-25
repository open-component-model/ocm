// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package download

import (
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/pkg/utils"
)

func From(o *output.Options) *Option {
	var opt *Option
	o.Get(&opt)
	return opt
}

func NewOptions(silent ...bool) *Option {
	return &Option{SilentOption: utils.Optional(silent...)}
}

type Option struct {
	SilentOption bool
	UseHandlers  bool
}

func (o *Option) SetUseHandlers(ok ...bool) *Option {
	o.UseHandlers = utils.OptionalDefaultedBool(true, ok...)
	return o
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	if !o.SilentOption {
		fs.BoolVarP(&o.UseHandlers, "download-handlers", "d", false, "use download handler if possible")
	}
}

func (o *Option) Usage() string {
	s := `
The library supports some downloads with semantics based on resource types. For example a helm chart
can be download directly as helm chart archive, even if stored as OCI artifact.
This is handled by download handler. Their usage can be enabled with the <code>--download-handlers</code>
option. Otherwise the resource as returned by the access method is stored.
`
	return s
}
