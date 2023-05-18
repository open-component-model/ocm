// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package download

import (
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
)

func From(o *output.Options) *Option {
	var opt *Option
	o.Get(&opt)
	return opt
}

func NewOptions() *Option {
	return &Option{}
}

type Option struct {
	UseHandlers bool
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.UseHandlers, "download-handlers", "d", false, "use download handler if possible")

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
