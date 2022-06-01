// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package signoption

import (
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	ocmsign "github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
	"github.com/spf13/pflag"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

var _ options.Options = (*Option)(nil)

func New(sign bool) *Option {
	return &Option{SignMode: sign}
}

type Option struct {
	SignMode    bool
	algorithm   string
	publicKeys  []string
	privateKeys []string

	// Verify the digestes
	Verify bool
	// Signature name
	Signature string
	Update    bool
	Signer    signing.Signer
	Keys      signing.KeyRegistry
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.StringArrayVarP(&o.publicKeys, "public-key", "k", nil, "public key setting")
	fs.StringArrayVarP(&o.privateKeys, "private-key", "K", nil, "private key setting")
	fs.StringVarP(&o.Signature, "signature", "s", "", "signature name")
	fs.StringVarP(&o.algorithm, "algorithm", "", "", "signature handler")
	fs.BoolVarP(&o.Verify, "verify", "V", true, "verify existing digests")
	fs.BoolVarP(&o.Update, "update", "", o.SignMode, "update digest in component versions")
}

func (o *Option) Complete(ctx clictx.Context) error {
	if o.Keys == nil {
		o.Keys = signing.NewKeyRegistry()
	}
	if o.SignMode {
		if o.algorithm == "" {
			o.algorithm = rsa.Algorithm
		}
		o.Signer = signingattr.Get(ctx).GetSigner(o.algorithm)
		if o.Signer == nil {
			return errors.ErrUnknown(compdesc.KIND_SIGN_ALGORITHM, o.algorithm)
		}
	}
	err := o.handleKeys(ctx, "public key", o.publicKeys, o.Keys.RegisterPublicKey)
	if err != nil {
		return err
	}
	err = o.handleKeys(ctx, "private key", o.privateKeys, o.Keys.RegisterPrivateKey)
	if err != nil {
		return err
	}
	return nil
}

func (o *Option) handleKeys(ctx clictx.Context, desc string, keys []string, add func(string, interface{})) error {
	for i, k := range keys {
		name := o.Signature
		file := k
		sep := strings.Index(k, "=")
		if sep >= 0 {
			name = k[:sep]
			file = k[i+1:]
		}
		data, err := vfs.ReadFile(ctx.FileSystem(), file)
		if err != nil {
			return errors.Wrapf(err, "cannot read %s file %q", desc, file)
		}
		if name == "" {
			return errors.Newf("signature name required")
		}
		add(name, data)
	}
	return nil
}

func (o *Option) Usage() string {
	s := `
The <code>--public-key</code> and <code>--private-key</code> options can be
used to define public and private keys on the command line. The options have an
argument of the form <code>[&lt;name>=]&lt;filepath></code>. The optional name
specifies the signature name the key should be used for. By default this is the
signature name specified with the option <code>--signature</code>.
`
	return s
}

var _ ocmsign.Option = (*Option)(nil)

func (o *Option) ApplySigningOption(opts *ocmsign.Options) {
	if o.Signer != nil {
		opts.Signer = o.Signer
	}
	opts.SignatureName = o.Signature
	opts.Verify = o.Verify
	opts.Keys = o.Keys
	opts.Update = o.Update
}
