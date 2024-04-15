// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package keyoption

import (
	"crypto/x509"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	ocmsign "github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/signutils"
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
	issuers     []string
	rootCAs     []string
	RootCerts   signutils.GenericCertificatePool
	Keys        signing.KeyRegistry
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.StringArrayVarP(&o.publicKeys, "public-key", "k", nil, "public key setting")
	fs.StringArrayVarP(&o.privateKeys, "private-key", "K", nil, "private key setting")
	fs.StringArrayVarP(&o.issuers, "issuer", "I", nil, "issuer name or distinguished name (DN) (optionally for dedicated signature) ([<name>:=]<dn>")
	fs.StringArrayVarP(&o.rootCAs, "ca-cert", "", nil, "additional root certificate authorities (for signing certificates)")
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
	for _, i := range o.issuers {
		name := o.DefaultName
		is := i
		sep := strings.Index(i, ":=")
		if sep >= 0 {
			name = i[:sep]
			is = i[sep+1:]
		}
		old := o.Keys.GetIssuer(name)
		dn, err := signutils.ParseDN(is)
		if err != nil {
			return errors.Wrapf(err, "issuer %q", i)
		}
		if old != nil && !reflect.DeepEqual(old, dn) {
			return fmt.Errorf("issuer already set (%s)", i)
		}

		o.Keys.RegisterIssuer(name, dn)
	}

	if len(o.rootCAs) > 0 {
		var list []*x509.Certificate
		for _, r := range o.rootCAs {
			data, err := utils.ReadFile(r, ctx.FileSystem())
			if err != nil {
				return errors.Wrapf(err, "root CA")
			}
			certs, err := signutils.GetCertificateChain(data, false)
			if err != nil {
				return errors.Wrapf(err, "root CA")
			}
			list = append(list, certs...)
		}
		o.RootCerts = list
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
			return errors.Newf("%s: empty file name", desc)
		}
		var data []byte
		var err error
		switch file[0] {
		case '=', '!', '@':
			data, err = utils.ResolveData(file, ctx.FileSystem())
		default:
			data, err = utils.ReadFile(file, ctx.FileSystem())
		}
		if err != nil {
			return errors.Wrapf(err, "cannot read %s file %q", desc, file)
		}
		if name == "" {
			return errors.Newf("%s: key name required", desc)
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

With <code>--issuer</code> it is possible to declare expected issuer 
constraints for public key certificates provided as part of a signature
required to accept the provisioned public key (besides the successful
validation of the certificate). By default, the issuer constraint is
derived from the signature name. If it is not a formal distinguished name,
it is assumed to be a plain common name.

With <code>--ca-cert</code> it is possible to define additional root
certificates for signature verification, if public keys are provided
by a certificate delivered with the signature.
`
	return s
}

var _ ocmsign.Option = (*Option)(nil)

func (o *Option) ApplySigningOption(opts *ocmsign.Options) {
	opts.Keys = o.Keys
	opts.RootCerts = o.RootCerts
}
