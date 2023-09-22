// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package keyoption

import (
	"encoding/base64"
	"strings"

	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	ocmsign "github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/utils"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

var _ options.Options = (*Option)(nil)

func New() *Option {
	return &Option{}
}

type Option struct {
	DefaultName string
	publicKeys  []string
	privateKeys []string
	Keys        signing.KeyRegistry
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.StringArrayVarP(&o.publicKeys, "public-key", "k", nil, "public key setting")
	fs.StringArrayVarP(&o.privateKeys, "private-key", "K", nil, "private key setting")
}

func (o *Option) Configure(ctx clictx.Context) error {
	if o.Keys == nil {
		o.Keys = signing.NewKeyRegistry()
	}
	err := o.HandleKeys(ctx, "public key", o.publicKeys, o.Keys.RegisterPublicKey)
	if err != nil {
		return err
	}
	err = o.HandleKeys(ctx, "private key", o.privateKeys, o.Keys.RegisterPrivateKey)
	if err != nil {
		return err
	}
	return nil
}

func (o *Option) HandleKeys(ctx clictx.Context, desc string, keys []string, add func(string, interface{})) error {
	name := o.DefaultName
	for _, k := range keys {
		file := k
		sep := strings.Index(k, "=")
		if sep > 0 {
			name = k[:sep]
			file = k[sep+1:]
		}
		if len(file) == 0 {
			return errors.Newf("empty file name")
		}
		var data []byte
		var err error
		switch file[0] {
		case '=':
			data = []byte(file[1:])
		case '!':
			data, err = base64.StdEncoding.DecodeString(file[1:])
		case '@':
			data, err = utils.ReadFile(ctx.FileSystem(), file[1:])
		default:
			data, err = utils.ReadFile(ctx.FileSystem(), file)
		}
		if err != nil {
			return errors.Wrapf(err, "cannot read %s file %q", desc, file)
		}
		if name == "" {
			return errors.Newf("key name required")
		}
		add(name, data)
	}
	return nil
}

func Usage() string {
	s := `
The <code>--public-key</code> and <code>--private-key</code> options can be
used to define public and private keys on the command line. The options have an
argument of the form <code>&lt;name>=&lt;filepath></code>. The name is the name
of the key and represents the context is used for (For example the signature
name of a component version)

Alternatively a key can be specified as base64 encoded string if the argument
start with the prefix <code>!</code> or as direct string with the prefix
<code>=</code>.
`
	return s
}

var _ ocmsign.Option = (*Option)(nil)

func (o *Option) ApplySigningOption(opts *ocmsign.Options) {
	opts.Keys = o.Keys
}
