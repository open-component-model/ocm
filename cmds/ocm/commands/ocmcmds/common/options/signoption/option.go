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
	"crypto/x509"
	"encoding/base64"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/normalizations/jsonv1"
	ocmsign "github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha256"
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
	rootca []string

	local         bool
	SignMode      bool
	signAlgorithm string
	hashAlgorithm string
	publicKeys    []string
	privateKeys   []string
	Issuer        string
	RootCerts     *x509.CertPool
	// Verify the digests
	Verify bool

	// Recursively sign component versions
	Recursively bool
	// SignatureNames is a list of signatures to handle (only the first one
	// will be used for signing
	SignatureNames []string
	Update         bool
	Signer         signing.Signer
	Hasher         signing.Hasher
	Keys           signing.KeyRegistry
	NormAlgorithm  string
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.StringArrayVarP(&o.SignatureNames, "signature", "s", nil, "signature name")
	fs.StringArrayVarP(&o.publicKeys, "public-key", "k", nil, "public key setting")
	if o.SignMode {
		fs.StringArrayVarP(&o.privateKeys, "private-key", "K", nil, "private key setting")
		fs.StringVarP(&o.signAlgorithm, "algorithm", "S", rsa.Algorithm, "signature handler")
		fs.StringVarP(&o.NormAlgorithm, "normalization", "N", jsonv1.Algorithm, "normalization algorithm")
		fs.StringVarP(&o.hashAlgorithm, "hash", "H", sha256.Algorithm, "hash algorithm")
		fs.StringVarP(&o.Issuer, "issuer", "I", "", "issuer name")
		fs.BoolVarP(&o.Update, "update", "", o.SignMode, "update digest in component versions")
		fs.BoolVarP(&o.Recursively, "recursive", "R", false, "recursively sign component versions")
	} else {
		fs.BoolVarP(&o.local, "local", "L", false, "verification based on information found in component versions, only")
	}
	fs.BoolVarP(&o.Verify, "verify", "V", o.SignMode, "verify existing digests")
	fs.StringArrayVarP(&o.rootca, "ca-cert", "", o.rootca, "Additional root certificates")
}

func (o *Option) Complete(ctx clictx.Context) error {
	if len(o.SignatureNames) > 0 {
		for i, n := range o.SignatureNames {
			n = strings.TrimSpace(n)
			o.SignatureNames[i] = n
			if n == "" {
				return errors.Newf("empty signature name (name %d) not possible", i)
			}
		}
	} else {
		o.SignatureNames = nil
	}
	if o.Keys == nil {
		o.Keys = signing.NewKeyRegistry()
	}
	if o.SignMode {
		if o.NormAlgorithm == "" {
			o.NormAlgorithm = jsonv1.Algorithm
		}
		if o.signAlgorithm == "" {
			o.signAlgorithm = rsa.Algorithm
		}
		if o.hashAlgorithm == "" {
			o.hashAlgorithm = sha256.Algorithm
		}
		x := compdesc.Normalizations.Get(o.NormAlgorithm)
		if x == nil {
			return errors.ErrUnknown(compdesc.KIND_NORM_ALGORITHM, o.NormAlgorithm)
		}
		o.Signer = signingattr.Get(ctx).GetSigner(o.signAlgorithm)
		if o.Signer == nil {
			return errors.ErrUnknown(compdesc.KIND_SIGN_ALGORITHM, o.signAlgorithm)
		}
		o.Hasher = signingattr.Get(ctx).GetHasher(o.hashAlgorithm)
		if o.Hasher == nil {
			return errors.ErrUnknown(compdesc.KIND_HASH_ALGORITHM, o.hashAlgorithm)
		}
	} else {
		o.Recursively = !o.local
	}

	err := o.handleKeys(ctx, "public key", o.publicKeys, o.Keys.RegisterPublicKey)
	if err != nil {
		return err
	}
	err = o.handleKeys(ctx, "private key", o.privateKeys, o.Keys.RegisterPrivateKey)
	if err != nil {
		return err
	}

	if len(o.rootca) > 0 {
		pool, err := signing.BaseRootPool()
		if err != nil {
			return err
		}
		for _, r := range o.rootca {
			data, err := vfs.ReadFile(ctx.FileSystem(), r)
			if err != nil {
				return errors.Wrapf(err, "cannot read ca file %q", r)
			}
			ok := pool.AppendCertsFromPEM(data)
			if !ok {
				return errors.Newf("cannot add rot certs from %q", r)
			}
		}
		o.RootCerts = pool
	}
	return nil
}

func (o *Option) handleKeys(ctx clictx.Context, desc string, keys []string, add func(string, interface{})) error {
	for i, k := range keys {
		name := ""
		if len(o.SignatureNames) > 0 {
			name = o.SignatureNames[0]
		}
		file := k
		sep := strings.Index(k, "=")
		if sep >= 0 {
			name = k[:sep]
			file = k[i+1:]
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
		default:
			data, err = vfs.ReadFile(ctx.FileSystem(), file)
		}
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

Alternatively a key can be specified as base64 encoded string if the argument
start with the prefix <code>!</code> or as direct string with the prefix
<code>=</code>.
`
	if o.SignMode {
		s += `
If in signing mode a public key is specified, existing signatures for the
given signature name will be verified, instead of recreated.
`
		s += `

The following signing types are supported with option <code>--algorithm</code>:
` + utils.FormatList(rsa.Algorithm, signing.DefaultRegistry().SignerNames()...)

		s += `

The following normalization modes are supported with option <code>--normalization</code>:
` + utils.FormatList(jsonv1.Algorithm, compdesc.Normalizations.Names()...)

		s += `

The following hash modes are supported with option <code>--hash</code>:
` + utils.FormatList(sha256.Algorithm, signing.DefaultRegistry().HasherNames()...)

		signing.DefaultRegistry().HasherNames()
	}
	return s
}

var _ ocmsign.Option = (*Option)(nil)

func (o *Option) ApplySigningOption(opts *ocmsign.Options) {
	if o.Signer != nil {
		opts.Signer = o.Signer
	}
	opts.SignatureNames = o.SignatureNames
	opts.Verify = o.Verify
	opts.Recursively = o.Recursively
	opts.Keys = o.Keys
	opts.NormalizationAlgo = o.NormAlgorithm
	opts.Hasher = o.Hasher
	if o.Issuer != "" {
		opts.Issuer = o.Issuer
	}
	if o.RootCerts != nil {
		opts.RootCerts = o.RootCerts
	}
	if len(o.SignatureNames) > 0 {
		opts.VerifySignature = o.Keys.GetPublicKey(o.SignatureNames[0]) != nil
	}
	opts.Update = o.Update
}
