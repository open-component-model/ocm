// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package rsakeypair

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"strings"
	"time"

	parse "github.com/mandelsoft/spiff/dynaml/x509"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/out"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

var (
	Names = names.RSAKeyPair
	Verb  = verbs.Create
)

type Command struct {
	utils.BaseCommand

	Subject     *pkix.Name
	MoreIssuers []string
	priv        string
	pub         string

	attrs  map[string]string
	cacert string
	cakey  string

	Validity time.Duration
	CACert   *x509.Certificate
	CAKey    interface{}
}

var _ utils.OCMCommand = (*Command)(nil)

// NewCommand creates a new artefact command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<private key file> [<public key file>]] {<subject-attribute>=<value>}",
		Short: "create RSA public key pair",
		Long: `
Create an RSA public key pair and save to files.

The default for the filename to store the private key is <code>rsa.priv</code>.
If no public key file is specified, its name will be derived from the filename for
the private key (suffix <code>.pub</code> for public key or <code>.cert</code> for certificate).
If a certificate authority is given (<code>--cacert</code>) the public key
will be signed. In this case a subject (at least common name/issuer) and a private
key (<code>--cakey</code>) is required. If only a subject is given, the public key will be self-signed.

For signing the public key the following subject attributes are supported:
- <code>CN</code>, <code>common-name</code>, <code>issuer</code>: Common Name/Issuer
- <code>O</code>, <code>organization</code>, <code>org</code>: Organization
- <code>OU</code>, <code>organizational-unit</code>, <code>org-unit</code>: Organizational Unit
- <code>STREET</code> (multiple): Street Address
- <code>POSTALCODE</code>, <code>postal-code</code> (multiple): Postal Code
- <code>L</code>, <code>locality</code> (multiple): Locality
- <code>S</code>, <code>province</code>, (multiple): Province
- <code>C</code>, <code>country</code>, (multiple): Country

	`,
		Example: `
$ ocm create rsakeypair mandelsoft.priv mandelsoft.cert issuer=mandelsoft
`,
	}
}

func (o *Command) AddFlags(set *pflag.FlagSet) {
	set.StringVarP(&o.cacert, "cacert", "", "", "certificate authority to sign public key")
	set.StringVarP(&o.cakey, "cakey", "", "", "private key for certificate authority")
	set.DurationVarP(&o.Validity, "validity", "", 10*24*365*time.Hour, "certificate validity")
}

func (o *Command) FilterSettings(args ...string) []string {
	o.attrs, args = common.FilterSettings(args...)
	return args
}

func (o *Command) Complete(args []string) error {
	args = o.FilterSettings(args...)
	if len(args) > 2 {
		return errors.Newf("only a maximum of two filenames possible")
	}
	if o.attrs != nil && len(o.attrs) > 0 {
		var subject pkix.Name
		for k, v := range o.attrs {
			switch strings.ToLower(k) {
			case "issuer", "common-name", "cn":
				if subject.CommonName == "" {
					subject.CommonName = v
				} else {
					o.MoreIssuers = append(o.MoreIssuers, v)
				}
			case "street":
				subject.StreetAddress = append(subject.StreetAddress, v)
			case "province", "st":
				subject.Province = append(subject.Province, v)
			case "country", "c":
				subject.Country = append(subject.Country, v)
			case "origanization", "org", "o":
				subject.Country = append(subject.Country, v)
			case "origanizational-unit", "org-unit", "ou":
				subject.OrganizationalUnit = append(subject.OrganizationalUnit, v)
			case "postal-code", "postalcode":
				subject.PostalCode = append(subject.PostalCode, v)
			case "locality", "l":
				subject.Locality = append(subject.Locality, v)
			default:
				return errors.ErrUnknown("subject attribute,", k)
			}
		}
		o.Subject = &subject
	}

	if o.cacert != "" {
		cert, err := parse.ParseCertificate(o.cacert)
		if err != nil {
			data, err := vfs.ReadFile(o.Context.FileSystem(), o.cacert)
			if err != nil {
				return errors.Wrapf(err, "cannot read ca cert file %q", o.cacert)
			}
			cert, err = parse.ParseCertificate(string(data))
			if err != nil {
				return errors.Wrapf(err, "no cert in file %q", o.cacert)
			}
		}
		o.CACert = cert
	}
	if o.cakey != "" {
		key, err := parse.ParsePrivateKey(o.cakey)
		if err != nil {
			data, err := vfs.ReadFile(o.Context.FileSystem(), o.cacert)
			if err != nil {
				return errors.Wrapf(err, "cannot read provate key file %q", o.cacert)
			}
			key, err = parse.ParsePrivateKey(string(data))
			if err != nil {
				return errors.Wrapf(err, "unknown private key in file %q", o.cacert)
			}
		}
		o.CAKey = key
	}
	if o.CACert != nil && o.CAKey == nil {
		return errors.Newf("private key required for signing public key")
	}
	if o.CACert == nil && o.CAKey != nil {
		return errors.Newf("ca certificate required for signing public key")
	}

	if o.Subject != nil {
		if o.Subject.CommonName == "" {
			return errors.Newf("at least the common-name for a subject must be given")
		}
	}
	if len(args) > 0 {
		o.priv = args[0]
	} else {
		o.priv = "rsa.priv"
	}
	if len(args) > 1 {
		o.pub = args[1]
	} else {
		suf := "pub"
		if o.Subject != nil {
			suf = "cert"
		}
		if strings.HasSuffix(o.priv, ".priv") {
			o.pub = o.priv[:len(o.priv)-4] + suf
		} else {
			o.pub = o.priv + "." + suf
		}
	}

	return nil
}

func (o *Command) Run() error {
	priv, pub, err := rsa.Handler{}.CreateKeyPair()
	if err != nil {
		return err
	}

	if o.Subject != nil {
		key := o.CAKey
		if o.CAKey == nil {
			key = priv
		}
		pub, err = signing.CreateCertificate(*o.Subject, nil, o.Validity, pub, o.CACert, key, false, o.MoreIssuers...)
		if err != nil {
			return errors.Wrapf(err, "signing failed")
		}
	}
	if err := o.WriteKey(priv, o.priv); err != nil {
		return errors.Wrapf(err, "failed to write private key file %q", o.priv)
	}
	if err := o.WriteKey(pub, o.pub); err != nil {
		return errors.Wrapf(err, "failed to write public key file %q", o.pub)
	}
	out.Outf(o.Context, "created rsa key pair %s[%s]\n", o.priv, o.pub)
	return nil
}

func (o *Command) WriteKey(key interface{}, path string) error {
	fd, err := o.Context.FileSystem().OpenFile(path, vfs.O_CREATE|vfs.O_WRONLY, 0o600)
	if err != nil {
		return err
	}

	if certdata, ok := key.([]byte); ok {
		block := &pem.Block{Type: "CERTIFICATE", Bytes: certdata}
		err = pem.Encode(fd, block)
	} else {
		err = rsa.WriteKeyData(key, fd)
	}
	if err != nil {
		fd.Close()
		o.Context.FileSystem().Remove(path)
		return err
	}
	err = fd.Close()
	if err != nil {
		return err
	}
	return o.FileSystem().Chmod(path, 0o400)
}
